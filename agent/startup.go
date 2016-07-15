package main

import (
	sp "github.com/amorwilliams/bodoni/lib/services"
)

func startup() {
	go sigHandler()
	// init services discovery
	sp.Init("game", "snowflake")
}
