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
