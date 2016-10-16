package main

import (
	"net"
	"os"

	"google.golang.org/grpc"

	_ "lib/logger"
	pb "lib/proto/snowflake"
	_ "lib/statsd-pprof"

	log "github.com/Sirupsen/logrus"
)

const (
	port = ":50003"
)

func main() {
	// listen
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}
	log.Info("listening on ", l.Addr())

	// register service
	s := grpc.NewServer()
	ins := &server{}
	ins.init()
	pb.RegisterSnowflakeServiceServer(s, ins)

	// start service
	s.Serve(l)
}
