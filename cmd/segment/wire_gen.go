// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
