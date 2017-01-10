package main

import (
	"log"
	"time"
)

// XenPoller struct for manager of polling a host for data
type XenPoller struct {
	server *XenServer
	doneCh chan bool
}

// NewXenPoller instantiates and launches goroutine for a XenPoller
func NewXenPoller(server *XenServer) *XenPoller {
	p := &XenPoller{}
	p.server = server
	p.doneCh = make(chan bool)
	go p.loop()
	return p
}

func (p *XenPoller) loop() {
	defer log.Printf("returning from polling loop for %s", p.server.Hostname)
	for {
		select {
		case <-p.doneCh:
			return
		default:
			alarm := time.After(time.Duration(1) * time.Second)
			log.Printf("starting polling loop for %s", p.server.Hostname)
			<-alarm
		}
	}
}
