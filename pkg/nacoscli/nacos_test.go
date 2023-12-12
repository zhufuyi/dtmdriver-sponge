package nacoscli

import (
	"testing"
)

var (
	ipAddr      = "127.0.0.1"
	port        = 8848
	namespaceID = "public"
)

func TestNewNamingClient(t *testing.T) {
	namingClient, err := NewNamingClient(ipAddr, port, namespaceID)
	t.Log(err, namingClient)
}
