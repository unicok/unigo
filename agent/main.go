package main

import (
	"encoding/binary"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/xtaci/kcp-go"

	. "agent/types"
	"lib/utils"
)

const (
	port = ":8888" // the incoming address for this agent, you can use docker -p to map port
)

const (
	Service = "[AGENT]"
)

func main() {
	// to catch all uncaught panic
	defer utils.PrintPanicStack()

	// open profiling
	go func() {
		log.Info(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	// startup
	startup()

	go tcpServer()
	go udpServer()

	//wait forever
	select {}
}

func tcpServer() {
	// resolve address & start listening
	addr, err := net.ResolveTCPAddr("tcp4", port)
	checkError(err)

	l, err := net.ListenTCP("tcp", addr)
	checkError(err)

	log.Info("listening on:", l.Addr())

	// loop accepting
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Warning("accept failed:", err)
			continue
		}
		// set socket read buffer
		conn.SetReadBuffer(SocketRcviveBuffer)
		// set socket write buffer
		conn.SetWriteBuffer(SocketWriteBuffer)
		// start a goroutine for every incoming connection for reading
		go handleClient(conn)

		// check server close signal
		select {
		case <-die:
			l.Close()
			return
		default:
		}
	}
}

func udpServer() {
	l, err := kcp.ListenWithOptions(port, nil, 0, 0)
	checkError(err)

	log.Info("listening on:", l.Addr())

	// loop accepting
	for {
		conn, err := l.AcceptKCP()
		if err != nil {
			log.Warning("accept failed:", err)
			continue
		}
		// set kcp params
		conn.SetNoDelay(1, 30, 2, 1)
		// start a goroutine for every incoming connection for reading
		go handleClient(conn)

		// check server close signal
		select {
		case <-die:
			l.Close()
			return
		default:
		}
	}
}

// PIPELINE #1: handleClient
// the goroutine is used for reading incoming PACKETS
// each packet is defined as :
// | 2B size |     DATA       |
//
func handleClient(conn net.Conn) {
	defer utils.PrintPanicStack()
	// for reading the 2-byte header
	header := make([]byte, 2)
	// the input channel for agent()
	in := make(chan []byte)
	defer func() {
		close(in)
	}()

	// create a new session object for the connection
	// and record it's IP address
	var sess = &Session{}
	host, port, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Error("cannot get remote address:", err)
		return
	}
	log.Infof("new connection from:%v port:%v", host, port)
	sess.IP = net.ParseIP(host)

	// session die signal, will be triggered  by agent()
	sess.Die = make(chan struct{})

	//create a write buffer
	out := newBuffer(conn, sess.Die)
	go out.start()

	// start agent for PACKET processing
	wg.Add(1)
	go agent(sess, in, out)

	// read loop
	for {
		// solve dead link problem:
		// physical disconnection without any communcation between client and server
		// will cause the read to block FOREVER, so a timeout is a rescue.
		conn.SetReadDeadline(time.Now().Add(TcpReadDeadline * time.Second))

		// read 2b header
		n, err := io.ReadFull(conn, header)
		if err != nil {
			log.Warningf("read header failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}
		size := binary.BigEndian.Uint16(header)

		// alloc a byte slice of the size defined in the header for reading DATA
		payload := make([]byte, size)
		n, err = io.ReadFull(conn, payload)
		if err != nil {
			log.Warningf("read payload failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}

		// deliver the data to the input queue of agent()
		select {
		case in <- payload: // payload queued
		case <-sess.Die:
			log.Warningf("connection closed by logic, flag:%v ip:%v", sess.Flag, sess.IP)
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
