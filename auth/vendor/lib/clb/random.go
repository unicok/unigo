package clb

import (
	"fmt"
	"math/rand"
)

func NewRandomClb(lib Lookup) *RandomClb {
	lb := new(RandomClb)
	lb.dnsLib = lib
	return lb
}

type RandomClb struct {
	dnsLib Lookup
}

func (lb *RandomClb) GetAddress(name string) (Address, error) {
	add := Address{}

	srvs, err := lb.dnsLib.LookupSRV(name)
	if err != nil {
		return add, err
	}
	if len(srvs) == 0 {
		return add, fmt.Errorf("no SRV records found")
	}

	//	log.Printf("%+v", srvs)
	srv := srvs[rand.Intn(len(srvs))]

	// ip, err := lb.dnsLib.LookupA(srv.Target)
	ip, err := lb.dnsLib.LookupA(name)
	if err != nil {
		return add, err
	}

	return Address{Address: ip, Port: srv.Port}, nil
}
