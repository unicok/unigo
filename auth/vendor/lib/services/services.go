package services

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"
	etcdclient "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	etcdHostEnv        = "ETCD_HOST"
	DefaultETCD        = "http://172.17.42.1:2379"
	DefaultServicePath = "/backends"
	defaultNameFile    = "/backends/names"
)

// client is a single connection
type client struct {
	key  string
	conn *grpc.ClientConn
}

// service is a kind of service
type service struct {
	clients []client
	idx     uint32 // for round-robin purpose
}

// all services
type servicePool struct {
	services        map[string]*service
	knownNames      map[string]bool // store names.txt
	enableNameCheck bool
	client          etcdclient.Client
	cbs             map[string][]chan string // service add callback notify
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
	// etcd client
	machines := []string{DefaultETCD}
	if env := os.Getenv(etcdHostEnv); env != "" {
		machines = strings.Split(env, ";")
	}

	// init etcd client
	conf := etcdclient.Config{
		Endpoints: machines,
		Transport: etcdclient.DefaultTransport,
	}
	c, err := etcdclient.New(conf)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}
	p.client = c

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
		p.knownNames[DefaultServicePath+"/"+strings.TrimSpace(v)] = true
	}

	// start connection
	p.connectAll(DefaultServicePath)
}

func (p *servicePool) loadNames() []string {
	kapi := etcdclient.NewKeysAPI(p.client)
	// get the keys under directory
	log.Info("reading names:", defaultNameFile)
	resp, err := kapi.Get(context.Background(), defaultNameFile, nil)
	if err != nil {
		log.Error(err)
		return nil
	}

	// validation check
	if resp.Node.Dir {
		log.Error("names is not a file")
		return nil
	}

	// split names
	return strings.Split(resp.Node.Value, "\n")
}

// connect to all services
func (p *servicePool) connectAll(dir string) {
	kapi := etcdclient.NewKeysAPI(p.client)
	// get the keys under directory
	log.Info("connecting services under:", dir)
	resp, err := kapi.Get(context.Background(), dir, &etcdclient.GetOptions{Recursive: true})
	if err != nil {
		log.Error(err)
		return
	}

	// validation check
	if !resp.Node.Dir {
		log.Error("not a directory")
		return
	}

	for _, node := range resp.Node.Nodes {
		if node.Dir {
			for _, service := range node.Nodes {
				p.addService(service.Key, service.Value)
			}
		}
	}
	log.Info("services add completed")

	go p.watcher()
}

func (p *servicePool) watcher() {
	kapi := etcdclient.NewKeysAPI(p.client)
	w := kapi.Watcher(DefaultServicePath, &etcdclient.WatcherOptions{Recursive: true})
	for {
		resp, err := w.Next(context.Background())
		if err != nil {
			log.Error(err)
			continue
		}
		if resp.Node.Dir {
			continue
		}

		switch resp.Action {
		case "set", "create", "update", "compareAndSwap":
			p.addService(resp.Node.Key, resp.Node.Value)
		case "delete":
			p.removeService(resp.PrevNode.Key)
		}
	}
}

func (p *servicePool) addService(key, val string) {
	p.Lock()
	defer p.Unlock()

	// name check
	serviceName := filepath.Dir(key)
	if p.enableNameCheck && !p.knownNames[serviceName] {
		return
	}

	// try new service kind init
	if p.services[serviceName] == nil {
		p.services[serviceName] = &service{}
	}

	// create service connection
	service := p.services[serviceName]
	if conn, err := grpc.Dial(val, grpc.WithBlock(), grpc.WithInsecure()); err == nil {
		service.clients = append(service.clients, client{key, conn})
		log.Info("service added:", key, "-->", val)
		for k := range p.cbs[serviceName] {
			select {
			case p.cbs[serviceName][k] <- key:
			default:
			}
		}
	} else {
		log.Error("did not connect:", key, "-->", val, " error:", err)
	}
}

func (p *servicePool) removeService(key string) {
	p.Lock()
	defer p.Unlock()

	// name check
	serviceName := filepath.Dir(key)
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
		if service.clients[k].key == key {
			service.clients = append(service.clients[:k], service.clients[k+1:]...)
			log.Info("service removed:", key)
			return
		}
	}
}

func (p *servicePool) getServiceWithID(path, id string) *grpc.ClientConn {
	p.RLock()
	defer p.RUnlock()

	// check existence
	service := p.services[path]
	if service == nil {
		return nil
	}

	if len(service.clients) == 0 {
		return nil
	}

	fullpath := string(path) + "/" + id
	for k := range service.clients {
		if service.clients[k].key == fullpath {
			return service.clients[k].conn
		}
	}

	return nil
}

func (p *servicePool) getService(path string) (conn *grpc.ClientConn, key string) {
	p.RLock()
	defer p.RUnlock()

	// check existence
	service := p.services[path]
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

func (p *servicePool) registerCallback(path string, cb chan string) {
	p.Lock()
	defer p.Unlock()
	if p.cbs == nil {
		p.cbs = make(map[string][]chan string)
	}

	p.cbs[path] = append(p.cbs[path], cb)
	if s, ok := p.services[path]; ok {
		for k := range s.clients {
			cb <- s.clients[k].key
		}
	}
	log.Info("register callback on:", path)
}

// GetService is getting a service in round-robin style
// especially useful for load-balance with state-less services
func GetService(path string) (*grpc.ClientConn, string) {
	conn, key := defaultPool.getService(path)
	return conn, key
}

// GetServieWithID provide a specific key for a service, eg:
// path:/backends/snowflake, id:s1
//
// the full cannonical path for this service is: /backends/snowflake/s1
//
func GetServieWithID(path, id string) *grpc.ClientConn {
	return defaultPool.getServiceWithID(path, id)
}

//RegisterCallback is a warpper of
func RegisterCallback(path string, cb chan string) {
	defaultPool.registerCallback(path, cb)
}
