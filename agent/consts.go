package main

const (
	TcpReadDeadline    = 120   // 秒(没有网络包进入的最大间隔)
	SocketRcviveBuffer = 32767 // 每个连接的接收缓冲区
	SocketWriteBuffer  = 65535 // 每个连接的发送缓冲区

	PaddingLimit        = 8   // 小于此的返回包，加入填充
	PaddingSize         = 8   // 填充最大字节数
	PaddingUpdatePeriod = 300 // 填充字符更新周期

	MaxProtoNum   = 1000 // agent能处理的最大协议号
	DefaultMQSize = 512  // 默认玩家异步消息大小
	CustomTimer   = 60   // 玩家定时器间隔

	RPMLimit = 300 // 每分钟请求数控制，超过此值可以判定为DOS攻击
)
