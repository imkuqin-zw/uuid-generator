package grpc

import (
	"context"
	"net/url"
	"strings"

	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/proto"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"google.golang.org/grpc/resolver"
)

type resolverBuilder struct {
}

func (rb *resolverBuilder) Scheme() string {
	return scheme
}

func (rb *resolverBuilder) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	opts resolver.BuildOptions,
) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(context.Background())
	d := &seqsvrResolver{
		cc:     cc,
		ctx:    ctx,
		cancel: cancel,
		target: target,
	}
	if err := d.start(target.URL.Host); err != nil {
		return nil, err
	}
	go d.watcher()
	return d, nil
}

type seqsvrResolver struct {
	ctx     context.Context
	cancel  context.CancelFunc
	cc      resolver.ClientConn
	target  resolver.Target
	version uint64
}

// ResolveNow The method is called by the gRPC framework to resolve the target name
func (sr *seqsvrResolver) ResolveNow(opt resolver.ResolveNowOptions) {}

func (sr *seqsvrResolver) start(endpoint string) error {
	addrList := strings.Split(endpoint, ",")
	state := resolver.State{}
	for _, addr := range addrList {
		state.Addresses = append(state.Addresses, resolver.Address{
			Addr: addr,
		})
	}
	if err := sr.cc.UpdateState(state); nil != err {
		log.Errorf("fail to do update service %s: %+v", sr.target.URL.Host, err)
		return err
	}
	return nil
}

func (sr *seqsvrResolver) watcher() {
	for {
		select {
		case <-sr.ctx.Done():
			return
		case router := <-eventChan:
			if sr.version >= router.Version {
				continue
			}
			state, err := sr.resolveRouter(router)
			if err != nil {
				sr.cc.ReportError(err)
			} else {
				err = sr.cc.UpdateState(*state)
				if nil != err {
					log.Errorf("fail to do update service %s: %+v", sr.target.URL.Host, err)
				}
			}
		}
	}
}

func (sr *seqsvrResolver) resolveRouter(router *proto.Router) (*resolver.State, error) {
	sr.version = router.Version
	state := &resolver.State{}
	sets := make([][]uint32, len(router.Sets))
	for i, set := range router.Sets {
		sets[i] = []uint32{set.ID, set.Size, set.SectionSize}
		for _, node := range set.Nodes {
			address := resolver.Address{}
			md := make(map[string]interface{})
			for k, v := range node.Metadata {
				md[k] = v
			}
			for _, item := range node.Endpoints {
				if strings.HasPrefix(item, "grpc") {
					uri, _ := url.Parse(item)
					address.Addr = uri.Host
					for k, v := range uri.Query() {
						md[k] = v
					}
					break
				}
			}
			if len(address.Addr) == 0 {
				continue
			}
			address.BalancerAttributes = address.BalancerAttributes.
				WithValue(mdTag{}, md).
				WithValue(secTag{}, node.Sections)
			state.Addresses = append(state.Addresses, address)
		}
	}
	state.Attributes = state.Attributes.WithValue(setTag{}, sets)
	state.Attributes = state.Attributes.WithValue(rvTag{}, router.Version)
	return state, nil
}

// Close resolver closed
func (sr *seqsvrResolver) Close() {
	sr.cancel()
}
