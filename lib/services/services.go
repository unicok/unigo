package services

import (
	"lib/services/dns"
	"strings"
	"sync"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
)

// client is a single connection
type client struct {
	key  string
	conn *grpc.ClientConn
}

type service struct {
	clients []client
	idx     uint32
}

// all services
type servicePool struct {
	services        map[string]*service
	knownNames      map[string]bool // store names.txt
	enableNameCheck bool
	sync.RWMutex
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

	// start connection
	p.connectAll()
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

// connect to all services
func (p *servicePool) connectAll() {

	for k := range p.knownNames {
		p.addServices(k)
	}
	log.Info("services add completed")

	go p.watcher()
}

func (p *servicePool) watcher() {
	// kapi := etcdclient.NewKeysAPI(p.client)
	// w := kapi.Watcher(DefaultServicePath, &etcdclient.WatcherOptions{Recursive: true})
	// for {
	// 	resp, err := w.Next(context.Background())
	// 	if err != nil {
	// 		log.Error(err)
	// 		continue
	// 	}
	// 	if resp.Node.Dir {
	// 		continue
	// 	}

	// 	switch resp.Action {
	// 	case "set", "create", "update", "compareAndSwap":
	// 		p.addService(resp.Node.Key, resp.Node.Value)
	// 	case "delete":
	// 		p.removeService(resp.PrevNode.Key)
	// 	}
	// }
}

func (p *servicePool) addServices(serviceName string) {
	// name check
	if p.enableNameCheck && !p.knownNames[serviceName] {
		return
	}

	addrs, err := dns.LookupHP(serviceName)
	if err != nil {
		log.Errorf("failed to resolve service host: %v, %v", serviceName, err)
		return
	}

	// try new service kind init
	if p.services[serviceName] == nil {
		p.services[serviceName] = &service{}
	}

	service := p.services[serviceName]
	for _, addr := range addrs {
		exists := false
		for _, c := range service.clients {
			if c.key == addr {
				exists = true
			}
		}

		if !exists {
			p.addService(serviceName, addr)
		}
	}
}

func (p *servicePool) addService(serviceName, addr string) {
	p.Lock()
	defer p.Unlock()

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
		service.clients = append(service.clients, client{addr, conn})
		log.Info("service added:", serviceName, "-->", addr)
	} else {
		log.Error("did not connect:", serviceName, "-->", addr, " error:", err)
	}
}

func (p *servicePool) removeService(serviceName, addr string) {
	p.Lock()
	defer p.Unlock()

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
		if service.clients[k].key == addr {
			service.clients = append(service.clients[:k], service.clients[k+1:]...)
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

	// check existence
	service := p.services[serviceName]
	if service == nil {
		return nil, ""
	}

	if len(service.clients) == 0 {
		return nil, ""
	}

	// get a service in round-robin sytle
	idx := int(atomic.AddUint32(&service.idx, 1)) % len(service.clients)
	return service.clients[idx].conn, service.clients[idx].key
}

// GetService is getting a service in round-robin style
// especially useful for load-balance with state-less services
func GetService(srvName string) (*grpc.ClientConn, string) {
	conn, key := defaultPool.getService(srvName)
	return conn, key
}

func SearchService(srvName string) ([]string, error) {
	return dns.LookupHP(srvName)
}
