// Package consul is registered as a service using consul.
package consul

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zhufuyi/dtmdriver-sponge/pkg/consulcli"
	"github.com/zhufuyi/dtmdriver-sponge/pkg/servicerd/registry"

	"github.com/hashicorp/consul/api"
)

var (
	_ registry.Registry  = &Registry{}
	_ registry.Discovery = &Registry{}
)

// Option is consul registry option.
type Option func(*Registry)

// WithHealthCheck with registry health check option.
func WithHealthCheck(enable bool) Option {
	return func(o *Registry) {
		o.enableHealthCheck = enable
	}
}

// Config is consul registry config
type Config struct {
	*api.Config
}

// Registry is consul registry
type Registry struct {
	cli               *Client
	enableHealthCheck bool
	registry          map[string]*serviceSet
	lock              sync.RWMutex
}

// NewRegistry instantiating the consul registry
// Note: If the consulcli.WithConfig(*api.Config) parameter is set, the consulAddr parameter is ignored!
func NewRegistry(consulAddr string, id string, instanceName string, instanceEndpoints []string, opts ...consulcli.Option) (registry.Registry, *registry.ServiceInstance, error) {
	serviceInstance := registry.NewServiceInstance(id, instanceName, instanceEndpoints)

	cli, err := consulcli.Init(consulAddr, opts...)
	if err != nil {
		return nil, nil, err
	}

	return New(cli, WithHealthCheck(true)), serviceInstance, nil
}

// New create a consul registry
func New(apiClient *api.Client, opts ...Option) *Registry {
	r := &Registry{
		cli:               NewClient(apiClient),
		registry:          make(map[string]*serviceSet),
		enableHealthCheck: true,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Register register service
func (r *Registry) Register(ctx context.Context, svc *registry.ServiceInstance) error {
	return r.cli.Register(ctx, svc, r.enableHealthCheck)
}

// Deregister deregister service
func (r *Registry) Deregister(ctx context.Context, svc *registry.ServiceInstance) error {
	// NOTE: invoke the func Deregister will block when err is not nil
	return r.cli.Deregister(ctx, svc.ID)
}

// GetService return service by name
func (r *Registry) GetService(_ context.Context, name string) (services []*registry.ServiceInstance, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	set := r.registry[name]
	if set == nil {
		return nil, fmt.Errorf("service %s not resolved in registry", name)
	}
	ss, _ := set.services.Load().([]*registry.ServiceInstance)
	if ss == nil {
		return nil, fmt.Errorf("service %s not found in registry", name)
	}
	services = append(services, ss...)
	return //nolint
}

// ListServices return service list.
func (r *Registry) ListServices() (allServices map[string][]*registry.ServiceInstance, err error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	allServices = make(map[string][]*registry.ServiceInstance)
	for name, set := range r.registry {
		var services []*registry.ServiceInstance
		ss, _ := set.services.Load().([]*registry.ServiceInstance)
		if ss == nil {
			continue
		}
		services = append(services, ss...)
		allServices[name] = services
	}
	return //nolint
}

// Watch resolve service by name
func (r *Registry) Watch(_ context.Context, name string) (registry.Watcher, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	set, ok := r.registry[name]
	if !ok {
		set = &serviceSet{
			watcher:     make(map[*watcher]struct{}),
			services:    &atomic.Value{},
			serviceName: name,
		}
		r.registry[name] = set
	}

	w := &watcher{
		event: make(chan struct{}, 1),
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.set = set
	set.lock.Lock()
	set.watcher[w] = struct{}{}
	set.lock.Unlock()
	ss, _ := set.services.Load().([]*registry.ServiceInstance)
	if len(ss) > 0 {
		// If the service has a value, it needs to be pushed to the watcher,
		// otherwise the initial data may be blocked forever during the watch.
		w.event <- struct{}{}
	}

	if !ok {
		go r.resolve(set)
	}
	return w, nil
}

func (r *Registry) resolve(ss *serviceSet) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	services, idx, err := r.cli.Service(ctx, ss.serviceName, 0, true)
	cancel()
	if err == nil && len(services) > 0 {
		ss.broadcast(services)
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
		tmpService, tmpIdx, err := r.cli.Service(ctx, ss.serviceName, idx, true)
		cancel()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		if len(tmpService) != 0 && tmpIdx != idx {
			services = tmpService
			ss.broadcast(services)
		}
		idx = tmpIdx
	}
}
