package main

import (
	"errors"

	. "agent/types"
	pb "lib/proto/game"

	log "github.com/Sirupsen/logrus"
)

var (
	ErrorStreamNotOpen = errors.New("stream not opened yet")
)

// forward messages to game server
func forward(sess *Session, p []byte) error {
	frame := &pb.Game_Frame{
		Type:    pb.Game_Message,
		Message: p,
	}

	// check stream
	if sess.Stream == nil {
		return ErrorStreamNotOpen
	}

	// forward the frame to game
	if err := sess.Stream.Send(frame); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
