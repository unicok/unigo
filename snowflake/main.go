package main

import (
	"net"
	"os"

	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
	_ "github.com/amorwilliams/bodoni/lib/statsd-pprof"
	"github.com/amorwilliams/bodoni/snowflake/proto"
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
	proto.RegisterSnowflakeServiceServer(s, ins)

	// start service
	s.Serve(l)
}
