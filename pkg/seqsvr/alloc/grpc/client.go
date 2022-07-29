package grpc

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/imkuqin-zw/uuid-generator/pkg/genproto/api"
	grpcimpl "github.com/imkuqin-zw/uuid-generator/pkg/genproto/api/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/client/grpc"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	grpc2 "google.golang.org/grpc"
)

func newClient(ctx context.Context, name string, opts ...grpc2.DialOption) api.AllocClient {
	c := &grpc.Config{}
	if err := config.Scan(fmt.Sprintf("yggdrasil.client.{%s}.grpc", name), c); err != nil {
		log.Fatalf("fault to get config, err: %s", err.Error())
		return nil
	}
	c.Balancer = "seqsvr"
	uri, err := url.Parse(c.Target)
	if err != nil {
		log.Fatalf("target format fault, err: %s", err.Error())
		return nil
	}
	if len(uri.Scheme) == 0 {
		uri.Scheme = "seqsvr"
		c.Target = uri.String()
	} else if uri.Scheme != "seqsvr" {
		log.Fatal("target format fault: scheme must be seqsvr")
		return nil
	}
	c.Name = name
	if len(name) == 0 {
		c.Name = "com.github.imkuqin_zw.uuid-generator.seqsvr.allloc"
	}
	c.WithUnaryInterceptor(routerVersionInterceptor)
	c.WithDialOption(opts...)
	c.UnaryFilter = append(config.GetStringSlice("yggdrasil.grpc.unaryFilter"), c.UnaryFilter...)
	c.StreamFilter = append(config.GetStringSlice("yggdrasil.grpc.streamFilter"), c.StreamFilter...)
	return grpcimpl.NewAllocClient(grpc.DialByConfig(ctx, c))
}

func Dial(name string, opts ...grpc2.DialOption) api.AllocClient {
	return newClient(context.Background(), name, opts...)
}

func DialContext(ctx context.Context, name string, opts ...grpc2.DialOption) api.AllocClient {
	return newClient(ctx, name, opts...)
}
