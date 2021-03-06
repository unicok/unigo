package kv

import (
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestConsulAPI(t *testing.T) {
	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	if err != nil {
		t.Error(err)
	}

	t.Log(config.Scheme)
	t.Log(config.Address)
	t.Log(config.Datacenter)

	kv := client.KV()

	p := &api.KVPair{Key: "foo", Value: []byte("test")}
	client.KV().Put(p, nil)
	kv.Put(p, nil)
}
