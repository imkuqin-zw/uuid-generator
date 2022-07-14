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
