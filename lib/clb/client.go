package clb

import (
	"fmt"
)

type LoadBalancerType int

const (
	Random     LoadBalancerType = iota
	RoundRobin LoadBalancerType = iota
)

type CacheType int

const (
	None CacheType = iota
	Ttl  CacheType = iota
)

type LoadBalancer interface {
	GetAddress(name string) (Address, error)
}

func New() LoadBalancer {
	return NewDefaultClb(RoundRobin)
}

func NewDefaultClb(lbType LoadBalancerType) LoadBalancer {
	lib := NewDefaultLookupLib()

	return buildClb(lib, lbType)
}

func NewClb(address string, port string, lbType LoadBalancerType) LoadBalancer {
	lib := NewLookupLib(fmt.Sprintf("%s:%s", address, port))

	return buildClb(lib, lbType)
}

func NewTtlCacheClb(address string, port string, lbType LoadBalancerType, ttl int) LoadBalancer {
	lib := NewLookupLib(fmt.Sprintf("%s:%s", address, port))
	cache := NewTtlCache(lib, ttl)

	return buildClb(cache, lbType)
}

func buildClb(lib Lookup, lbType LoadBalancerType) LoadBalancer {
	switch lbType {
	case RoundRobin:
		return NewRoundRobinClb(lib)
	case Random:
		return NewRandomClb(lib)
	}
	return nil
}
