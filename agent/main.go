package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/xtaci/kcp-go"

	"github.com/unicok/unigo/agent/utils"
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
	addr, err := net.ResloveTCPAddr("tcp4", port)
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
	l, err := kcp.Listen(port)
	checkError(err)

	log.Info("listening on:", l.Addr())

	// loop accepting
	for {
		conn, err := l.Accept()
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

func handleClient(conn net.Conn) {

}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
