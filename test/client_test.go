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

package test

import (
	"context"
	"testing"
	"time"

	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	grpcimpl "github.com/imkuqin-zw/uuid-generator/pkg/genproto/api/grpc"
	"github.com/imkuqin-zw/yggdrasil"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/logger/zap"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/polaris"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/polaris/grpc"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/promethues"
	_ "github.com/imkuqin-zw/yggdrasil/contrib/trace/jaeger"
	"github.com/imkuqin-zw/yggdrasil/pkg/client/grpc"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/client/grpc/trace"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/server/governor"
	_ "github.com/imkuqin-zw/yggdrasil/pkg/server/grpc/trace"
)

func TestSegmentClient(t *testing.T) {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		log.Fatal(err)
	}
	go yggdrasil.Run("com.github.imkuqin_zw.uuid-generator.segment.client_test")
	client := grpcimpl.NewSegmentClient(grpc.Dial("com.github.imkuqin_zw.uuid-generator.segment"))
	f := func() {
		res, err := client.FetchNext(context.TODO(), &api.FetchSegmentNextReq{Tag: "test"})
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("call res: %d", res.Val)
		}
	}
	// f()
	ticker := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ticker.C:
			f()
			ticker.Reset(time.Second * 2)
		}
	}
}

func TestSnowflakeClient(t *testing.T) {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		log.Fatal(err)
	}
	go yggdrasil.Run("com.github.imkuqin_zw.uuid-generator.snowflake.client_test")
	client := grpcimpl.NewSnowflakeClient(grpc.Dial("com.github.imkuqin_zw.uuid-generator.snowflake"))
	f := func() {
		res, err := client.FetchNext(context.TODO(), &api.FetchSnowflakeNextReq{})
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("call res: %d", res.Val)
		}
	}
	// f()
	ticker := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-ticker.C:
			f()
			ticker.Reset(time.Second * 2)
		}
	}
}
