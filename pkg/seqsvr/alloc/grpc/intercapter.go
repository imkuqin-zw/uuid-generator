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
	"context"
	"fmt"
	"runtime"

	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/consts"
	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	"github.com/imkuqin-zw/yggdrasil/pkg/errors"
	"github.com/imkuqin-zw/yggdrasil/pkg/md"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func routerVersionInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) error {
	v := routerVersion.Load()
	ctx = md.WithTrailerOptCtx(ctx)
	if q, ok := req.(*api.FetchSeqNextReq); ok {
		ctx = context.WithValue(ctx, pickTag{}, q.ID)
	}
	if m, ok := metadata.FromOutgoingContext(ctx); ok {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Join(m, metadata.MD{
			consts.MDTagRouterVersion: []string{fmt.Sprintf("%d", v)},
		}))
	} else {
		ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
			consts.MDTagRouterVersion: []string{fmt.Sprintf("%d", v)},
		})
	}
	err := invoker(ctx, method, req, reply, cc, opts...)
	if err == nil {
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return err
		default:
		}
		if v == routerVersion.Load() {
			runtime.Gosched()
			continue
		}
		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			if errors.IsReason(err, api.SeqsvrErrReason_NOT_FOUND_SEQUENCE, api.SeqsvrErrReason_ERROR_REASON_UNSPECIFIED) {
				continue
			}
			return err
		}
		return err
	}
}
