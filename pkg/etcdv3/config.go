package etcdv3

import (
	"fmt"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type Config struct {
	Endpoints            []string
	Username             string
	Password             string
	DialTimeout          time.Duration
	DialKeepAliveTime    time.Duration
	DialKeepAliveTimeout time.Duration
	AutoSyncInterval     time.Duration
	TlsEnable            bool
	CertFile             string
	KeyFile              string
	CAFile               string

	dialOptions []grpc.DialOption
}

func StdConfig(name string) *Config {
	return RawConfig(fmt.Sprintf("%s.%s", "etcdv3", name))
}

func RawConfig(key string) *Config {
	c := new(Config)
	if err := config.Get(key).Scan(c); err != nil {
		log.Fatalf("fault to scan etcdv3 config, errors: %+v", err)
		return nil
	}
	return c
}

//WithDialOptions
func (c *Config) WithDialOptions(opts ...grpc.DialOption) {
	if c.dialOptions == nil {
		c.dialOptions = make([]grpc.DialOption, 0, len(opts))
	}
	c.dialOptions = append(c.dialOptions, opts...)
}

//Build
func (c *Config) Build() (*clientv3.Client, error) {
	return newClient(c)
}
