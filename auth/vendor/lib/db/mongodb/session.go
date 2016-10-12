package mongodb

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	DefaultMGOTimeout = 300
	DefaultConcurrent = 128
	// DefaultMongodbURL = "mongodb://172.17.0.1/mydb"
	// EnvMongodb        = "MONGODB_URL"
)

type session struct {
	*mgo.Session
}

type DialContext struct {
	sync.Mutex
	latch chan *session
}

func Dial(url string, concurrent int) (*DialContext, error) {
	return DialWithTimeout(url, concurrent, 10*time.Second, DefaultMGOTimeout*time.Second)
}

func DialWithTimeout(url string, concurrent int, dialTimeout time.Duration, timeout time.Duration) (*DialContext, error) {
	if concurrent <= 0 {
		concurrent = DefaultConcurrent
		log.Warnf("invalid concurrent, reset to %v", concurrent)
	}

	s, err := mgo.DialWithTimeout(url, dialTimeout)
	if err != nil {
		return nil, err
	}

	s.SetMode(mgo.Strong, true)
	s.SetSyncTimeout(timeout)
	s.SetSocketTimeout(timeout)
	s.SetCursorTimeout(0)

	c := &DialContext{}
	// create latch
	c.latch = make(chan *session, concurrent)
	for i := 0; i < concurrent; i++ {
		c.latch <- &session{s.Copy()}
	}

	return c, nil
}

func (c *DialContext) Close() {
	// c.Lock()
	// defer c.Unlock()
	for s := range c.latch {
		s.Close()
	}
}

// ExecuteFunc define
type ExecuteFunc func(s *session) error

// DBActionFunc define
type DBActionFunc func(c *mgo.Collection) error

// Query excute command
func (c *DialContext) Execute(f ExecuteFunc) error {
	// latch control
	s := <-c.latch
	defer func() {
		c.latch <- s
	}()
	s.Refresh()
	return f(s)
}

func (c *DialContext) DBAction(db, col string, f DBActionFunc) error {
	// latch control
	s := <-c.latch
	defer func() {
		c.latch <- s
	}()
	s.Refresh()
	cl := s.DB(db).C(col)
	return f(cl)
}

func (c *DialContext) EnsureCounter(db, col, id string) error {
	return c.Execute(func(s *session) error {
		err := s.DB(db).C(col).Insert(bson.M{
			"_id": id,
			"seq": 0,
		})
		if mgo.IsDup(err) {
			return nil
		}
		return err
	})
}

func (c *DialContext) NextSeq(db, col, id string) (int, error) {
	// result struct
	var res struct{ Seq int }
	err := c.Execute(func(s *session) error {
		_, err := s.DB(db).C(col).FindId(id).Apply(mgo.Change{
			Update:    bson.M{"$inc": bson.M{"seq": 1}},
			ReturnNew: true,
		}, &res)

		return err
	})

	return res.Seq, err
}

func (c *DialContext) EnsureIndex(db, col string, keys []string) error {
	return c.Execute(func(s *session) error {
		return s.DB(db).C(col).EnsureIndex(mgo.Index{
			Key:    keys,
			Unique: false,
			Sparse: true,
		})
	})
}

func (c *DialContext) EnsureUniqueIndex(db, col string, keys []string) error {
	return c.Execute(func(s *session) error {
		return s.DB(db).C(col).EnsureIndex(mgo.Index{
			Key:    keys,
			Unique: true,
			Sparse: true,
		})
	})
}
