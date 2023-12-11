package driver

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/zhufuyi/sponge/pkg/consulcli"
	"github.com/zhufuyi/sponge/pkg/etcdcli"
	"github.com/zhufuyi/sponge/pkg/nacoscli"
	"github.com/zhufuyi/sponge/pkg/servicerd/discovery"
	"github.com/zhufuyi/sponge/pkg/servicerd/registry"
	"github.com/zhufuyi/sponge/pkg/servicerd/registry/consul"
	"github.com/zhufuyi/sponge/pkg/servicerd/registry/etcd"
	"github.com/zhufuyi/sponge/pkg/servicerd/registry/nacos"

	"github.com/dtm-labs/dtmdriver"
	"google.golang.org/grpc/resolver"
)

func init() {
	dtmdriver.Register(&SpongeDriver{})
}

const (
	DriverName = "dtm-driver-sponge"
	consulType = "consul"
	etcdType   = "etcd"
	nacosType  = "nacos"
)

// SpongeDriver is a dtm driver for sponge
type SpongeDriver struct{}

// GetName returns the driver name
func (d *SpongeDriver) GetName() string {
	return DriverName
}

// RegisterAddrResolver register addr resolver
func (d *SpongeDriver) RegisterAddrResolver() {

}

// RegisterService register dtm service and resolver your service
func (d *SpongeDriver) RegisterService(target string, endpoint string) error {
	if target == "" {
		return nil
	}

	c, err := parseTarget(target)
	if err != nil {
		return err
	}
	mark, err := parseEndpoint(endpoint)
	if err != nil {
		return err
	}
	id := c.name + "_" + mark

	// register dtm service to consul, etcd, nacos
	err = c.register(endpoint, id)
	if err != nil {
		return err
	}

	// resolver your service from consul, etcd, nacos
	return c.resolver()
}

// ParseServerMethod parse server and method
func (d *SpongeDriver) ParseServerMethod(uri string) (server string, method string, err error) {
	if !strings.Contains(uri, "//") {
		sep := strings.IndexByte(uri, '/')
		if sep == -1 {
			return "", "", fmt.Errorf("bad url: '%s'. no '/' found", uri)
		}
		return uri[:sep], uri[sep:], nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return "", "", nil
	}
	index := strings.IndexByte(u.Path[1:], '/') + 1
	return u.Scheme + "://" + u.Host + u.Path[:index], u.Path[index:], nil
}

// -----------------------------------------------------------------------------------------

type consulConfig struct {
	addr  string // includes port, e.g. 127.0.0.1:8500
	token string
}

type etcdConfig struct {
	addrs    []string // includes port, e.g. [127.0.0.1:2379]
	username string
	password string
}

type nacosConfig struct {
	host        string // excluding port, e.g. 127.0.0.1
	port        int
	namespaceID string
	username    string
	password    string
}

type driverConfig struct {
	Type string // consul, etcd, nacos
	name string // default name is dtmservice

	consul *consulConfig
	etcd   *etcdConfig
	nacos  *nacosConfig
}

// register dtm service to consul, etcd, nacos
func (c *driverConfig) register(instanceEndpoint string, id string) error {
	var (
		err       error
		iRegistry registry.Registry
		instance  *registry.ServiceInstance
	)

	switch c.Type {
	case consulType:
		iRegistry, instance, err = consul.NewRegistry(
			c.consul.addr,
			id,
			c.name,
			[]string{instanceEndpoint},
			consulcli.WithToken(c.consul.token),
		)
		if err != nil {
			return err
		}

	case etcdType:
		iRegistry, instance, err = etcd.NewRegistry(
			c.etcd.addrs,
			id,
			c.name,
			[]string{instanceEndpoint},
			etcdcli.WithAuth(c.etcd.username, c.etcd.password),
		)
		if err != nil {
			return err
		}

	case nacosType:
		iRegistry, instance, err = nacos.NewRegistry(
			c.nacos.host,
			c.nacos.port,
			c.nacos.namespaceID,
			id,
			c.name,
			[]string{instanceEndpoint},
			nacoscli.WithAuth(c.nacos.username, c.nacos.password),
		)
	}
	if err != nil {
		return err
	}

	return iRegistry.Register(context.Background(), instance)
}

// resolver discovery:///your-service-name from consul, etcd, nacos
func (c *driverConfig) resolver() error {
	var iDiscovery registry.Discovery

	switch c.Type {
	case consulType:
		cli, err := consulcli.Init(c.consul.addr)
		if err != nil {
			return err
		}
		iDiscovery = consul.New(cli)

	case etcdType:
		cli, err := etcdcli.Init(c.etcd.addrs)
		if err != nil {
			return err
		}
		iDiscovery = etcd.New(cli)

	case nacosType:
		cli, err := nacoscli.NewNamingClient(
			c.nacos.host,
			c.nacos.port,
			c.nacos.namespaceID)
		if err != nil {
			return err
		}
		iDiscovery = nacos.New(cli)
	}

	builder := discovery.NewBuilder(iDiscovery, discovery.WithInsecure(true))
	// register a global resolver so that the dtmservice can resolve discovery:///your-service-name.
	resolver.Register(builder)
	return nil
}

func parseTarget(target string) (*driverConfig, error) {
	cfg := &driverConfig{name: "dtmservice"}

	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	cfg.Type = u.Scheme
	if cfg.Type != consulType && cfg.Type != etcdType && cfg.Type != nacosType {
		return nil, fmt.Errorf("invalid registry type: %s, only supports consul, etcd, nacos,"+
			" e.g. consul://127.0.0.1:8500/dtmservice", cfg.Type)
	}

	if u.Path != "" {
		pathParts := strings.Split(u.Path, "/")
		cfg.name = pathParts[len(pathParts)-1]
	}

	var (
		port       int
		host, addr string
	)
	if u.Port() == "" {
		err = fmt.Errorf("port is empty: %s", u.Host)
	} else {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, err
		}
		if u.Host == "" {
			return nil, fmt.Errorf("registry address is empty")
		}
		host = strings.ReplaceAll(u.Host, ":"+u.Port(), "")
		addr = u.Host
	}

	params := u.Query()
	if err != nil {
		return nil, err
	}

	switch cfg.Type {
	case consulType:
		token := params.Get("token")
		cfg.consul = &consulConfig{
			addr:  addr,
			token: token,
		}
	case etcdType:
		username := params.Get("username")
		password := params.Get("password")
		cfg.etcd = &etcdConfig{
			addrs:    []string{addr},
			username: username,
			password: password,
		}
	case nacosType:
		username := params.Get("username")
		password := params.Get("password")
		namespaceID := params.Get("namespaceID")
		cfg.nacos = &nacosConfig{
			host:        host,
			port:        port,
			namespaceID: namespaceID,
			username:    username,
			password:    password,
		}
	}

	return cfg, nil
}

func parseEndpoint(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	if u.Scheme != "grpc" && u.Scheme != "http" {
		return "", fmt.Errorf("invalid dtm service protocol: %s, only supports grpc, http, e.g. grpc://localhost:36790", u.Scheme)
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", err
	}

	return u.Scheme + "_" + host, nil
}
