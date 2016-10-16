package kv

import (
	"net/url"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"

	"lib/services"
	"lib/utils"
)

const (
	EnvConsulHTTPAddr    = "CONSUL_HTTP_ADDR"
	DefaultConsulAPIName = "consul-http"
)

var (
	defaultKVAPI *ConsulKV
	defaultAddr  string
)

func init() {
	if env := os.Getenv(EnvConsulHTTPAddr); env != "" {
		defaultAddr = env
	}

	defaultKVAPI = NewDefaultKVAPI()
}

func NewDefaultKVAPI() *ConsulKV {
	if defaultAddr == "" {
		addr, err := services.GetServiceAddress(DefaultConsulAPIName)
		if err != nil {
			log.Panic("resolve consul domain err: ", err)
		}
		defaultAddr = addr
	}
	uri, err := url.Parse(defaultAddr)
	if err != nil {
		log.Panic("bad consul api uri: " + defaultAddr)
	}
	return NewKVAPI(uri)
}

func NewKVAPI(uri *url.URL) *ConsulKV {
	config := api.DefaultConfig()
	config.Address = uri.Path
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal("consulkv: ", uri.Path)
	}

	log.Debug("new kv api client: ", uri.Path)
	return &ConsulKV{client: client, kv: client.KV(), path: uri.Path}
}

type ConsulKV struct {
	client *api.Client
	kv     *api.KV
	path   string
}

func (r *ConsulKV) Ping() error {
	status := r.client.Status()
	leader, err := status.Leader()
	if err != nil {
		return err
	}
	log.Info("consulkv: current leader ", leader)
	return nil
}

func (r *ConsulKV) GetKVPair(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	key = strings.TrimPrefix(key, "/")
	log.Debugf("get kv pair: %s", key)
	pair, meta, err := r.kv.Get(key, q)
	if err != nil {
		log.Error("consulkv: failed to get value from key: ", key, err)
	}
	log.Debugf("get kv pair: %s result: %s", key, utils.Bytes2Str(pair.Value))
	return pair, meta, err
}

func (r *ConsulKV) SetKVPair(kv *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error) {
	kv.Key = strings.TrimPrefix(kv.Key, "/")
	log.Debug("set kv pair: ", kv.Key)
	meta, err := r.kv.Put(kv, q)
	if err != nil {
		log.Println("consulkv: failed to set value from key:", kv.Key, err)
	}
	return meta, err
}

func (r *ConsulKV) CASKVPair(kv *api.KVPair, q *api.WriteOptions) (bool, *api.WriteMeta, error) {
	kv.Key = strings.TrimPrefix(kv.Key, "/")
	log.Debugf("cas kv pair: %s, modifyindex", kv.Key, kv.ModifyIndex)
	ok, meta, err := r.kv.CAS(kv, q)
	if err != nil {
		log.Println("consulkv: failed to set value from key:", kv.Key, err)
	}
	return ok, meta, err
}

func (r *ConsulKV) Get(key string, q *api.QueryOptions) (string, *api.QueryMeta, error) {
	pair, meta, err := r.GetKVPair(key, q)
	if err != nil {
		return "", meta, err
	}
	return utils.Bytes2Str(pair.Value), meta, err
}

func (r *ConsulKV) Set(key, val string, q *api.WriteOptions) (*api.WriteMeta, error) {
	return r.SetKVPair(&api.KVPair{
		Key:   key,
		Value: utils.Str2Bytes(val)}, q)
}

func (r *ConsulKV) CAS(key, val string, modifyIndex uint64, q *api.WriteOptions) (bool, *api.WriteMeta, error) {
	return r.CASKVPair(&api.KVPair{
		Key:         key,
		Value:       utils.Str2Bytes(val),
		ModifyIndex: modifyIndex}, q)
}

func (r *ConsulKV) Del(key string) error {
	_, err := r.kv.Delete(key, nil)
	if err != nil {
		log.Println("consulkv: failed to del key:", key, err)
	}
	return err
}

func GetKVPair(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	return defaultKVAPI.GetKVPair(key, q)
}

func SetKVPair(kv *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error) {
	return defaultKVAPI.SetKVPair(kv, q)
}

func CASKVPair(kv *api.KVPair, q *api.WriteOptions) (bool, *api.WriteMeta, error) {
	return defaultKVAPI.CASKVPair(kv, q)
}

func Get(key string, q *api.QueryOptions) (string, *api.QueryMeta, error) {
	return defaultKVAPI.Get(key, q)
}

func Set(key, val string, q *api.WriteOptions) (*api.WriteMeta, error) {
	return defaultKVAPI.Set(key, val, q)
}

func CAS(key, val string, modifyIndex uint64, q *api.WriteOptions) (bool, *api.WriteMeta, error) {
	return defaultKVAPI.CAS(key, val, modifyIndex, q)
}

func Del(key string) error {
	return defaultKVAPI.Del(key)
}
