package main

import (
	"errors"
	"sync"
	"time"

	pb "lib/proto/chat"
)

const (
	SERVICE = "[CHAT]"
)

const (
	BoltDBFile    = "/data/CHAT.DAT"
	BoltDBBucket  = "EPS"
	MaxQueueSize  = 128 // num of message kept
	PendingSize   = 65536
	CheckInterval = time.Minute
)

var (
	Ok                 = &pb.ChatNil{}
	ErrorAlreadyExists = errors.New("id already exists")
	ErrorNotExists     = errors.New("id not exists")
)

// Endpoint definition
type EndPoint struct {
	inbox []pb.ChatMssage
	ps    *PubSub
	sync.Mutex
}

// Push a message to this Endpoint
func (ep *EndPoint) Push(msg *pb.ChatMessage) {
	ep.Lock()
	defer ep.Unlock()
	if len(ep.inbox) > MaxQueueSize {
		ep.inbox = append(ep.inbox[1:], *msg)
	} else {
		ep.inbox = append(ep.inbox, *msg)
	}
}
