package main

import (
	"errors"
	"io"
	"strconv"

	"game/handler"
	tp "game/types"

	"lib/packet"
	pb "lib/proto"
	"lib/registry"
	"lib/utils"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

const (
	Service = "[GAME]"
)

var (
	ErrorIncorrectFrameType = errors.New("incorrect frame type")
	ErrorServiceNotBind     = errors.New("service not bind")
)

type server struct{}

// PIPELINE #1 stream receiver
// this function is to make the stream receiving SELECTABLE
func (s *server) recv(stream pb.GameService_StreamServer, sess_die chan struct{}) chan *pb.Game_Frame {
	ch := make(chan *pb.Game_Frame, 1)
	go func() {
		defer func() {
			close(ch)
		}()
		for {
			in, err := stream.Recv()
			if err == io.EOF { // client close
				return
			}
			if err != nil {
				log.Error(err)
				return
			}
			select {
			case ch <- in:
			case <-sess_die:
			}
		}
	}()
	return ch
}

// PIPELINE #2 stream processing
// the center of game logic
func (s *server) Stream(stream pb.GameService_StreamServer) error {
	defer utils.PrintPanicStack()

	// session init
	var sess tp.Session
	sess_die := make(chan struct{})
	ch_agent := s.recv(stream, sess_die)
	ch_ipc := make(chan *pb.Game_Frame, DefaultChIPCSize)

	defer func() {
		registry.Unregister(sess.UserId)
		close(sess_die)
		log.Debug("stream end:", sess.UserId)
	}()

	// read metadata from context
	md, ok := metadata.FromContext(stream.Context())
	if !ok {
		log.Error("cannot read metadata from context")
		return ErrorIncorrectFrameType
	}
	// read key
	if len(md["userid"]) == 0 {
		log.Error("cannot read key:userid from metadata")
		return ErrorIncorrectFrameType
	}
	// parse userid
	userId, err := strconv.Atoi(md["userid"][0])
	if err != nil {
		log.Error(err)
		return ErrorIncorrectFrameType
	}

	// register user
	sess.UserId = int32(userId)
	registry.Register(sess.UserId, ch_ipc)
	log.Debug("userid", sess.UserId)

	// >> main message loop <<
	for {
		select {
		case frame, ok := <-ch_agent: // frames from agent
			if !ok { // EOF
				return nil
			}
			switch frame.Type {
			case pb.Game_Message: // the passthrough message from client->agent->game
				// locate handler by proto number
				reader := packet.Reader(frame.Message)
				c, err := reader.ReadS16()
				if err != nil {
					log.Error(err)
					return err
				}
				h := handler.Handlers[c]
				if h == nil {
					log.Error("service not bind:", c)
					return ErrorServiceNotBind
				}

				// handle request
				ret := h(&sess, reader)

				// construct frame & return message from logic
				if err != nil {
					if err := stream.Send(&pb.Game_Frame{Type: pb.Game_Message, Message: ret}); err != nil {
						log.Error(err)
						return err
					}
				}

				// session control by logic
				if sess.Flag&tp.SessKickOut != 0 { //logic kick out
					if err := stream.Send(&pb.Game_Frame{Type: pb.Game_Kick}); err != nil {
						log.Error(err)
						return err
					}
					return nil
				}
			case pb.Game_Ping:
				if err := stream.Send(&pb.Game_Frame{Type: pb.Game_Ping, Message: frame.Message}); err != nil {
					log.Error(err)
					return err
				}
				log.Debug("pinged")
			default:
				log.Error("incorrect frame type:", frame.Type)
				return ErrorIncorrectFrameType
			}
		case frame := <-ch_ipc: // forward async messages from interprocess(goroutines) communication
			if err := stream.Send(frame); err != nil {
				log.Error(err)
				return err
			}
		}
	}
}
