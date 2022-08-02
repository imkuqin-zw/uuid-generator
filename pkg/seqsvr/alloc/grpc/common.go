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
