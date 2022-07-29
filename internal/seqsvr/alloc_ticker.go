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
	"net/url"
	"os"
	"time"

	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/proto"
	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/util"
	"github.com/imkuqin-zw/uuid-generator/pkg/filelock"
	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
	"github.com/pkg/errors"
)

type TickType int

const (
	TickBase      TickType = 0
	TickHeartbeat TickType = 1
)

type tickSchedule struct {
	runAt    int64
	interval int64
}

type Ticker struct {
	setID    uint32
	nodeInfo RegisterNodeInfo

	baseTickInterval time.Duration
	tick             int64
	schedules        []tickSchedule

	heartbeatElapsed int64
	heartbeatTimeout int64

	router   *Router
	sequence *Sequence
	storage  Storage

	stop chan struct{}
}

func NewTicker(router *Router, sequence *Sequence, storeClient Storage) *Ticker {
	heartbeatTicks := config.GetInt64("seqsvr.HeartbeatTicks", 2)
	heartbeatTimeout := config.GetInt64("seqsvr.heartbeatTimeoutTicks", 5)
	if heartbeatTimeout <= heartbeatTicks {
		log.Fatalf("heartbeatTimeoutTicks must greater than heartbeatTicks")
		return nil
	}
	IP, err := xnet.Extract("")
	if err != nil {
		log.Fatalf("fault to extract private IP, err: %+v", err)
		return nil
	}
	ticker := &Ticker{
		setID:            uint32(config.GetInt("seqsvr.setID", 1)),
		baseTickInterval: config.GetDuration("seqsvr.baseTickInterval", time.Millisecond*500),
		schedules:        make([]tickSchedule, 2),
		heartbeatTimeout: heartbeatTimeout,
		router:           router,
		sequence:         sequence,
		storage:          storeClient,
		stop:             make(chan struct{}),
	}
	ticker.nodeInfo.IP = IP
	ticker.schedules[TickBase].interval = 1
	ticker.schedules[TickHeartbeat].interval = heartbeatTicks
	return ticker
}

func (t *Ticker) Run() error {
	if err := Lock(); err != nil {
		if errors.Is(err, filelock.ErrLocked) {
			log.Warn("worker already exists")
			os.Exit(1)
		}
		return err
	}
	t.nodeInfo.Endpoints = make([]*RegisterEndpoint, 0, len(yggdrasil.Endpoints()))
	node := &proto.AllocNode{IP: t.nodeInfo.IP}
	for _, endpoint := range yggdrasil.Endpoints() {
		urlVal := url.Values{}
		urlVal.Set("kind", string(endpoint.Kind()))
		for k, v := range endpoint.Metadata() {
			urlVal.Set(k, v)
		}
		node.Endpoints = append(node.Endpoints, endpoint.Endpoint()+"?"+urlVal.Encode())
		t.nodeInfo.Endpoints = append(t.nodeInfo.Endpoints, &RegisterEndpoint{
			Scheme:   endpoint.Scheme(),
			Kind:     endpoint.Kind(),
			Host:     endpoint.Host(),
			Metadata: endpoint.Metadata(),
		})
	}
	if err := t.storage.Register(context.TODO(), &t.nodeInfo); err != nil {
		return err
	}
	defer func() {
		if err := t.storage.Unregister(context.TODO()); err != nil {
			log.Errorf("fault to unregister node, err: %+v", err)
		}
	}()
	for i := range t.schedules {
		t.schedule(TickType(i))
	}
	timer := time.Tick(t.baseTickInterval)
	for {
		select {
		case <-timer:
			t.onTick()
		case <-t.stop:
			return nil
		}
	}
}

func (t *Ticker) Stop() {
	close(t.stop)
}

func (t *Ticker) heartbeat() {
	// 发送心跳
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1)
	defer cancel()
	paused := t.sequence.Paused()
	if paused {
		if err := t.storage.Register(ctx, &t.nodeInfo); err != nil {
			log.Errorf("fault to register again, err: %+v", err)
			return
		}
	}
	router, err := t.storage.Heartbeat(ctx, t.router.Version())
	if err != nil {
		log.Errorf("fault to send heartbeat, err: %+v", err)
		return
	}

	if router != nil {
		if paused {
			t.sequence.Reset()
		}
		log.Info("router change")
		t.router.UpdateRouter(router)
		applyTicks := t.tick + t.heartbeatTimeout
		set := util.SearchSet(router.Sets, t.setID)
		var sections []uint32
		var sectionSize uint32
		if set != nil {
			node := util.SearchNode(set.Nodes, t.nodeInfo.IP)
			if node != nil {
				sections = node.Sections
			}
			sectionSize = set.SectionSize
		}
		t.sequence.CommitRouter(router.Version, applyTicks, sectionSize, sections)
	}
	if paused {
		t.sequence.Open()
	}
	t.heartbeatElapsed = 0
}

func (t *Ticker) baseTick() {
	// 检查心跳超时
	if t.heartbeatElapsed >= t.heartbeatTimeout {
		// 停止服务
		t.sequence.Pause()
	} else {
		t.heartbeatElapsed++
		t.sequence.ApplyRouter(t.tick)
	}
}

func (t *Ticker) tickClock() {
	t.tick++
}

func (t *Ticker) isOnTick(tp TickType) bool {
	sched := &t.schedules[int(tp)]
	return sched.runAt == t.tick
}

func (t *Ticker) schedule(tp TickType) {
	sched := &t.schedules[int(tp)]
	if sched.interval <= 0 {
		sched.runAt = -1
		return
	}
	sched.runAt = t.tick + sched.interval
}

func (t *Ticker) onBaseTick() {
	t.baseTick()
	t.schedule(TickBase)
}

func (t *Ticker) onHeartbeatTick() {
	t.heartbeat()
	t.schedule(TickHeartbeat)
}

func (t *Ticker) onTick() {
	t.tickClock()
	if t.isOnTick(TickBase) {
		t.onBaseTick()
	}
	if t.isOnTick(TickHeartbeat) {
		t.onHeartbeatTick()
	}
}
