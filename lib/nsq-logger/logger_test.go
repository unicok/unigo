package nsqlogger

import "testing"

func TestLogger(t *testing.T) {
	SetPrefix("[TEST]")
	SetLogLevel(TRACE)
	Finest("finest")
	Fine("fine")
	Debug("debug")
	Trace("trace")
	Info("info")
	Warn("warn")
	Error("error")
	Critical("critical")
	Flush()
}

func BenchmarkLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Debug("benchmark")
	}
}
