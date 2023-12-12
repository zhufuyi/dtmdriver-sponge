package driver

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/dtm-labs/dtmdriver"
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
