package etcdcli

import (
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	endpoints := []string{"127.0.0.1:2379"}
	cli, err := Init(endpoints,
		WithDialTimeout(time.Second*2),
		WithAuth("", ""),
		WithAutoSyncInterval(0),
		WithLog(zap.NewNop()),
	)
	t.Log(err, cli)

	cli, err = Init(nil, WithConfig(&clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 2,
		Username:    "",
		Password:    "",
	}))
	t.Log(err, cli)

	// test error
	_, err = Init(endpoints,
		WithDialTimeout(time.Second),
		WithSecure("foo", "notfound.crt"))
	t.Log(err)
	endpoints = nil
	_, err = Init(endpoints)
	t.Log(err)
}
