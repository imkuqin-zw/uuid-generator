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
	"context"
	"os"
	"sync"
	"time"

	"github.com/imkuqin-zw/yggdrasil"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/pkg/errors"
)

const (
	maxWorkID   = uint64(-1 ^ (-1 << 5))
	maxCenterID = uint64(-1 ^ (-1 << 5))
	maxSeqNum   = uint64(-1 ^ (-1 << 12))
)

type Options struct {
	WorkID       uint64
	DataCenterID uint64
}

type Snowflake struct {
	sync.Mutex
	workID       uint64
	dataCenterID uint64
	seqNum       uint64
	lastMilli    uint64
}

// GetUID 得到全局唯一ID int64类型
// 首位0(1位) + 毫秒时间戳(41位) + 数据中心标识(5位) + 工作机器标识(5位) + 自增id(12位)
// 时间可以保证400年不重复
// 数据中心和机器标识一起标识节点，最多支持1024个节点
// 每个节点每一毫秒能生成最多4096个id
// 63      62            21            16        11       0
// +-------+-------------+-------------+---------+--------+
// | 未使用 | 毫秒级时间戳 | 数据中心标识 | 工作机器 | 自增id |
// +-------+-------------+-------------+---------+--------+
func (s *Snowflake) FetchNext() (uint64, error) {
	s.Lock()
	defer s.Unlock()
	now := s.getUnixMill()
	if now > s.lastMilli {
		s.seqNum = 0
		s.lastMilli = now
	} else if now == s.lastMilli {
		s.seqNum++
		if s.seqNum > maxSeqNum {
			for now <= s.lastMilli {
				now = s.getUnixMill()
			}
		}
	} else {
		return 0, errors.New("clock back")
	}
	return now<<22 | s.dataCenterID<<17 | s.workID<<12 | (s.seqNum & maxSeqNum), nil
}

func (s *Snowflake) getUnixMill() uint64 {
	return uint64(time.Now().UnixNano() / int64(time.Millisecond))
}

// 初始化节点标识
func NewWorker(coordinator Coordinator) *Snowflake {
	dc := uint64(config.GetInt64("snowflake.dc", 0))
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*30)
	defer cancel()
	worker, err := coordinator.GetOrCreateWorker(ctx)
	if err != nil {
		log.Fatalf("fault to create worker, err: %+v", err)
		return nil
	}
	threshold := config.GetDuration("snowflake.maxClockBackThreshold", time.Second)
	subTime := time.Now().Sub(time.Unix(worker.LastTime/1000, worker.LastTime%1000*1000))
	if subTime < 0 && subTime < -threshold {
		time.Sleep(subTime)
	}
	go func() {
		d := time.Second
		t := time.NewTimer(d)
		for {
			select {
			case <-t.C:
				f := func() error {
					ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
					defer cancel()
					if err := coordinator.SyncTime(ctx); err != nil {
						return err
					}
					log.Info("sync time")
					return nil
				}
				backoff := time.Millisecond * 500
				counter := 0
				for {
					if err := f(); err != nil {
						if counter < 3 {
							counter++
							time.Sleep(backoff)
							log.Warnf("fault to sync time, err: %+v", err)
							continue
						}
						log.Errorf("fault to sync time, err: %+v", err)
						if err := yggdrasil.Stop(); err != nil {
							os.Exit(-1)
						}
					}
					break
				}
				t.Reset(d)
			}
		}

	}()
	return &Snowflake{
		workID:       *worker.WorkerID,
		dataCenterID: dc,
	}
}
