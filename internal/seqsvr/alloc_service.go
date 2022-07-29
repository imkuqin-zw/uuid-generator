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
