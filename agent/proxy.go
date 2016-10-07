package main

import (
	"agent/handler"
	. "agent/types"
	"fmt"
	"lib/packet"
	"lib/utils"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/peterbourgon/g2s"
)

const (
	StatsdPrefix      = "API."
	EnvStatsd         = "STATSD_HOST"
	DefaultStatsdHost = "172.17.42.1:8125"
)

var _statter g2s.Statter

func init() {
	addr := DefaultStatsdHost
	if env := os.Getenv(EnvStatsd); env != "" {
		addr = env
	}

	s, err := g2s.Dial("udp", addr)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
	_statter = s
}

// client protocol handle proxy
func proxyUserRequest(sess *Session, p []byte) []byte {
	start := time.Now()
	defer utils.PrintPanicStack()

	//解密
	if sess.Flag&SessEncrypt != 0 {
		sess.Decoder.XORKeyStream(p, p)
	}

	//封装为reader
	reader := packet.Reader(p)

	// 读客户端数据包序列号(1,2,3...)
	// 客户端发送的数据包必须包含一个自增的序号，必须严格递增
	// 加密后，可避免重放攻击-REPLAY-ATTACK
	seq_id, err := reader.ReadU32()
	if err != nil {
		log.Error("read client timestamp failed:", err)
		sess.Flag |= SessKickOut
		return nil
	}

	// 数据包序列号验证
	if seq_id != sess.PacketCount {
		log.Errorf("illegal packet sequence id:%v should be:%v size:%v", seq_id, sess.PacketCount, len(p)-6)
		sess.Flag |= SessKickOut
		return nil
	}

	// 读协议号
	b, err := reader.ReadS16()
	if err != nil {
		log.Error("read protocol number failed.")
		sess.Flag |= SessKickOut
		return nil
	}

	// 根据协议号断做服务划分
	// 协议号的划分采用分割协议区间, 用户可以自定义多个区间，用于转发到不同的后端服务
	var ret []byte
	if b > MaxProtoNum {
		if err := forward(sess, p[4:]); err != nil {
			log.Errorf("service id:%v execute failed, error:%v", b, err)
			sess.Flag |= SessKickOut
			return nil
		}
	} else {
		if h := handler.Handlers[b]; h != nil {
			ret = h(sess, reader)
		} else {
			log.Errorf("service id:%v not bind", b)
			sess.Flag |= SessKickOut
			return nil
		}
	}

	// 监控协议处理时间
	// 监控数值会发送到statsd,格式为:
	// API.XXX_REQ = 10ms
	elasped := time.Now().Sub(start)
	if b != 0 { // 排除心跳包日志
		log.Debug("[REQ]", handler.RCode[b])
		_statter.Timing(1.0, fmt.Sprintf("%v%v", StatsdPrefix, handler.RCode[b]), elasped)
	}

	return ret
}
