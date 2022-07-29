package seqsvr

import (
	"sync/atomic"

	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/proto"
)

type Router struct {
	cache atomic.Value
}

func NewRouter() *Router {
	r := &Router{}
	r.cache.Store(&proto.Router{})
	return r
}

func (r *Router) Version() uint64 {
	return r.cache.Load().(*proto.Router).Version
}

func (r *Router) GetRouter() *proto.Router {
	return r.cache.Load().(*proto.Router)
}

func (r *Router) UpdateRouter(router *proto.Router) {
	r.cache.Store(router)
}
