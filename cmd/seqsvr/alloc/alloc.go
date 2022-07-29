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

package alloc

import (
	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr"
	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	grpcimpl "github.com/imkuqin-zw/uuid-generator/pkg/genproto/api/grpc"
	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/server/grpc"
	"github.com/urfave/cli/v2"
)

func NewCmd() *cli.Command {
	return &cli.Command{
		Name:  "alloc",
		Usage: "seqsvr alloc server",
		Action: func(c *cli.Context) error {
			appName := config.GetString("application.name", "com.github.imkuqin_zw.uuid-generator.seqsvr.allloc")
			service, ticker := NewServiceAndTicker()
			grpc.RegisterService(&grpcimpl.AllocServiceDesc, service)
			opt := yggdrasil.WithBeforeStartHook(func() error {
				go func() {
					if err := ticker.Run(); err != nil {
						log.Fatalf("fault to run ticker, err: %+v", err)
					}
				}()
				return nil
			})
			return yggdrasil.Run(appName, opt)
		},
	}
}

func NewServiceAndTicker() (api.AllocServer, *seqsvr.Ticker) {
	router := newRouter()
	storage := newStorage()
	sequence := seqsvr.NewSequence(storage)
	ticker := seqsvr.NewTicker(router, sequence, storage)
	service := seqsvr.NewAllocService(sequence, router)
	return service, ticker
}
