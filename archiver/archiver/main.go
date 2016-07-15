package main

import _ "github.com/amorwilliams/bodoni/lib/statsd-pprof"

func main() {
	arch := &Archiver{}
	arch.init()
	<-arch.stop
}
