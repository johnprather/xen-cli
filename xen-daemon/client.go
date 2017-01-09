package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

// ClientID is a numeric identifier for connected clients
type ClientID uint64

// Client is a connected client
type Client struct {
	conn   net.Conn
	server *Server
	ID     ClientID
	sendCh chan string
	doneCh chan bool
}

// NewClient returns an instantiated Client object
func NewClient(server *Server, conn net.Conn) *Client {
	client := &Client{}
	client.conn = conn
	client.server = server
	client.ID = server.nextClientID
	server.nextClientID++
	client.sendCh = make(chan string)
	client.doneCh = make(chan bool)
	return client
}

func (c *Client) listen() {
	go c.listenRead()
	c.listenWrite()
}

func (c *Client) listenRead() {

	defer c.conn.Close()
	reader := bufio.NewReader(c.conn)
	for {
		select {
		case <-c.doneCh:
			c.server.delClient(c)
			c.doneCh <- true
			return
		default:
			data, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					c.doneCh <- true
				} else {
					log.Printf("reader.ReadString(): %s", err)
				}
			} else {
				log.Printf("Received: %s", data)
				req := NewRequest(c, strings.TrimSpace(data))
				req.handle()
			}
		}
	}
}

func (c *Client) listenWrite() {
	for {
		select {
		case <-c.doneCh:
			c.server.delClient(c)
			c.doneCh <- true
			return
		case msg := <-c.sendCh:
			c.conn.Write([]byte(msg))
		}
	}
}

// this throws data into the channel to be sent, and is how everything
// should route its messages to the client
func (c *Client) send(data string) {
	c.sendCh <- data
}
