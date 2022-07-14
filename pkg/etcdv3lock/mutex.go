package etcdv3lock

import (
	"context"
	"time"

	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// Mutex ...
type Mutex struct {
	s *concurrency.Session
	m *concurrency.Mutex
}

func NewMutex(client *clientv3.Client, key string, opts ...concurrency.SessionOption) (mutex *Mutex, err error) {
	mutex = &Mutex{}
	mutex.s, err = concurrency.NewSession(client, opts...)
	if err != nil {
		return
	}
	mutex.m = concurrency.NewMutex(mutex.s, key)
	return
}

// Lock ...
func (mutex *Mutex) Lock(timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return mutex.m.Lock(ctx)
}

// TryLock ...
func (mutex *Mutex) TryLock(timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return mutex.m.Lock(ctx)
}

// Unlock ...
func (mutex *Mutex) Unlock() (err error) {
	err = mutex.m.Unlock(context.TODO())
	if err != nil {
		return
	}
	return mutex.s.Close()
}

// LockWithCtx ...
func (mutex *Mutex) LockWithCtx(ctx context.Context) error {
	return mutex.m.Lock(ctx)
}

// LockWithCtx ...
func (mutex *Mutex) UnLockWithCtx(ctx context.Context) error {
	err := mutex.m.Unlock(ctx)
	if err != nil {
		return err
	}
	return mutex.s.Close()
}
