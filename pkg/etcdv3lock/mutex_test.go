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

package etcdv3lock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func Test_MutexLock(t *testing.T) {
	config := clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	}
	client, err := clientv3.New(config)
	if err != nil {
		t.Fatal(err)
	}
	etcdMutex1, err := NewMutex(
		client,
		"/test/lock",
		concurrency.WithTTL(int(3)))
	assert.Nil(t, err)

	err = etcdMutex1.Lock(time.Second * 1)
	assert.Nil(t, err)
	defer etcdMutex1.Unlock()

	// Grab the lock
	etcdMutex, err := NewMutex(
		client,
		"/test/lock",
		concurrency.WithTTL(int(3)))
	assert.Nil(t, err)
	defer etcdMutex.Unlock()

	err = etcdMutex.Lock(time.Second * 1)
	assert.NotNil(t, err)
}
