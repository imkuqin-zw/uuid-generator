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
