package nacoscli

import (
	"os"
	"testing"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
)

var (
	ipAddr      = "127.0.0.1"
	port        = 8848
	namespaceID = "3454d2b5-2455-4d0e-bf6d-e033b086bb4c"
)

func TestParse(t *testing.T) {
	params := &Params{
		IPAddr:      ipAddr,
		Port:        uint64(port),
		NamespaceID: namespaceID,
		Group:       "dev",
		DataID:      "serverNameExample.yml",
		Format:      "yaml",
	}

	format, data, err := GetConfig(params)
	t.Log(err, format, data)

	params = &Params{
		Group:  "dev",
		DataID: "serverNameExample.yml",
		Format: "yaml",
	}
	clientConfig := &constant.ClientConfig{
		NamespaceId:         namespaceID,
		TimeoutMs:           1000,
		NotLoadCacheAtStart: true,
		LogDir:              os.TempDir() + "/nacos/log",
		CacheDir:            os.TempDir() + "/nacos/cache",
	}
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: ipAddr,
			Port:   uint64(port),
		},
	}

	format, data, err = GetConfig(params,
		WithClientConfig(clientConfig),
		WithServerConfigs(serverConfigs),
		WithAuth("foo", "bar"),
	)
	t.Log(err, format, data)
}

func TestNewNamingClient(t *testing.T) {
	namingClient, err := NewNamingClient(ipAddr, port, namespaceID)
	t.Log(err, namingClient)
}
