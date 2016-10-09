package main

import (
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"

	_ "lib/logger"
	pb "lib/proto/chat"
	_ "lib/statsd-pprof"
)

const (
	port = ":50008"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Panic(err)
		os.Exit(-1)
	}
	log.Info("listening on:", lis.Addr())

	// 注册服务
	s := grpc.NewServer()
	ins := &server{}
	ins.init()
	pb.RegisterChatServiceServer(s, ins)

	// 开始服务
	s.Serve(lis)
}
