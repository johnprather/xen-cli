package main

import (
	"bufio"
	"io"
	"net"
	"strings"
	"time"
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

func (c *Client) done() {
	c.doneCh <- true
	c.doneCh <- true
}

func (c *Client) listenRead() {
	defer c.conn.Close()

	reader := bufio.NewReader(c.conn)

	inLine := ""
	inBytes := make([]byte, 1)
	for {
		select {
		case <-c.doneCh:
			c.server.delClient(c)
			return
		default:
			c.conn.SetReadDeadline(time.Now().Add(time.Duration(1) * time.Second))
			numBytes, err := reader.Read(inBytes)
			if err == io.EOF {
				go c.done()

			} else if err != nil {
				if err, ok := err.(net.Error); !ok || !err.Timeout() {
					go c.done()
				}
			} else if numBytes > 0 {
				if inBytes[0] == '\n' {
					req := NewRequest(c, strings.TrimSpace(inLine))
					inLine = ""
					req.handle()
				} else {
					inLine += string(inBytes)
				}
			}
		}
	}
}

func (c *Client) listenWrite() {
	for {
		select {
		case <-c.doneCh:
			return
		case msg := <-c.sendCh:
			if msg[len(msg)-1] != '\n' {
				msg += "\n"
			}
			msg += ".\n"
			c.conn.Write([]byte(msg))
		}
	}
}

// this throws data into the channel to be sent, and is how everything
// should route its messages to the client
func (c *Client) send(data string) {
	c.sendCh <- data
}
