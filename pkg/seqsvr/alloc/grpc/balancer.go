package grpc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/consts"
	proto2 "github.com/imkuqin-zw/uuid-generator/internal/seqsvr/proto"
	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/util"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"
	"go.uber.org/atomic"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
	"google.golang.org/protobuf/proto"
)

type balancerBuilder struct {
}

// Build 创建一个Balancer
func (bb *balancerBuilder) Build(cc balancer.ClientConn, opts balancer.BuildOptions) balancer.Balancer {
	return &seqsvrNamingBalancer{
		cc:       cc,
		target:   opts.Target,
		subConns: make(map[resolver.Address]balancer.SubConn),
		scStates: make(map[balancer.SubConn]connectivity.State),
		csEvltr:  &balancer.ConnectivityStateEvaluator{},
	}
}

// Name return name
func (bb *balancerBuilder) Name() string {
	return scheme
}

type seqsvrNamingBalancer struct {
	cc      balancer.ClientConn
	target  resolver.Target
	rwMutex sync.RWMutex

	csEvltr *balancer.ConnectivityStateEvaluator
	state   connectivity.State

	subConns map[resolver.Address]balancer.SubConn
	scStates map[balancer.SubConn]connectivity.State

	router Router

	v2Picker balancer.Picker

	resolverErr error // the last error reported by the resolver; cleared on successful resolution
	connErr     error // the last connection error; cleared upon leaving TransientFailure
}

// Close closes the balancer. The balancer is not required to call
// ClientConn.RemoveSubConn for its existing SubConns.
func (p *seqsvrNamingBalancer) Close() {

}

func (p *seqsvrNamingBalancer) createSubConnection(addr resolver.Address) {
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()
	if _, ok := p.subConns[addr]; !ok {
		// a is a new address (not existing in b.subConns).
		sc, err := p.cc.NewSubConn([]resolver.Address{addr}, balancer.NewSubConnOptions{HealthCheckEnabled: true})
		if err != nil {
			log.Warnf("balancer: failed to create new SubConn: %v", err)
			return
		}
		p.subConns[addr] = sc
		p.scStates[sc] = connectivity.Idle
		sc.Connect()
	}
}

// UpdateClientConnState is called by gRPC when the state of the ClientConn
// changes.  If the error returned is ErrBadResolverState, the ClientConn
// will begin calling ResolveNow on the active name resolver with
// exponential backoff until a subsequent call to UpdateClientConnState
// returns a nil error.  Any other errors are currently ignored.
func (p *seqsvrNamingBalancer) UpdateClientConnState(state balancer.ClientConnState) error {
	if log.Enable(types.LvInfo) {
		grpclog.Infof("balancer: got new ClientConn state: %+v", state)
	}
	if len(state.ResolverState.Addresses) == 0 {
		p.ResolverError(errors.New("produced zero addresses"))
		return balancer.ErrBadResolverState
	}

	// Successful resolution; clear resolver error and ensure we return nil.
	p.resolverErr = nil
	// addrsSet is the set converted from addrs, it's used for quick lookup of an address.
	addrsSet := make(map[resolver.Address]struct{})
	router := Router{secTable: map[uint32]*string{}}
	if attr := state.ResolverState.Attributes; attr != nil {
		router.sets, _ = attr.Value(setTag{}).([][]uint32)
		router.version, _ = attr.Value(rvTag{}).(uint64)
		if router.version < p.router.version {
			return nil
		}
	}
	for _, a := range state.ResolverState.Addresses {
		router.all = append(router.all, a.Addr)
		addr := a.Addr
		if attr := a.BalancerAttributes; attr != nil {
			secTable, _ := attr.Value(secTag{}).([]uint32)
			for _, secID := range secTable {
				router.secTable[secID] = &(addr)
			}
		}
		addrsSet[a] = struct{}{}
		p.createSubConnection(a)
	}
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()
	for a, sc := range p.subConns {
		// a was removed by resolver.
		if _, ok := addrsSet[a]; !ok {
			p.cc.RemoveSubConn(sc)
			delete(p.subConns, a)
			// Keep the state of this sc in b.scStates until sc's state becomes Shutdown.
			// The entry will be deleted in HandleSubConnStateChange.
		}
	}
	p.router = router
	routerVersion.Store(router.version)
	return nil
}

// ResolverError is called by gRPC when the name resolver reports an error.
func (p *seqsvrNamingBalancer) ResolverError(err error) {
	p.resolverErr = err
	if len(p.subConns) == 0 {
		p.state = connectivity.TransientFailure
	}
	if p.state != connectivity.TransientFailure {
		// The picker will not change since the balancer does not currently
		// report an error.
		return
	}
	p.regeneratePicker()
	p.cc.UpdateState(balancer.State{
		ConnectivityState: p.state,
		Picker:            p.v2Picker,
	})
}

// UpdateSubConnState is called by gRPC when the state of a SubConn
// changes.
func (p *seqsvrNamingBalancer) UpdateSubConnState(sc balancer.SubConn, state balancer.SubConnState) {
	s := state.ConnectivityState
	if log.Enable(types.LvInfo) {
		grpclog.Infof("base.baseBalancer: handle SubConn state change: %p, %v", sc, s)
	}
	oldS, quit := func() (connectivity.State, bool) {
		p.rwMutex.Lock()
		defer p.rwMutex.Unlock()
		oldS, ok := p.scStates[sc]
		if !ok {
			if log.Enable(types.LvInfo) {
				grpclog.Infof("base.baseBalancer: got state changes for an unknown SubConn: %p, %v", sc, s)
			}
			return connectivity.TransientFailure, true
		}
		if oldS == connectivity.TransientFailure && s == connectivity.Connecting {
			// Once a subconn enters TRANSIENT_FAILURE, ignore subsequent
			// CONNECTING transitions to prevent the aggregated state from being
			// always CONNECTING when many backends exist but are all down.
			return oldS, true
		}
		p.scStates[sc] = s
		switch s {
		case connectivity.Idle:
			sc.Connect()
		case connectivity.Shutdown:
			// When an address was removed by resolver, b called RemoveSubConn but
			// kept the sc's state in scStates. Remove state for this sc here.
			delete(p.scStates, sc)
		case connectivity.TransientFailure:
			// Save error to be reported via picker.
			p.connErr = state.ConnectionError
		}
		return oldS, false
	}()
	if quit {
		return
	}
	p.state = p.csEvltr.RecordTransition(oldS, s)

	// Regenerate picker when one of the following happens:
	//  - this sc entered or left ready
	//  - the aggregated state of balancer is TransientFailure
	//    (may need to update error message)
	if (s == connectivity.Ready) != (oldS == connectivity.Ready) ||
		p.state == connectivity.TransientFailure {
		p.regeneratePicker()
	}

	p.cc.UpdateState(balancer.State{ConnectivityState: p.state, Picker: p.v2Picker})
}

// regeneratePicker takes a snapshot of the balancer, and generates a picker
// from it. The picker is
//  - errPicker if the balancer is in TransientFailure,
//  - built by the pickerBuilder with all READY SubConns otherwise.
func (p *seqsvrNamingBalancer) regeneratePicker() {
	if p.state == connectivity.TransientFailure {
		p.v2Picker = base.NewErrPicker(p.mergeErrors())
		return
	}
	readySCs := make(map[string]balancer.SubConn)
	p.rwMutex.RLock()
	defer p.rwMutex.RUnlock()
	// Filter out all ready SCs from full subConn map.
	for addr, sc := range p.subConns {
		if st, ok := p.scStates[sc]; ok && st == connectivity.Ready {
			readySCs[addr.Addr] = sc
		}
	}
	p.v2Picker = &seqsvrNamingPicker{
		router:   &p.router,
		readySCs: readySCs,
	}
}

// mergeErrors builds an error from the last connection error and the last
// resolver error.  Must only be called if b.state is TransientFailure.
func (p *seqsvrNamingBalancer) mergeErrors() error {
	// connErr must always be non-nil unless there are no SubConns, in which
	// case resolverErr must be non-nil.
	if p.connErr == nil {
		return fmt.Errorf("last resolver error: %v", p.resolverErr)
	}
	if p.resolverErr == nil {
		return fmt.Errorf("last connection error: %v", p.connErr)
	}
	return fmt.Errorf("last connection error: %v; last resolver error: %v", p.connErr, p.resolverErr)
}

type Router struct {
	version  uint64
	all      []string
	sets     [][]uint32
	secTable map[uint32]*string
}

func (r *Router) getAddr(info balancer.PickInfo) (string, error) {
	if len(r.sets) == 0 {
		if len(r.all) == 0 {
			return "", balancer.ErrNoSubConnAvailable
		}
		addr := r.all[rand.Intn(len(r.all))]
		return addr, nil
	}
	ID := info.Ctx.Value(pickTag{}).(uint32)
	_, set := util.SearchSetByID(r.sets, ID)
	if len(set) == 0 {
		return "", balancer.ErrNoSubConnAvailable
	}
	secID := util.CalcSectionIDByID(ID, set[2])
	addr, ok := r.secTable[secID]
	if !ok {
		return "", balancer.ErrNoSubConnAvailable
	}
	return *addr, nil
}

type seqsvrNamingPicker struct {
	routerChanged atomic.Bool
	router        *Router
	readySCs      map[string]balancer.SubConn
}

// Pick returns the connection to use for this RPC and related information.
//
// Pick should not block.  If the balancer needs to do I/O or any blocking
// or time-consuming work to service this call, it should return
// ErrNoSubConnAvailable, and the Pick call will be repeated by gRPC when
// the Picker is updated (using ClientConn.UpdateState).
//
// If an error is returned:
//
// - If the error is ErrNoSubConnAvailable, gRPC will block until a new
//   Picker is provided by the balancer (using ClientConn.UpdateState).
//
// - If the error implements IsTransientFailure() bool, returning true,
//   wait for ready RPCs will wait, but non-wait for ready RPCs will be
//   terminated with this error's Error() string and status code
//   Unavailable.
//
// - Any other errors terminate all RPCs with the code and message
//   provided.  If the error is not a status error, it will be converted by
//   gRPC to a status error with code Unknown.
func (pnp *seqsvrNamingPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	addr, err := pnp.router.getAddr(info)
	if err != nil {
		return balancer.PickResult{}, err
	}
	subSc, ok := pnp.readySCs[addr]
	if ok {
		reporter := &resultReporter{picker: pnp}
		return balancer.PickResult{
			SubConn: subSc,
			Done:    reporter.report,
		}, nil
	}
	return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
}

type resultReporter struct {
	picker *seqsvrNamingPicker
}

func (r *resultReporter) report(info balancer.DoneInfo) {
	if info.Trailer == nil {
		return
	}
	routerData := info.Trailer[consts.MDTagRouter]
	if len(routerData) == 0 {
		return
	}
	if r.picker.routerChanged.CAS(false, true) {
		xgo.Go(func() {
			if !r.syncRouter(routerData[0]) {
				r.picker.routerChanged.Store(false)
			}
		}, func(err interface{}) {
			r.picker.routerChanged.Store(false)
		})
	}
}

func (r *resultReporter) syncRouter(routerData string) bool {
	router := &proto2.Router{}
	protoData, _ := hex.DecodeString(routerData)
	if err := proto.Unmarshal(protoData, router); err != nil {
		log.Warnf("fault to unmarshal router, err: %s", err.Error())
		return false
	}
	eventChan <- router
	return true
}
