package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/miekg/dns"
)

const (
	EnvDns         = "DNS_ADDR"
	DefaultDnsAddr = "172.17.42.1:53"
)

var (
	dnsAddr string
)

func init() {
	dnsAddr = DefaultDnsAddr
	if env := os.Getenv(EnvDns); env != "" {
		dnsAddr = env
	}
}

// LookupHost query service address and port from dns server
func LookupHP(srv string) ([]string, error) {
	return lookupHP(srv, dnsAddr)
}

func lookupHP(srv, ds string) ([]string, error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(srv), dns.TypeSRV)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, ds)
	if err != nil {
		return nil, err
	}

	if r == nil || r.Rcode != dns.RcodeSuccess {
		return nil, errors.New(fmt.Sprint("failed dns query ", ds))
	}

	var eps []string
	for _, a := range r.Answer {
		if b, ok := a.(*dns.SRV); ok {
			m.SetQuestion(dns.Fqdn(srv), dns.TypeA)
			r1, _, err := c.Exchange(m, ds)
			if err != nil || r1 == nil {
				continue
			}
			for _, a1 := range r1.Answer {
				if c, ok := a1.(*dns.A); ok {
					eps = append(eps, fmt.Sprintf("%s:%v", c.A, b.Port))
				}
			}
		}
	}
	return eps, nil
}
