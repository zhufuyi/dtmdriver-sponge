package etcd

import (
	"context"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type wt struct{}

func (w wt) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	c := make(chan clientv3.WatchResponse)
	return c
}

func (w wt) RequestProgress(ctx context.Context) error {
	return nil
}

func (w wt) Close() error {
	return nil
}

func newWatch(first bool) *watcher {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	r := New(&clientv3.Client{})

	return &watcher{
		key:         "foo",
		ctx:         ctx,
		cancel:      cancelFunc,
		watchChan:   make(clientv3.WatchChan),
		watcher:     &wt{},
		kv:          r.kv,
		first:       first,
		serviceName: "host",
	}
}

func Test_watcher(t *testing.T) {
	w := newWatch(false)
	instances, err := w.Next()
	t.Log(instances, err)

	go func() {
		defer func() { recover() }()
		w = newWatch(true)
		instances, err = w.Next()
		t.Log(instances, err)
	}()

	go func() {
		defer func() { recover() }()
		instances, err = w.getInstance()
		t.Log(instances, err)
	}()

	time.Sleep(time.Second)

	err = w.Stop()
	t.Log(err)
}
