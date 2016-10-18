package services

import (
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/benschw/dns-clb-go/clb"
	"google.golang.org/grpc"
)

const (
	EnvConsulHost     = "CONSUL_HOST"
	EnvConsulDNSPort  = "CONSUL_DNS_PORT"
	DefaultConsulHost = "172.17.0.1"
	DefaultDNSPort    = "53"
	DefaultDnsDomain  = "service.consul"
)

var (
	consulHost    string
	consulDNSPort string
	consulAPIPort string

	lbclient clb.LoadBalancer
)

func init() {
	consulHost = DefaultConsulHost
	if env := os.Getenv(EnvConsulHost); env != "" {
		consulHost = env
	}

	consulDNSPort = DefaultDNSPort
	if env := os.Getenv(EnvConsulDNSPort); env != "" {
		consulDNSPort = env
	}

	lbclient = clb.NewClb(consulHost, consulDNSPort, clb.RoundRobin)
	log.Infof("dns server: %s:%s", consulHost, consulDNSPort)
}

type service struct {
	clients map[string]*grpc.ClientConn
}

// all services
type servicePool struct {
	sync.RWMutex
	services        map[string]*service
	knownNames      map[string]bool // store names.txt
	enableNameCheck bool
}

var (
	defaultPool servicePool
	once        sync.Once
)

// Init ***MUST*** be called before using
func Init(names ...string) {
	once.Do(func() {
		defaultPool.init(names...)
	})
}

func (p *servicePool) init(names ...string) {
	// init
	p.services = make(map[string]*service)
	p.knownNames = make(map[string]bool)

	// names init
	if len(names) == 0 { // names not provided
		names = p.loadNames() //try read from names.txt
	}
	if len(names) > 0 {
		p.enableNameCheck = true
	}

	log.Info("all service names:", names)
	for _, v := range names {
		p.knownNames[strings.TrimSpace(v)] = true
	}
}

func (p *servicePool) loadNames() []string {
	// kapi := etcdclient.NewKeysAPI(p.client)
	// // get the keys under directory
	// log.Info("reading names:", defaultNameFile)
	// resp, err := kapi.Get(context.Background(), defaultNameFile, nil)
	// if err != nil {
	// 	log.Error(err)
	// 	return nil
	// }

	// // validation check
	// if resp.Node.Dir {
	// 	log.Error("names is not a file")
	// 	return nil
	// }

	// // split names
	// return strings.Split(resp.Node.Value, "\n")
	return nil
}

func (p *servicePool) addService(serviceName, addr string) {
	// name check
	if p.enableNameCheck && !p.knownNames[serviceName] {
		return
	}

	// try new service kind init
	if p.services[serviceName] == nil {
		p.services[serviceName] = &service{}
	}

	// create service connection
	service := p.services[serviceName]
	if conn, err := grpc.Dial(addr, grpc.WithBlock(), grpc.WithInsecure()); err == nil {
		service.clients[addr] = conn
		log.Info("service added:", serviceName, "-->", addr)
	} else {
		log.Error("did not connect:", serviceName, "-->", addr, " error:", err)
	}
}

func (p *servicePool) removeService(serviceName, addr string) {
	// name check
	if p.enableNameCheck && !p.knownNames[serviceName] {
		return
	}

	// check service kind
	service := p.services[serviceName]
	if service == nil {
		log.Error("no such service:", serviceName)
		return
	}

	// remove a service
	for k := range service.clients {
		if k == addr {
			delete(service.clients, addr)
			log.Infof("service removed:%s addr:%s", serviceName, addr)
			return
		}
	}
}

// func (p *servicePool) getServiceWithID(serviceName, tag string) *grpc.ClientConn {
// 	p.RLock()
// 	defer p.RUnlock()

// 	// check existence
// 	service := p.services[path]
// 	if service == nil {
// 		return nil
// 	}

// 	if len(service.clients) == 0 {
// 		return nil
// 	}

// 	fullpath := string(path) + "/" + id
// 	for k := range service.clients {
// 		if service.clients[k].key == fullpath {
// 			return service.clients[k].conn
// 		}
// 	}

// 	return nil
// }

func (p *servicePool) getService(serviceName string) (conn *grpc.ClientConn, key string) {
	p.RLock()
	defer p.RUnlock()

	// try new service kind init
	if p.services[serviceName] == nil {
		p.services[serviceName] = &service{}
	}

	service := p.services[serviceName]
	addr, err := GetServiceAddress(serviceName)
	if err != nil {
		log.Error(err)
		return nil, ""
	}

	conn = service.clients[addr]
	if conn != nil {
		return conn, addr
	}

	if addr != "" {
		p.addService(serviceName, addr)
		// get a service in round-robin sytle
		return service.clients[addr], addr
	}

	return nil, addr
}

// GetService is getting a service in round-robin style
// especially useful for load-balance with state-less services
func GetService(srvName string) (*grpc.ClientConn, string) {
	conn, key := defaultPool.getService(srvName)
	return conn, key
}

func GetServiceAddress(srvName string) (string, error) {
	addr, err := lbclient.GetAddress(fmt.Sprintf("%s.%s", srvName, DefaultDnsDomain))
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}
