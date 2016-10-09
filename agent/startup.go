package main

import sp "lib/services"

func startup() {

	go signalHandler()

	// init services discovery
	sp.Init("game", "snowflake")
}
