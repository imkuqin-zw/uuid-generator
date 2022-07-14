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

package snowflake

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/imkuqin-zw/uuid-generator/pkg/etcdv3"
	"github.com/imkuqin-zw/uuid-generator/pkg/etcdv3lock"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/client/v3"
)

type Worker struct {
	WorkerID *uint64 `json:"workerID,omitempty"`
	IP       string  `json:"ip,omitempty"`
	LastTime int64   `json:"lastTime"`
	Working  bool    `json:"working,omitempty"`
}

type Coordinator interface {
	GetOrCreateWorker(ctx context.Context) (*Worker, error)
	SyncTime(ctx context.Context) error
}

type Etcdv3Coordinator struct {
	dc           uint64
	ip           string
	workerID     uint64
	maxWorkerLen uint
	client       *clientv3.Client
	locker       *etcdv3lock.Mutex
}

func NewEtcdv3Coordinator() Coordinator {
	dc := uint64(config.GetInt64("snowflake.dc", 0))
	client, err := etcdv3.StdConfig("snowflake").Build()
	if err != nil {
		return nil
	}
	ip, err := xnet.Extract("")
	if err != nil {
		log.Fatalf("fault to get private ip, err: %+v", err)
		return nil
	}
	mutex, err := etcdv3lock.NewMutex(client, "/uuid-generator/lock/worker")
	if err != nil {
		log.Fatalf("fault to create etcd locker, err: %+v", err)
		return nil
	}
	return &Etcdv3Coordinator{dc: dc, ip: ip, maxWorkerLen: uint(maxWorkID), client: client, locker: mutex}
}

func (e *Etcdv3Coordinator) GetOrCreateWorker(ctx context.Context) (*Worker, error) {
	worker, err := e.loadWorkerByIp(ctx)
	if err != nil {
		return nil, err
	}
	if worker != nil {
		return worker, err
	}
	return e.genWorker(ctx)
}

func (e *Etcdv3Coordinator) SyncTime(ctx context.Context) error {
	ipNode := &Worker{
		WorkerID: &e.workerID,
		LastTime: e.getUnixMill(),
	}
	idNode := &Worker{
		LastTime: e.getUnixMill(),
		IP:       e.ip,
		Working:  true,
	}
	ipData, _ := json.Marshal(ipNode)
	idData, _ := json.Marshal(idNode)
	res, err := e.client.Txn(ctx).Then(
		clientv3.OpPut(e.workerIDKey(), xstrings.Bytes2str(idData)),
		clientv3.OpPut(e.workerIPKey(), xstrings.Bytes2str(ipData)),
	).Commit()
	if err != nil {
		return errors.WithStack(err)
	}
	if !res.Succeeded { // 事务成功
		return errors.New("fault to sync time")
	}
	return nil
}

func (e *Etcdv3Coordinator) workerIPKey() string {
	return fmt.Sprintf("/uuid-generator/dc/%d/IP/%s", e.dc, e.ip)
}

func (e *Etcdv3Coordinator) workerIDKey() string {
	return fmt.Sprintf("/uuid-generator/dc/%d/ID/%d", e.dc, e.workerID)
}

func (e *Etcdv3Coordinator) WorkerIDKeyPrefix() string {
	return fmt.Sprintf("/uuid-generator/dc/%d/ID", e.dc)
}

func (e *Etcdv3Coordinator) loadWorkerByIp(ctx context.Context) (*Worker, error) {
	res, err := e.client.Get(ctx, e.workerIPKey())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if res.Count != 0 {
		worker := &Worker{}
		if err := json.Unmarshal(res.Kvs[0].Value, worker); err != nil {
			return nil, errors.WithStack(err)
		}
		return worker, nil
	}
	return nil, nil
}

func (e *Etcdv3Coordinator) getWorkerInfoByWorkerID(ctx context.Context) (uint64, int64, error) {
	res, err := e.client.Get(ctx, fmt.Sprintf("/uuid-generator/worker/%d", e.workerID))
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}
	if res.Count != 0 {
		worker := &Worker{}
		if err := json.Unmarshal(res.Kvs[0].Value, worker); err != nil {
			return 0, 0, errors.WithStack(err)
		}
		e.workerID = *worker.WorkerID
		return *worker.WorkerID, worker.LastTime, nil
	}
	return e.workerID, 0, nil
}

func (e *Etcdv3Coordinator) genWorker(ctx context.Context) (*Worker, error) {
	if err := e.locker.LockWithCtx(ctx); err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		_ = e.locker.UnLockWithCtx(ctx)
	}()
	workerID, err := e.getDCWorkerID(ctx)
	if err != nil {
		return nil, err
	}
	e.workerID = workerID
	return e.saveWorker(ctx)
}

func (e *Etcdv3Coordinator) getDCWorkerID(ctx context.Context) (uint64, error) {
	res, err := e.client.Get(ctx, e.WorkerIDKeyPrefix(),
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
	)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	if len(res.Kvs) == 0 {
		return 0, nil
	}
	if uint(len(res.Kvs)) == e.maxWorkerLen {
		return 0, errors.New("the number of working nodes exceeds the limit")
	}
	now := e.getUnixMill()
	for i, kv := range res.Kvs {
		if !bytes.HasSuffix(kv.Key, xstrings.Str2bytes(fmt.Sprintf("%d", i))) {
			return uint64(i), nil
		}
		w := &Worker{}
		if err := json.Unmarshal(kv.Value, w); err != nil {
			log.Warnf("fault to unmarshal workerID value, err: %+v", err)
			continue
		}
		if !w.Working {
			if now < w.LastTime {
				return 0, errors.New("clock back")
			}
			return uint64(i), nil
		}
	}
	return uint64(len(res.Kvs)), nil
}

func (e *Etcdv3Coordinator) getUnixMill() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (e *Etcdv3Coordinator) saveWorker(ctx context.Context) (*Worker, error) {
	ipNode := &Worker{
		WorkerID: &e.workerID,
		LastTime: e.getUnixMill(),
	}
	idNode := &Worker{
		LastTime: e.getUnixMill(),
		IP:       e.ip,
		Working:  true,
	}
	ipData, _ := json.Marshal(ipNode)
	idData, _ := json.Marshal(idNode)
	res, err := e.client.Txn(ctx).Then(
		clientv3.OpPut(e.workerIDKey(), xstrings.Bytes2str(idData)),
		clientv3.OpPut(e.workerIPKey(), xstrings.Bytes2str(ipData)),
	).Commit()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !res.Succeeded { // 事务成功
		return nil, errors.New("fault to save worker")
	}
	return ipNode, nil
}
