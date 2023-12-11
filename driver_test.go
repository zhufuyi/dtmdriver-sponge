package driver

import (
	"testing"
	"time"
)

func Test_parseTarget(t *testing.T) {
	targets := []string{
		"consul://127.0.0.1:8500/dtmservice",
		"consul://foobar.com:8500/dtmservice?token=your-token",

		"etcd://127.0.0.1:2379/dtmservice",
		"etcd://foobar.com:2379/dtmservice?username=your-username&password=your-password",

		"nacos://127.0.0.1:8848/dtmservice",
		"nacos://foobar.com:8848/dtmservice?namespaceID=3454d2b5-2455",
		"nacos://foobar.com:8848/dtmservice?namespaceID=3454d2b5-2455&username=your-username&password=your-password",
	}

	for _, target := range targets {
		cfg, err := parseTarget(target)
		if err != nil {
			t.Error(err)
		}
		switch cfg.Type {
		case consulType:
			t.Log(cfg.Type, cfg.name, cfg.consul)
		case etcdType:
			t.Log(cfg.Type, cfg.name, cfg.etcd)
		case nacosType:
			t.Log(cfg.Type, cfg.name, cfg.nacos)
		}
	}
}

func Test_parseEndpoint(t *testing.T) {
	endpoints := []string{
		"grpc://127.0.0.1:36790",
		"grpc://foobar.com:36790",

		"http://127.0.0.1:36789",
		"http://foobar.com:36789",
	}

	for _, endpoint := range endpoints {
		mark, err := parseEndpoint(endpoint)
		if err != nil {
			t.Error(err)
		}
		t.Log(mark)
	}
}

func TestSpongeDriver_RegisterService(t *testing.T) {
	var (
		targetType  = "etcd"
		dtmEndpoint = "grpc://127.0.0.1:36790"
		consulDsn   = "consul://127.0.0.1:8500/dtmservice"
		etcdDsn     = "etcd://127.0.0.1:2379/dtmservice"
		nacosDsn    = "nacos://127.0.0.1:8848/dtmservice"
		err         error
	)

	d := new(SpongeDriver)
	t.Log(d.GetName())

	switch targetType {
	case consulType:
		// consul
		err = d.RegisterService(consulDsn, dtmEndpoint)
	case etcdType:
		// etcd
		err = d.RegisterService(etcdDsn, dtmEndpoint)
	case nacosType:
		// nacos
		_ = d.RegisterService(nacosDsn, dtmEndpoint)
	}

	if err != nil {
		t.Errorf("register dtm service to %v failed, err=%+v", targetType, err)
		return
	}

	t.Log("register dtm service to", targetType, "success")
	time.Sleep(time.Minute)
}
