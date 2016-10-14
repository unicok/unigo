package dns

import (
	"log"
	"testing"
)

const (
	srvName   = "consul.service.consul"
	agentAddr = "120.26.104.246:53"
)

func TestLookupHost(t *testing.T) {
	eps, err := lookupHP(srvName, agentAddr)
	if err != nil {
		t.Error(err)
	}
	for _, ep := range eps {
		log.Print(ep)
	}
}
