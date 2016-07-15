package main

import (
	"encoding/binary"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	nsq "github.com/bitly/go-nsq"
	"github.com/boltdb/bolt"
)

const (
	defaultNsqlookupd  = "http://172.17.42.1:4161"
	envNsqlookupd      = "NSQLOOKUPD_HOST"
	topic              = "REDOLOG"
	channel            = "ARCH"
	service            = "[ARCH]"
	redoTimeFormat     = "REDO-2006-01-02.rdo"
	redoRotateInterval = 24 * time.Hour
	boltDBBucket       = "REDOLOG"
	dataPath           = "/data/"
	envDataPath        = "DATA_PATH"
	batchSize          = 1024
	syncInterval       = 10 * time.Millisecond
)

// Archiver is a struct
type Archiver struct {
	pending chan []byte
	stop    chan bool
}

func (p *Archiver) init() {
	p.pending = make(chan []byte, batchSize)
	p.stop = make(chan bool)
	cfg := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, channel, cfg)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}

	// message process
	consumer.AddHandler(nsq.HandlerFunc(func(msg *nsq.Message) error {
		p.pending <- msg.Body
		return nil
	}))

	// read enviroment varialbe
	addrs := []string{defaultNsqlookupd}
	if env := os.Getenv(envNsqlookupd); env != "" {
		addrs = strings.Split(env, ";")
	}

	// connect to nsqlookupd
	log.Debug("connect to nsqlookupd ip:", addrs)
	if err := consumer.ConnectToNSQLookupds(addrs); err != nil {
		log.Error(err)
		return
	}
	log.Info("nsqlookupd connected")

	go p.archiveTask()
}

func (p *Archiver) archiveTask() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	timer := time.After(redoRotateInterval)
	syncTicker := time.NewTicker(syncInterval)
	db := p.newRedolog()
	key := make([]byte, 8)

	for {
		select {
		case <-syncTicker.C:
			n := len(p.pending)
			if n == 0 {
				continue
			}

			// put log to bucket
			db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(boltDBBucket))
				for i := 0; i < n; i++ {
					id, err := b.NextSequence()
					if err != nil {
						log.Error(err)
						continue
					}
					binary.BigEndian.PutUint64(key, uint64(id))
					if err = b.Put(key, <-p.pending); err != nil {
						log.Error(err)
						continue
					}
				}
				return nil
			})
		case <-timer:
			db.Close()
			// rotate redolog
			db = p.newRedolog()
			timer = time.After(redoRotateInterval)
		case <-sig:
			db.Close()
			log.Info("SIGTERM")
			os.Exit(0)
		}
	}
}

func (p *Archiver) newRedolog() *bolt.DB {
	// read enviroment varialbe
	path := dataPath
	if env := os.Getenv(envDataPath); env != "" {
		if env == "." || env == "./" {
			path = ""
		} else {
			if strings.HasSuffix(path, "/") {
				path = env
			} else {
				path = env + "/"
			}
		}
	}

	filename := path + time.Now().Format(redoTimeFormat)
	log.Info(filename)
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}

	// create bulket
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(boltDBBucket))
		if err != nil {
			log.Errorf("create bucket: %s", err)
			return err
		}
		return nil
	})

	return db
}
