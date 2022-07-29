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
