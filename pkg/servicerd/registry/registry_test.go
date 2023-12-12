package registry

import (
	"testing"
)

func TestNewServiceInstance(t *testing.T) {
	s := NewServiceInstance("foo", "bar", []string{"grpc://127.0.0.1:8282"},
		WithVersion("v1.0.0"),
		WithMetadata(map[string]string{"foo": "bar"}),
	)
	t.Log(s)
}
