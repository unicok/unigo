package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"

	"github.com/unicok/unigo/snowflake/etcdclient"
	pb "github.com/unicok/unigo/snowflake/proto"
)

const (
	service      = "[SNOWFLAKE]"
	envMachineID = "MACHINE_ID" // specific machine id
	path         = "/seqs/"
	uuidKey      = "/seqs/snowflake-uuid"
	backoff      = 100  // max backoff delay millisecond
	concurrent   = 128  // max concurrent connections to etcd
	uuidQueue    = 1024 // uuid process queue
)

const (
	tsMask        = 0x1FFFFFFFFFF // 41bit
	snMask        = 0xFFF         // 12bit
	machineIDMask = 0x3FF         // 10bit
)

type server struct {
	machineID  uint64 // 10-bit machine append
	clientPool chan etcd.KeysAPI
	procCh     chan chan uint64
}

func (p *server) init() {
	p.clientPool = make(chan etcd.KeysAPI, concurrent)
	p.procCh = make(chan chan uint64, uuidQueue)

	// init client pool
	for i := 0; i < concurrent; i++ {
		p.clientPool <- etcdclient.KeysAPI()
	}

	// check if user specified machine id is set
	if env := os.Getenv(envMachineID); env != "" {
		id, err := strconv.Atoi(env)
		if err != nil {
			log.Panic(err)
			os.Exit(-1)
		}

		p.machineID = (uint64(id) & machineIDMask) << 12
		log.Info("machine id specified:", id)
	} else {
		p.initMachineID()
	}

	go p.uuidTask()
}

func (p *server) initMachineID() {
	client := <-p.clientPool
	defer func() { p.clientPool <- client }()

	for {
		// get the key
		resp, err := client.Get(context.Background(), uuidKey, nil)
		if err != nil {
			log.Panic(err)
			os.Exit(-1)
		}

		// get preValue & preIndex
		prevValue, err := strconv.Atoi(resp.Node.Value)
		if err != nil {
			log.Panic(err)
			os.Exit(-1)
		}
		prevIndex := resp.Node.ModifiedIndex

		// compareAndSwap
		resp, err = client.Set(context.Background(), uuidKey, fmt.Sprint(prevValue+1), &etcd.SetOptions{PrevIndex: prevIndex})
		if err != nil {
			casDelay()
			continue
		}

		// record serial number of this service, already shifted
		p.machineID = (uint64(prevValue+1) & machineIDMask) << 12
		return
	}
}

// Next is get next value of a key, like auto-incrememt in mysql
func (p *server) Next(ctx context.Context, in *pb.Snowflake_Key) (*pb.Snowflake_Value, error) {
	client := <-p.clientPool
	defer func() { p.clientPool <- client }()
	key := path + in.Name
	for {
		// get the key
		resp, err := client.Get(context.Background(), key, nil)
		if err != nil {
			log.Error(err)
			return nil, errors.New("Key not exists, need to create first")
		}

		// get prevValue & prevIndex
		prevValue, err := strconv.Atoi(resp.Node.Value)
		if err != nil {
			log.Error(err)
			return nil, errors.New("marlformed value")
		}
		prevIndex := resp.Node.ModifiedIndex

		// compareAndSwap
		resp, err = client.Set(context.Background(), key, fmt.Sprint(prevValue+1), &etcd.SetOptions{PrevIndex: prevIndex})
		if err != nil {
			casDelay()
			continue
		}
		return &pb.Snowflake_Value{Value: int64(prevValue + 1)}, nil
	}
}

// GetUUID is generate an unique uuid
func (p *server) GetUUID(context.Context, *pb.Snowflake_NullRequest) (*pb.Snowflake_UUID, error) {
	req := make(chan uint64, 1)
	p.procCh <- req
	return &pb.Snowflake_UUID{Uuid: <-req}, nil
}

// uuid generator
func (p *server) uuidTask() {
	var sn uint64    // 12-bit serial no
	var lastts int64 // last timestamp
	for {
		ret := <-p.procCh
		// get a correct serial number
		t := ts()
		// clock shift backward
		if t < lastts {
			log.Error("clock shift happened, waiting until the clock moving to the next millisecond.")
			t = p.waitMs(lastts)
		}

		// same millisecond
		if lastts == t {
			sn = (sn + 1) & snMask
			// serial number overflows, wait until next ms
			if sn == 0 {
				t = p.waitMs(lastts)
			}
		} else { // new millsecond, reset serial number to 0
			sn = 0
		}
		// remember last timestamp
		lastts = t

		// generate uuid, format:
		//
		// 0		0.................0		0..............0	0........0
		// 1-bit	41bit timestamp			10bit machine-id	12bit sn
		var uuid uint64
		uuid |= (uint64(t) & tsMask) << 22
		uuid |= p.machineID
		uuid |= sn
		ret <- uuid
	}
}

// waitMs will spin wait till next millisecond
func (p *server) waitMs(lastts int64) int64 {
	t := ts()
	for t <= lastts {
		t = ts()
	}
	return t
}

// random delay
func casDelay() {
	<-time.After(time.Duration(rand.Int63n(backoff)) * time.Millisecond)
}

// get timestamp
func ts() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
