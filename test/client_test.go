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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/proto"
	"github.com/imkuqin-zw/uuid-generator/pkg/etcdv3"
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
	clientv3 "go.etcd.io/etcd/client/v3"
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

func TestEtcd(t *testing.T) {
	if err := config.LoadSource(file.NewSource("./config.yaml", false)); err != nil {
		log.Fatal(err)
	}
	client, err := etcdv3.StdConfig().Build()
	if err != nil {
		t.Fatal(err)
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	routerKey := "/router"
	// routerVersionKey := fmt.Sprintf("%s/version", routerKey)
	allocKey := fmt.Sprintf("%s/sets", routerKey)

	nodes := make(map[string]*proto.AllocNode)
	ip := 52
	for i := 1; i < 20000000; i += 10000000 {
		setKey := fmt.Sprintf("%s/%d_%d_%d", allocKey, i, 10000000, 100000)
		for j := i; j < i+10000000; j += 5000000 {
			nodeKey := fmt.Sprintf("%s/192.168.3.%d", setKey, ip)
			node := &proto.AllocNode{
				Endpoints: []string{fmt.Sprintf("grpc://192.168.3.%d:32323", ip)},
			}
			ip++
			for k := j; k < j+5000000; k += 100000 {
				node.Sections = append(node.Sections, uint32(k))
			}
			nodes[nodeKey] = node
		}
	}

	for key, node := range nodes {
		val, _ := json.Marshal(node)
		client.Do(ctx, clientv3.OpPut(key, string(val)))
	}

	// txn := client.Txn(ctx).
	// 	If(clientv3.Compare(clientv3.Value(routerVersionKey), ">", fmt.Sprintf("%d", 0))).
	// 	Then(clientv3.OpGet(routerKey, clientv3.WithPrefix()))
	// txnRes, err := txn.Commit()
	// if err != nil {
	// 	t.Fatal(err)
	// 	return
	// }
	// for _, item := range txnRes.Responses {
	// 	rangeRes := item.GetResponseRange()
	// 	for _, kv := range rangeRes.Kvs {
	// 		fmt.Println(string(kv.Key))
	// 		fmt.Println(string(kv.Value))
	// 	}
	// }
}
