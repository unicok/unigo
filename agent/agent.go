package main

import (
	"time"

	. "agent/types"
	pb "lib/proto"
	"lib/utils"
)

// PIPELINE #2: agent
// all the packets from handleClient() will be handled
func agent(sess *Session, in chan []byte, out *Buffer) {
	defer wg.Done() // will decrease waitgroup by one, useful for manual server shutdown
	defer utils.PrintPanicStack()

	// init Session
	sess.MQ = make(chan pb.Game_Frame, DefaultMQSize)
	sess.ConnectTime = time.Now()
	sess.LastPacketTime = time.Now()
	// minute timer
	min_timer := time.After(time.Minute)

	// cleanup work
	defer func() {
		close(sess.Die)
		if sess.Stream != nil {
			sess.Stream.CloseSend()
		}
	}()

	// >> the main message loop <<
	// handles 4 types of message:
	//  1. from client
	//  2. from game service
	//  3. timer
	//  4. server shutdown signal
	for {
		select {
		case msg, ok := <-in: // packet from network
			if !ok {
				return
			}
			sess.PacketCount++
			sess.PacketTime = time.Now()
			if result := proxyUserRequest(sess, msg); result != nil {
				out.send(sess, result)
			}
		case frame := <-sess.MQ: // packets from game
			switch frame.Type {
			case pb.Game_Message:
				out.send(sess, frame.Message)
			case pb.Game_Kick:
				sess.Flag |= SessKickOut
			}
		case <-min_timer: // minutes timer
			timer_work(sess, out)
			min_timer = time.After(time.Minute)
		case <-die: // server is shuting down...
			sess.Flag |= SessKickOut
		}

		// see if the player should be kicked out
		if sess.Flag&SessKickOut != 0 {
			return
		}
	}
}
