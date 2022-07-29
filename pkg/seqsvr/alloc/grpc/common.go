package grpc

import (
	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/proto"
	"go.uber.org/atomic"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

const scheme = "seqsvr"

type (
	pickTag struct{}
	setTag  struct{}
	rvTag   struct{}
	secTag  struct{}
	mdTag   struct{}
)

var (
	eventChan     = make(chan *proto.Router, 1)
	routerVersion atomic.Uint64
)

func init() {
	resolver.Register(&resolverBuilder{})
	balancer.Register(&balancerBuilder{})
}
