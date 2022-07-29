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

package seqsvr

import (
	"context"
	"fmt"
	"strconv"

	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/consts"
	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	"github.com/imkuqin-zw/yggdrasil/pkg/md"
	"google.golang.org/protobuf/proto"
)

type AllocService struct {
	sequence *Sequence
	router   *Router
	api.UnimplementedAllocServer
}

func NewAllocService(sequence *Sequence, router *Router) *AllocService {
	return &AllocService{sequence: sequence, router: router}
}

func (a *AllocService) FetchNext(ctx context.Context, req *api.FetchSeqNextReq) (*api.UUID, error) {
	if metadata, ok := md.FromInContext(ctx); ok {
		routerVersion := metadata[consts.MDTagRouterVersion]
		if len(routerVersion) > 0 {
			if version, err := strconv.ParseUint(routerVersion[0], 10, 64); err == nil {
				if version < a.router.Version() {
					data, _ := proto.Marshal(a.router.GetRouter())
					_ = md.SetTrailer(ctx, md.New(map[string]string{consts.MDTagRouter: fmt.Sprintf("%0x", data)}))
				}
			}
		}
	}
	val, err := a.sequence.FetchNextSeq(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	res := &api.UUID{Val: val}
	return res, nil
}
