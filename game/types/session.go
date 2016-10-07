package types

const (
	SessKickOut = 0x1 // 踢掉
)

// 会话:
// 会话是一个单独玩家的上下文，在连入后到退出前的整个生命周期内存在
// 根据业务自行扩展上下文
type Session struct {
	// 会话标记
	Flag   int32
	UserId int32
}
