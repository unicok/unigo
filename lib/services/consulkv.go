package services

import (
	"fmt"
	"net/url"

	log "github.com/Sirupsen/logrus"
	consulapi "github.com/hashicorp/consul/api"

	"lib/utils"
)

func NewDefaultKVAPI() *ConsulKV {
	u := fmt.Sprintf("http://%s:%v", consulHost, consulAPIPort)
	uri, err := url.Parse(u)
	if err != nil {
		log.Panic("bad consul api uri: " + u)
	}
	return NewKVAPI(uri)
}

func NewKVAPI(uri *url.URL) *ConsulKV {
	config := consulapi.DefaultConfig()
	config.Address = consulHost
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consulkv: ", uri.Path)
	}

	return &ConsulKV{client: client, path: uri.Path}
}

type ConsulKV struct {
	client *consulapi.Client
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

func (r *ConsulKV) GetKVPair(key string, q *consulapi.QueryOptions) (*consulapi.KVPair, *consulapi.QueryMeta, error) {
	path := r.path[1:] + "/" + key
	pair, meta, err := r.client.KV().Get(path, q)
	if err != nil {
		log.Error("consulkv: failed to get value from key: ", path, err)
	}
	return pair, meta, err
}

func (r *ConsulKV) SetKVPair(kv *consulapi.KVPair, q *consulapi.WriteOptions) (*consulapi.WriteMeta, error) {
	kv.Key = r.path[1:] + "/" + kv.Key
	meta, err := r.client.KV().Put(kv, q)
	if err != nil {
		log.Println("consulkv: failed to set value from key:", kv.Key, err)
	}
	return meta, err
}

func (r *ConsulKV) CASKVPair(kv *consulapi.KVPair, q *consulapi.WriteOptions) (bool, *consulapi.WriteMeta, error) {
	kv.Key = r.path[1:] + "/" + kv.Key
	ok, meta, err := r.client.KV().CAS(kv, q)
	if err != nil {
		log.Println("consulkv: failed to set value from key:", kv.Key, err)
	}
	return ok, meta, err
}

func (r *ConsulKV) Get(key string, q *consulapi.QueryOptions) (string, *consulapi.QueryMeta, error) {
	pair, meta, err := r.GetKVPair(key, q)
	if err != nil {
		return "", meta, err
	}
	return utils.Bytes2Str(pair.Value), meta, err
}

func (r *ConsulKV) Set(key, val string, q *consulapi.WriteOptions) (*consulapi.WriteMeta, error) {
	return r.SetKVPair(&consulapi.KVPair{
		Key:   key,
		Value: utils.Str2Bytes(val)}, q)
}

func (r *ConsulKV) CAS(key, val string, modifyIndex uint64, q *consulapi.WriteOptions) (bool, *consulapi.WriteMeta, error) {
	return r.CASKVPair(&consulapi.KVPair{
		Key:         key,
		Value:       utils.Str2Bytes(val),
		ModifyIndex: modifyIndex}, q)
}

func (r *ConsulKV) Del(key string) error {
	path := r.path[1:] + "/" + key
	_, err := r.client.KV().Delete(path, nil)
	if err != nil {
		log.Println("consulkv: failed to del key:", path, err)
	}
	return err
}
