package mgo

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"labix.org/v2/mgo"
)

const (
	DefaultMGOTimeout = 300
	DefaultConcurrent = 128
	DefaultMongodbURL = "mongodb://172.17.0.1/mydb"
	EnvMongodb        = "MONGODB_URL"
)

// Database define
type Database struct {
	session *mgo.Session
	latch   chan *mgo.Session
}

// ExecuteFunc define
type ExecuteFunc func(sess *mgo.Session) error

// Init Database
func (db *Database) Init() {
	// create latch
	db.latch = make(chan *mgo.Session, DefaultConcurrent)
	// connect db
	mongodbURL := DefaultMongodbURL
	if env := os.Getenv(EnvMongodb); env != "" {
		mongodbURL = env
	}
	sess, err := mgo.Dial(mongodbURL)
	if err != nil {
		log.Fatal("mongodb: cannot connect to", mongodbURL, err)
		os.Exit(-1)
	}

	// set params
	sess.SetMode(mgo.Strong, true)
	sess.SetSocketTimeout(DefaultMGOTimeout * time.Second)
	sess.SetCursorTimeout(0)
	db.session = sess

	for k := 0; k < cap(db.latch); k++ {
		db.latch <- sess.Copy()
	}
}

// Execute run
func (db *Database) Execute(f ExecuteFunc) error {
	// latch control
	sess := <-db.latch
	defer func() {
		db.latch <- sess
	}()
	sess.Refresh()
	return f(sess)
}
