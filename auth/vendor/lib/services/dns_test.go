package services

import (
	"log"
	"testing"
)

const (
	srvName   = "mongo.service.consul"
	agentAddr = "172.18.0.1:53"
)

func TestLookupHost(t *testing.T) {
	eps, err := LookupHP(srvName)
	if err != nil {
		t.Error(err)
	}
	for _, ep := range eps {
		log.Print(ep)
	}
}
