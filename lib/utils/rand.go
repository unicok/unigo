package utils

import "time"

const (
	Prerng = 1024
)

var (
	x0  = uint32(time.Now().UnixNano())
	a   = uint32(1664525)
	c   = uint32(1013904223)
	LCG chan uint32
)

func init() {
	LCG = make(chan uint32, Prerng)
	go func() {
		for {
			x0 = a*x0 + c
			LCG <- x0
		}
	}()
}
