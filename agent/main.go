package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	log "github.com/Sirupsen/logrus"

	"agent/utils"
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

}

func udpServer() {

}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
