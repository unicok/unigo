package main

import (
	. "agent/types"
	"time"

	log "github.com/Sirupsen/logrus"
)

// 玩家1分钟定时器
func timer_work(sess *Session, out *Buffer) {
	// 发包频率控制，太高的RPS直接踢掉
	interval := time.Now().Sub(sess.ConnectTime).Minutes()
	if interval > 1 { // 登录时长超过1分钟才开始统计rpm。防脉冲
		rpm := float64(sess.PacketCount) / interval
		if rpm > RPMLimit {
			sess.Flag |= SessKickOut
			log.Error("玩家RPM太高 RPM:", rpm)
			return
		}
	}
}
