package main

import (
	"errors"
	"reflect"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// a subscribers carries a callback funcion
// which must be of type:
// func(msg interface{})
type subscriber struct {
	f interface{}
}

func NewSubscriber(f interface{}) *subscriber {
	return &subscriber{f: f}
}

// PubSub for an endpoint
type PubSub struct {
	chMsg chan interface{}
	subs  []*subscriber
	sync.Mutex
}

func (ps *PubSub) init() {
	ps.chMsg = make(chan interface{})
	call := func(rf reflect.Value, in []reflect.Value) {
		defer func() { // do not let callback panic the pbsub
			if err := recover(); err != nil {
				log.Error(err)
			}
		}()
		rf.Call(in)
	}

	go func() {
		for v := range ps.chMsg {
			rv := reflect.ValueOf(v)
			ps.Lock()
			for _, sub := range ps.subs {
				rf := reflect.ValueOf(sub.f)
				if rv.Type() == reflect.ValueOf(sub.f).Type().In(0) { // parameter type match
					call(rf, []reflect.Value{rv})
				}
			}
			ps.Unlock()
		}
	}()
}

// Sub subscribe to the PubSub.
func (ps *PubSub) Sub(s *subscriber) error {
	// check wrapped function
	rf := reflect.ValueOf(s.f)
	if rf.Kind() != reflect.Func {
		return errors.New("Not a function")
	}
	if rf.Type().NumIn() != 1 {
		return errors.New("Number of arguments should be 1")
	}

	ps.Lock()
	defer ps.Unlock()
	ps.subs = append(ps.subs, s)
	return nil
}

// Unsubscribe to the PubSub.
func (ps *PubSub) Unsub(s *subscriber) {
	ptr := reflect.ValueOf(s).Pointer()
	ps.Lock()
	defer ps.Unlock()
	result := make([]*subscriber, 0, len(ps.subs)-1)
	last := 0
	for i, v := range ps.subs {
		if reflect.ValueOf(v).Pointer() == ptr {
			result = append(result, ps.subs[last:i]...)
			last = i + 1
		}
	}
	ps.subs = append(result, ps.subs[last:]...)
}

// Pub publish to the PubSub.
func (ps *PubSub) Pub(v interface{}) {
	ps.chMsg <- v
}

// Close closes PubSub.
// To inspect unbsubscribing for another subscruber,
// you must create message structure to notify them.
// After publish notifycations, Close should be called.
func (ps *PubSub) Close() {
	close(ps.chMsg)
}
