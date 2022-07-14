// +build wireinject

package snowflake

import (
	"github.com/google/wire"
	"github.com/imkuqin-zw/uuid-generator/internal/snowflake"
	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	grpcimpl "github.com/imkuqin-zw/uuid-generator/pkg/genproto/api/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/server/grpc"
)

var providerSet = wire.NewSet(
	snowflake.NewEtcdv3Coordinator,
	snowflake.NewWorker,
	snowflake.NewService,
)

func newApiServer() api.SnowflakeServer {
	panic(wire.Build(providerSet))
}

func RegisterService() {
	svr := newApiServer()
	grpc.RegisterService(&grpcimpl.SnowflakeServiceDesc, svr)
}
