package etcdv3

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func newClient(config *Config) (*clientv3.Client, error) {
	conf := clientv3.Config{
		Endpoints:            config.Endpoints,
		DialTimeout:          config.DialTimeout,
		DialKeepAliveTime:    config.DialKeepAliveTime,
		DialKeepAliveTimeout: config.DialKeepAliveTimeout,
		Username:             config.Username,
		Password:             config.Password,
		AutoSyncInterval:     config.AutoSyncInterval,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(),
		},
	}

	if len(config.dialOptions) > 0 {
		conf.DialOptions = append(conf.DialOptions, config.dialOptions...)
	}

	if config.TlsEnable {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		if config.CAFile != "" {
			certBytes, err := ioutil.ReadFile(config.CAFile)
			if err != nil {
				return nil, err
			}

			caCertPool := x509.NewCertPool()
			ok := caCertPool.AppendCertsFromPEM(certBytes)

			if ok {
				tlsConfig.RootCAs = caCertPool
			}
		}

		if config.CertFile != "" && config.KeyFile != "" {
			tlsCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{tlsCert}
		}
		conf.TLS = tlsConfig
	}

	cli, err := clientv3.New(conf)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

// NewClient
func NewClient(config *Config) (*clientv3.Client, error) {
	return config.Build()
}
