package service

import (
	"context"

	"github.com/imkuqin-zw/uuid-generator/internal/segment/data"
	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
)

type SegmentService struct {
	repo data.SegmentRepo
	api.UnimplementedSegmentServer
}

func (s *SegmentService) FetchNext(ctx context.Context, req *api.FetchSegmentNextReq) (*api.UUID, error) {
	segment, err := s.repo.GetSegmentByID(ctx, req.Tag)
	if err != nil {
		return nil, err
	}
	seq, err := segment.FetchNext(ctx)
	if err != nil {
		return nil, err
	}
	return &api.UUID{Val: seq}, nil
}

func NewSegmentService(repo data.SegmentRepo) api.SegmentServer {
	return &SegmentService{repo: repo}
}
