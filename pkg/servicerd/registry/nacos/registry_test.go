package nacos

import (
	"context"
	"testing"
	"time"

	"github.com/zhufuyi/dtmdriver-sponge/pkg/servicerd/registry"
)

func TestNewRegistry(t *testing.T) {
	nacosIPAddr := "127.0.0.1"
	nacosPort := 8848
	nacosNamespaceID := "public"

	id := "serverName_127.0.0.1"
	instanceName := "serverName"
	instanceEndpoints := []string{"grpc://192.168.3.27:8282"}

	iRegistry, instance, err := NewRegistry(nacosIPAddr, nacosPort, nacosNamespaceID, id, instanceName, instanceEndpoints)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(iRegistry, instance)
}

func newNacosRegistry() *Registry {
	return New(getCli(),
		WithPrefix("/micro"),
		WithWeight(1),
		WithCluster("cluster"),
		WithGroup("dev"),
		WithDefaultKind("grpc"),
	)
}

func TestRegistry(t *testing.T) {
	instance := registry.NewServiceInstance("foo", "bar", []string{"grpc://127.0.0.1:8282"})
	r := newNacosRegistry()

	go func() {
		defer func() { recover() }()
		_, err := r.Watch(context.Background(), "foo")
		t.Log(err)
	}()

	defer func() { recover() }()
	time.Sleep(time.Millisecond * 10)
	err := r.Register(context.Background(), instance)
	t.Log(err)
}

func TestDeregister(t *testing.T) {
	instance := registry.NewServiceInstance("foo", "bar", []string{"grpc://127.0.0.1:8282"})
	r := newNacosRegistry()

	defer func() { recover() }()
	err := r.Deregister(context.Background(), instance)
	t.Log(err)
}

func TestGetService(t *testing.T) {
	r := newNacosRegistry()

	defer func() { recover() }()
	_, err := r.GetService(context.Background(), "foo")
	t.Log(err)
}
