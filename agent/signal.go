package main

import (
	"lib/utils"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

var (
	wg  sync.WaitGroup
	die = make(chan struct{})
)

func signalHandler() {
	defer utils.PrintPanicStack()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)

	for {
		msg := <-ch
		switch msg {
		case syscall.SIGTERM:
			close(die)
			log.Info("sigterm received")
			log.Info("waiting for agents close, please wait...")
			wg.Wait()
			log.Info("agent shutdown.")
			os.Exit(0)
		}
	}
}
