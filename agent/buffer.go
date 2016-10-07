package main

import (
	"encoding/binary"
	"net"
	"time"

	. "agent/types"
	"lib/packet"
	"lib/utils"

	log "github.com/Sirupsen/logrus"
)

// Buffer is PIPELINE #3: buffer
// controls the packet sending for the client
type Buffer struct {
	ctrl    chan struct{} // recevice exit signal
	pending chan []byte   // pending packets
	conn    net.Conn      // connection
	cache   []byte        // for combined syscall write
}

var (
	// for padding packet, random content
	// add some random content to confuse packet decrypter
	_padding [PaddingSize]byte
)

func init() {
	go func() {
		for {
			for k := range _padding {
				_padding[k] = byte(<-utils.LCG)
			}
			log.Info("Padding Updated:", _padding)
			<-time.After(PaddingUpdatePeriod * time.Second)
		}
	}()
}

// packet sending procedure
func (b *Buffer) send(sess *Session, data []byte) {
	// in case of empty packet
	if data == nil {
		return
	}

	// padding
	// if the size of the data to return is tiny, pad with some random numbers
	// this strategy may change to randomize padding
	if len(data) < PaddingLimit {
		data = append(data, _padding[:]...)
	}

	// encryption
	// (NOT_ENCRYPTED) -> KEYEXCG -> ENCRYPT
	if sess.Flag&SessEncrypt != 0 { // encryption is enabled
		sess.Encoder.XORKeyStream(data, data)
	} else if sess.Flag&SessKeyDone != 0 { // key is exchanged, encryption is not yet enabled
		sess.Flag &^= SessKeyDone
		sess.Flag |= SessEncrypt
	}

	// queue the data for sending
	b.pending <- data
	return
}

// packet sending goroutine
func (b *Buffer) start() {
	defer utils.PrintPanicStack()
	for {
		select {
		case data := <-b.pending:
			b.rawSend(data)
		case <-b.ctrl: // receive session end signal
			close(b.pending)
			// close the connection
			b.conn.Close()
			return
		}
	}
}

// raw packet encapsulation and put it online
func (b *Buffer) rawSend(data []byte) bool {
	// combine output to reduce syscall.write
	sz := len(data)
	binary.BigEndian.PutUint16(b.cache, uint16(sz))
	copy(b.cache[2:], data)

	// wiret data
	n, err := b.conn.Write(b.cache[:sz+2])
	if err != nil {
		log.Warningf("Error send reply data, byte: %v reason %v", n, err)
		return false
	}

	return true
}

// create a associated write buffer for a session
func newBuffer(conn net.Conn, ctrl chan struct{}) *Buffer {
	buf := &Buffer{conn: conn}
	buf.pending = make(chan []byte)
	buf.ctrl = ctrl
	buf.cache = make([]byte, packet.PacketLimit+2)
	return buf
}
