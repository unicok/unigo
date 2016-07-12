package nsqredo

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/mgo.v2/bson"
)

const (
	nsqdEnv       = "NSQD_HOST"
	defaultPubURL = "http://172.17.42.1:4151/pub?topic=REDOLOG"
	mine          = "application/octet-strem"
)

//Change is a data change
type Change struct {
	Collection string
	Field      string
	Doc        interface{}
}

//RedoRecord is a redo record represents complete transaction
type RedoRecord struct {
	API     string
	UID     int32
	TS      uint64
	Changes []Change
}

//NewRedoRecord is return a new RedoRecord
func NewRedoRecord(uid int32, api string, ts uint64) *RedoRecord {
	return &RedoRecord{UID: uid, API: api, TS: ts}
}

//AddChange is add a change with o(old value) and n(new value)
func (r *RedoRecord) AddChange(collection string, field string, doc interface{}) {
	r.Changes = append(r.Changes, Change{Collection: collection, Field: field, Doc: doc})
}

var (
	pubAddr string
	prefix  string
	ch      chan []byte
)

func init() {
	pubAddr = defaultPubURL
	if env := os.Getenv(nsqdEnv); env != "" {
		pubAddr = env + "/pub?topic=REDOLOG"
	}
	ch = make(chan []byte, 4096)
	go publishTask()
}

func publishTask() {
	for {
		//post to nsqd
		bts := <-ch
		resp, err := http.Post(pubAddr, mine, bytes.NewReader(bts))
		if err != nil {
			log.Println(err)
			continue
		}

		//read response
		if _, err := ioutil.ReadAll(resp.Body); err != nil {
			log.Println(err)
		}

		//close
		resp.Body.Close()
	}
}

// Publish to nsqd (localhost nsqd is suggested!)
func Publish(r *RedoRecord) {
	//pack message
	if bts, err := bson.Marshal(r); err == nil {
		ch <- bts
	} else {
		log.Println(err, r)
		return
	}
}
