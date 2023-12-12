// Package etcdcli is use for connecting to the etcd service
package etcdcli

import (
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Init connecting to the etcd service
// Note: If the WithConfig(*clientv3.Config) parameter is set, the endpoints parameter is ignored!
func Init(endpoints []string, opts ...Option) (*clientv3.Client, error) {
	o := defaultOptions()
	o.apply(opts...)

	if o.config != nil {
		return clientv3.New(*o.config)
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints cannot be empty")
	}

	conf := clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout:          o.dialTimeout,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		DialOptions:          []grpc.DialOption{grpc.WithBlock()},
		AutoSyncInterval:     o.autoSyncInterval,
		Logger:               o.logger,
		Username:             o.username,
		Password:             o.password,
	}

	if !o.isSecure {
		conf.DialOptions = append(conf.DialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		cred, err := credentials.NewClientTLSFromFile(o.certFile, o.serverNameOverride)
		if err != nil {
			return nil, fmt.Errorf("NewClientTLSFromFile error: %v", err)
		}
		conf.DialOptions = append(conf.DialOptions, grpc.WithTransportCredentials(cred))
	}

	cli, err := clientv3.New(conf)
	if err != nil {
		return nil, fmt.Errorf("connecting to the etcd service error: %v", err)
	}

	return cli, nil
}
