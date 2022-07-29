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
	"sync"
	"sync/atomic"

	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/util"
)

const (
	defaultStep = 5000
)

const (
	SectionStateCreated = iota
	SectionStateInitialized
)

type Section struct {
	ID     uint32
	size   uint32
	maxID  uint64
	seq    []uint64
	step   uint64
	store  Storage
	state  uint32
	loadMu sync.Mutex
	next   func(ctx context.Context, ID uint32) (seq uint64, err error)
}

func NewSection(ID, size uint32, storeCli Storage) *Section {
	s := &Section{
		ID:    ID,
		size:  size,
		store: storeCli,
		step:  defaultStep,
	}
	s.next = s.initAndFetchNext
	return s
}

func (s *Section) init(ctx context.Context, seq uint64) error {
	s.loadMu.Lock()
	defer s.loadMu.Unlock()
	if s.maxID > seq {
		return nil
	}
	maxID, err := s.store.IncrAndGetMax(ctx, s.ID, s.maxID, 100)
	if err != nil {
		return err
	}
	s.maxID = maxID
	s.seq = util.MakeSeq(maxID, s.size)
	s.next = s.fetchNext
	return nil
}

func (s *Section) load(ctx context.Context, seq uint64) error {
	s.loadMu.Lock()
	defer s.loadMu.Unlock()
	if s.maxID > seq {
		return nil
	}
	maxID, err := s.store.IncrAndGetMax(ctx, s.ID, s.maxID, 100)
	if err != nil {
		return err
	}
	s.maxID = maxID
	return nil
}

func (s *Section) FetchNextSeq(ctx context.Context, ID uint32) (seq uint64, err error) {
	return s.next(ctx, ID)
}

func (s *Section) initAndFetchNext(ctx context.Context, ID uint32) (seq uint64, err error) {
	if err := s.init(ctx, 0); err != nil {
		return 0, err
	}
	return s.fetchNext(ctx, ID)
}

func (s *Section) fetchNext(ctx context.Context, ID uint32) (seq uint64, err error) {
	seq = atomic.AddUint64(&s.seq[util.CalcSeqIdxByID(ID, s.size)], 1)
	if s.maxID > seq {
		return
	}
	if err = s.load(ctx, seq); err != nil {
		return 0, err
	}
	return seq, err
}
