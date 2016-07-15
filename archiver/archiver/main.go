package main

import _ "github.com/unicok/unigo/lib/statsd-pprof"

func main() {
	arch := &Archiver{}
	arch.init()
	<-arch.stop
}
