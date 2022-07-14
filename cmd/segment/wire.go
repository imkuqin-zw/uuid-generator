// +build wireinject

package segment

import (
	"github.com/google/wire"
	"github.com/imkuqin-zw/uuid-generator/internal/segment/data"
	"github.com/imkuqin-zw/uuid-generator/internal/segment/service"
	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	grpcimpl "github.com/imkuqin-zw/uuid-generator/pkg/genproto/api/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/server/grpc"
)

var providerSet = wire.NewSet(
	data.NewSegmentRepo,
	data.NewDB,
	service.NewSegmentService,
)

func newApiServer() api.SegmentServer {
	panic(wire.Build(providerSet))
}

func RegisterService() {
	segmentSvr := newApiServer()
	grpc.RegisterService(&grpcimpl.SegmentServiceDesc, segmentSvr)
}
