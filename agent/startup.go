package main

import (
	sp "github.com/unicok/unigo/lib/services"
)

func startup() {
	go sigHandler()
	// init services discovery
	sp.Init("game", "snowflake")
}
