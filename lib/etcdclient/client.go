package etcdclient

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	etcdclient "github.com/coreos/etcd/client"
)

const (
	defaultETCD = "http://172.17.42.1:2379"
)

var (
	machines []string
	client   etcdclient.Client
)

func init() {
	// etcd client
	machines = []string{defaultETCD}
	if env := os.Getenv("ETCD_HOST"); env != "" {
		machines = strings.Split(env, ";")
	}

	// config
	cfg := etcdclient.Config{
		Endpoints: machines,
		Transport: etcdclient.DefaultTransport,
	}

	// create client
	c, err := etcdclient.New(cfg)
	if err != nil {
		log.Error(err)
		return
	}
	client = c
}

// KeysAPI is return etcd keys api handler
func KeysAPI() etcdclient.KeysAPI {
	return etcdclient.NewKeysAPI(client)
}
