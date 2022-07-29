package grpc

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	grpcimpl "github.com/imkuqin-zw/uuid-generator/pkg/genproto/api/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/client/grpc"
	"github.com/stretchr/testify/assert"
)

func Test_RouterChange(t *testing.T) {
	cfg := &grpc.Config{
		Name:     "com.github.imkuqin_zw.uuid-generator.seqsvr.allloc",
		Balancer: "seqsvr",
		Target:   "seqsvr://192.168.3.52:30008",
	}
	cfg.WithUnaryInterceptor(routerVersionInterceptor)
	client := grpcimpl.NewAllocClient(grpc.DialByConfig(context.Background(), cfg))
	w := &sync.WaitGroup{}
	w.Add(1)
	go func() {
		defer w.Done()
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*2)
		defer cancel()
		res, err := client.FetchNext(ctx, &api.FetchSeqNextReq{ID: 5000001})
		if !assert.NotNil(t, err) {
			fmt.Println(res)
		}
	}()
	w.Add(1)
	go func() {
		defer w.Done()
		for i := 0; i < 100; i++ {
			res, err := client.FetchNext(context.TODO(), &api.FetchSeqNextReq{ID: 1})
			assert.Nil(t, err)
			fmt.Println(res)
		}
	}()
	w.Wait()

}

func Test_concurrence(t *testing.T) {
	cfg := &grpc.Config{
		Name:     "com.github.imkuqin_zw.uuid-generator.seqsvr.allloc",
		Balancer: "seqsvr",
		Target:   "seqsvr://192.168.3.52:30008",
	}
	cfg.WithUnaryInterceptor(routerVersionInterceptor)
	client := grpcimpl.NewAllocClient(grpc.DialByConfig(context.Background(), cfg))
	w := &sync.WaitGroup{}
	count := 200
	data := make([]uint64, count)
	for i := 0; i < count; i++ {
		w.Add(1)
		j := i
		go func() {
			defer w.Done()
			res, err := client.FetchNext(context.TODO(), &api.FetchSeqNextReq{ID: 1})
			assert.Nil(t, err)
			data[j] = res.Val
		}()
	}
	w.Wait()
	resMap := make(map[uint64]struct{}, count)
	var exists = false
	for _, item := range data {
		if _, ok := resMap[item]; ok {
			exists = true
			break
		}
		resMap[item] = struct{}{}
	}
	assert.Equal(t, false, exists)
}
