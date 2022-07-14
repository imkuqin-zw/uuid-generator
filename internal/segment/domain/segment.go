package domain

import (
	"github.com/pkg/errors"
	"go.uber.org/atomic"
)
import "context"

type Segment struct {
	ID               string
	Step             uint32
	Seq              *atomic.Uint64
	Max              uint64
	PreloadThreshold float32
	repo             SegmentRepo
}

type SegmentRepo interface {
	FetchNextSegment(ctx context.Context, ID string, maxID uint64) (*Segment, error)
	SaveSegment(ctx context.Context, ID string, seq, max uint64, step uint32, preloadThreshold float32)
}

func (s *Segment) FetchNext(ctx context.Context) (uint64, error) {
	seq := s.Seq.Inc()
	if seq >= s.Max {
		s, err := s.repo.FetchNextSegment(ctx, s.ID, s.Max)
		if err != nil {
			return 0, err
		}
		select {
		case <-ctx.Done():
			return 0, errors.New("request timeout")
		default:
			return s.FetchNext(ctx)
		}
	}
	s.repo.SaveSegment(ctx, s.ID, seq, s.Max, s.Step, s.PreloadThreshold)
	return seq, nil
}

func NewSegment(do *Segment, repo SegmentRepo) *Segment {
	do.repo = repo
	return do
}
