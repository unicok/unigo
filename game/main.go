package main

import (
	"net"
	"os"

	"google.golang.org/grpc"

	_ "lib/logger"

	pbgame "lib/proto/game"
	sp "lib/services"

	log "github.com/Sirupsen/logrus"
)

const (
	_port = ":51000"
)

func main() {
	lis, err := net.Listen("tcp", _port)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}
	log.Info("listening on ", lis.Addr())

	// registry service
	s := grpc.NewServer()
	ins := new(server)
	pbgame.RegisterGameServiceServer(s, ins)

	sp.Init("snowflake")
	s.Serve(lis)
}
