package registry

import "sync"

type Registry struct {
	records map[int32]interface{} // id -> v
	sync.RWMutex
}

var (
	defaultRegistry Registry
)

func init() {
	defaultRegistry.init()
}

func (r *Registry) init() {
	r.records = make(map[int32]interface{})
}

// register a user
func (r *Registry) Registry(id int32, v interface{}) {
	r.Lock()
	defer r.Unlock()
	r.records[id] = v
}

// unregister a user
func (r *Registry) Unregistry(id int32) {
	r.Lock()
	defer r.Unlock()
	delete(r.records, id)
}

// query a user
func (r *Registry) Query(id int32) (x interface{}) {
	r.RLock()
	defer r.RUnlock()
	x = r.records[id]
	return
}

// return count of online users
func (r *Registry) Count() (count int) {
	r.RLock()
	defer r.RUnlock()
	count = len(r.records)
	return
}

func Register(id int32, v interface{}) {
	defaultRegistry.Registry(id, v)
}

func Unregister(id int32) {
	defaultRegistry.Unregistry(id)
}

func Query(id int32) interface{} {
	return defaultRegistry.Query(id)
}

func Count() int {
	return defaultRegistry.Count()
}
