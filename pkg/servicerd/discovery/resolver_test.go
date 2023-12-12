package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/zhufuyi/dtmdriver-sponge/pkg/servicerd/registry"

	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

type cliConn struct {
}

func (c cliConn) UpdateState(state resolver.State) error {
	return nil
}

func (c cliConn) ReportError(err error) {}

func (c cliConn) NewAddress(addresses []resolver.Address) {}

func (c cliConn) NewServiceConfig(serviceConfig string) {}

func (c cliConn) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	return &serviceconfig.ParseResult{}
}

func Test_discoveryResolver(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	r := &discoveryResolver{
		w:                &watcher{},
		cc:               &cliConn{},
		ctx:              ctx,
		cancel:           cancel,
		insecure:         true,
		debugLogDisabled: false,
	}
	defer r.Close()

	r.ResolveNow(resolver.ResolveNowOptions{})
	r.update([]*registry.ServiceInstance{registry.NewServiceInstance(
		"foo",
		"bar",
		[]string{"grpc://127.0.0.1:8282"},
	)})

	r.watch()
	time.Sleep(time.Second)
}
