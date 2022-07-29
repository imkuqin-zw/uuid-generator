// +build wireinject

package alloc

import (
	"github.com/google/wire"
	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr"
)

var providerSet = wire.NewSet(
	seqsvr.NewTicker,
	seqsvr.NewRouter,
	seqsvr.NewEtcdv3Storage,
)

func newRouter() *seqsvr.Router {
	panic(wire.Build(providerSet))
}

func newStorage() seqsvr.Storage {
	panic(wire.Build(providerSet))
}
