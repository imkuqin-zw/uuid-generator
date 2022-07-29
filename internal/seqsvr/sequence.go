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
	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	"github.com/imkuqin-zw/yggdrasil/pkg/errors"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	errors2 "github.com/pkg/errors"
)

const (
	StatePause uint32 = iota + 1
	StateOpen
)

type PrepareApply struct {
	Version     uint64
	ApplyTick   int64
	Sections    []uint32
	SectionSize uint32
}

type Sequence struct {
	state        uint32
	sectionSize  uint32
	sections     sync.Map
	storeCli     Storage
	prepareApply *PrepareApply
}

func NewSequence(storeCli Storage) *Sequence {
	return &Sequence{
		state:    StateOpen,
		storeCli: storeCli,
	}
}

func (s *Sequence) Pause() {
	atomic.StoreUint32(&s.state, StatePause)
}

func (s *Sequence) Paused() bool {
	return atomic.LoadUint32(&s.state) == StatePause
}

func (s *Sequence) Reset() {
	s.sections.Range(func(key, value interface{}) bool {
		s.sections.Delete(key)
		return true
	})
}

func (s *Sequence) Open() {
	atomic.StoreUint32(&s.state, StateOpen)
}

func (s *Sequence) CommitRouter(version uint64, applyTick int64, sectionSize uint32, sections []uint32) {
	needDel := make([]uint32, 0)
	s.sections.Range(func(key, value interface{}) bool {
		id := key.(uint32)
		if !util.HasSection(sections, id) {
			needDel = append(needDel, id)
		}
		return true
	})
	for _, id := range needDel {
		s.sections.Delete(id)
	}
	prepareApply := &PrepareApply{
		Version:     version,
		ApplyTick:   applyTick,
		SectionSize: sectionSize,
	}
	for _, sectionID := range sections {
		if _, ok := s.sections.Load(sectionID); !ok {
			prepareApply.Sections = append(prepareApply.Sections, sectionID)
		}
	}
	if len(prepareApply.Sections) > 0 {
		s.prepareApply = prepareApply
	}
	if len(needDel) > 0 || len(prepareApply.Sections) > 0 {
		log.Info("section change")
	}
}

func (s *Sequence) ApplyRouter(tick int64) {
	if s.prepareApply == nil || s.prepareApply.ApplyTick > tick {
		return
	}
	s.sectionSize = s.prepareApply.SectionSize
	for _, ID := range s.prepareApply.Sections {
		s.sections.Store(ID, NewSection(ID, s.sectionSize, s.storeCli))
	}
	s.prepareApply = nil
	log.Info("apply section change")
}

func (s *Sequence) FetchNextSeq(ctx context.Context, ID uint32) (seq uint64, err error) {
	if atomic.LoadUint32(&s.state) == StatePause {
		return 0, errors.WithReason(errors2.New("service paused"), api.SeqsvrErrReason_SERVICE_PAUSE, nil)
	}
	sectionID := util.CalcSectionIDByID(ID, s.sectionSize)
	section, ok := s.sections.Load(sectionID)
	if !ok {
		return 0, errors.WithReason(errors2.New("sequence not found"), api.SeqsvrErrReason_NOT_FOUND_SEQUENCE, nil)
	}
	return section.(*Section).FetchNextSeq(ctx, ID)
}
