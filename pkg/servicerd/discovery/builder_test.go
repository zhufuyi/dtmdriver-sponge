package discovery

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/zhufuyi/dtmdriver-sponge/pkg/servicerd/registry"

	"google.golang.org/grpc/resolver"
)

type discovery struct{}

func (d discovery) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	return []*registry.ServiceInstance{}, nil
}

func (d discovery) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return &watcher{}, nil
}

type watcher struct{}

func (w watcher) Next() ([]*registry.ServiceInstance, error) {
	return []*registry.ServiceInstance{}, nil
}

func (w watcher) Stop() error {
	return nil
}

func TestBuilder(t *testing.T) {
	b := NewBuilder(&discovery{},
		WithInsecure(false),
		WithTimeout(time.Second),
		DisableDebugLog(),
	)

	u := url.URL{
		Path: "ipv4.single.fake",
	}
	_, err := b.Build(resolver.Target{URL: u}, nil, resolver.BuildOptions{})
	t.Log(err)
}
