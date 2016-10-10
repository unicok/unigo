package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	pb "lib/proto/chat"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"golang.org/x/net/context"
	"gopkg.in/vmihailenco/msgpack.v2"
)

const (
	SERVICE = "[CHAT]"
)

const (
	envBoltDBFile     = "BOLT_DB_FILE"
	defaultBoltDBFile = "/data/CHAT.DAT"
	BoltDBBucket      = "EPS"
	MaxQueueSize      = 128 // num of message kept
	PendingSize       = 65536
	CheckInterval     = time.Minute
)

var (
	OK                 = &pb.Chat_Nil{}
	ErrorAlreadyExists = errors.New("id already exists")
	ErrorNotExists     = errors.New("id not exists")
	BoltDBFile         = defaultBoltDBFile
)

func init() {
	// check if user specified machine id is set
	if env := os.Getenv(envBoltDBFile); env != "" {
		BoltDBFile = env
		log.Info("bolt db file specified:", env)
	}
}

// Endpoint definition
type EndPoint struct {
	inbox []pb.Chat_Message
	ps    *PubSub
	sync.Mutex
}

// NewEndPoint create a new endpoint
func NewEndPoint() *EndPoint {
	u := &EndPoint{}
	u.ps = &PubSub{}
	u.ps.init()
	return u
}

// Push a message to this Endpoint
func (ep *EndPoint) Push(msg *pb.Chat_Message) {
	ep.Lock()
	defer ep.Unlock()
	if len(ep.inbox) > MaxQueueSize {
		ep.inbox = append(ep.inbox[1:], *msg)
	} else {
		ep.inbox = append(ep.inbox, *msg)
	}
}

// Read all messages from this Endpoint
func (ep *EndPoint) Read() []pb.Chat_Message {
	ep.Lock()
	defer ep.Unlock()
	return append([]pb.Chat_Message(nil), ep.inbox...)
}

// server definition
type server struct {
	eps     map[uint64]*EndPoint
	pending chan uint64 //dirty id pending
	sync.RWMutex
}

func (s *server) init() {
	s.eps = make(map[uint64]*EndPoint)
	s.pending = make(chan uint64, PendingSize)
	s.restore()
	go s.persistenceTask()
}

func (s *server) readEP(id uint64) *EndPoint {
	s.RLock()
	defer s.RUnlock()
	return s.eps[id]
}

func (s *server) Subscribe(p *pb.Chat_Id, stream pb.ChatService_SubscribeServer) error {
	// read endpoint
	ep := s.readEP(p.Id)
	if ep == nil {
		log.Errorf("cannot find endpoint %v when Subscribe", p.Id)
		return ErrorNotExists
	}

	// send history chat messages
	msgs := ep.Read()
	for k := range msgs {
		if err := stream.Send(&msgs[k]); err != nil {
			return nil
		}
	}

	// create subsciber
	e := make(chan error, 1)
	var once sync.Once
	f := NewSubscriber(func(msg *pb.Chat_Message) {
		if err := stream.Send(msg); err != nil {
			once.Do(func() { // protect for channel blocking
				e <- err
			})
		}
	})

	// subscibe to the endpoint
	log.Debugf("subscribe to :%v", p.Id)
	ep.ps.Sub(f)
	defer func() {
		ep.ps.Unsub(f)
		log.Debugf("unsubscribe from :%v", p.Id)
	}()

	// client send cancel to stop receiving, see service_test.go for example
	select {
	case <-stream.Context().Done():
	case <-e:
		log.Error(e)
	}
	return nil
}

func (s *server) Send(ctx context.Context, msg *pb.Chat_Message) (*pb.Chat_Nil, error) {
	ep := s.readEP(msg.Id)
	if ep == nil {
		log.Errorf("cannot find endpoint %v when Send", msg.Id)
		return nil, ErrorNotExists
	}

	ep.ps.Pub(msg)
	ep.Push(msg)
	s.pending <- msg.Id
	return OK, nil
}

func (s *server) Reg(ctx context.Context, p *pb.Chat_Id) (*pb.Chat_Nil, error) {
	s.Lock()
	defer s.Unlock()
	ep := s.eps[p.Id]
	if ep != nil {
		log.Errorf("id already exists:%v when Reg", p.Id)
		return nil, ErrorAlreadyExists
	}

	s.eps[p.Id] = NewEndPoint()
	log.Debug("eps size:", len(s.eps))
	s.pending <- p.Id
	return OK, nil
}

// persistenceTask persistence endpoints into db
func (s *server) persistenceTask() {
	timer := time.After(CheckInterval)
	db := s.openDB()
	changes := make(map[uint64]bool)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case key := <-s.pending:
			changes[key] = true
		case <-timer:
			s.dump(db, changes)
			if len(changes) > 0 {
				log.Infof("perisisted %v endpoints:", len(changes))
			}
			changes = make(map[uint64]bool)
			timer = time.After(CheckInterval)
		case nr := <-sig:
			s.dump(db, changes)
			db.Close()
			log.Info(nr)
			os.Exit(0)
		}
	}
}

func (s *server) openDB() *bolt.DB {
	db, err := bolt.Open(BoltDBFile, 0600, nil)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}
	// create bulket
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BoltDBBucket))
		if err != nil {
			log.Panicf("create bucket: %s", err)
			os.Exit(-1)
		}
		return nil
	})
	return db
}

func (s *server) dump(db *bolt.DB, changes map[uint64]bool) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BoltDBBucket))
		for k := range changes {
			ep := s.readEP(k)
			if ep == nil {
				log.Errorf("cannot find endpoint %v when dump", k)
				continue
			}

			// serialization and save
			bin, err := msgpack.Marshal(ep.Read())
			if err != nil {
				log.Error("cannot marshal:", err)
				continue
			}

			err = b.Put([]byte(fmt.Sprint(k)), bin)
			if err != nil {
				log.Error(err)
				continue
			}
		}
		return nil
	})
}

func (s *server) restore() {
	// restore data from db file
	db := s.openDB()
	defer db.Close()
	count := 0
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BoltDBBucket))
		b.ForEach(func(k, v []byte) error {
			var msg []pb.Chat_Message
			err := msgpack.Unmarshal(v, &msg)
			if err != nil {
				log.Error("unmarshal chat msg corrupted:", err)
				os.Exit(-1)
			}
			id, err := strconv.ParseUint(string(k), 0, 64)
			if err != nil {
				log.Error("conv chat id corrupted:", err)
				os.Exit(-1)
			}
			ep := NewEndPoint()
			ep.inbox = msg
			s.eps[id] = ep
			count++
			return nil
		})
		return nil
	})

	log.Infof("restored %v chats", count)
}
