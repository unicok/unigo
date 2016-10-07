package handler

import (
	tp "game/types"
	"lib/packet"
)

//----------------------------------- ping
func P_proto_ping_req(sess *tp.Session, reader *packet.Packet) []byte {
	tbl, _ := PKT_auto_id(reader)
	return packet.Pack(Code[E_proto_ping_ack], tbl, nil)
}
