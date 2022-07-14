// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package segment

import (
	"github.com/google/wire"
	"github.com/imkuqin-zw/uuid-generator/internal/segment/data"
	"github.com/imkuqin-zw/uuid-generator/internal/segment/service"
	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/server/grpc"
)

// Injectors from wire.go:

func newApiServer() api.SegmentServer {
	db := data.NewDB()
	segmentRepo := data.NewSegmentRepo(db)
	segmentServer := service.NewSegmentService(segmentRepo)
	return segmentServer
}

// wire.go:

var providerSet = wire.NewSet(data.NewSegmentRepo, data.NewDB, service.NewSegmentService)

func RegisterService() {
	segmentSvr := newApiServer()
	grpc.RegisterService(&grpcimpl.SegmentServiceDesc, segmentSvr)
}
