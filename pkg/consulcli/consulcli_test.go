package consulcli

import (
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
)

func TestInit(t *testing.T) {
	addr := "127.0.0.1:8500"
	cli, err := Init(addr,
		WithScheme("http"),
		WithWaitTime(time.Second*2),
		WithDatacenter(""),
		WithToken("your-token"),
	)
	t.Log(err, cli)

	cli, err = Init("", WithConfig(&api.Config{
		Address:    addr,
		Scheme:     "http",
		WaitTime:   time.Second * 2,
		Datacenter: "",
	}))
	t.Log(err, cli)
}
