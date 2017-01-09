package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func initSignalHUP() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	go func(c chan os.Signal) {
		for _ = range c {
			logWriter.refresh()
			log.Println("reopened log file")
		}
	}(c)
}
