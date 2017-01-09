package main

import (
	"log"
	"net"
	"os"
)

// Server is the server running on the socket
type Server struct {
	clients      map[ClientID]*Client
	nextClientID ClientID
	addCh        chan *Client
	delCh        chan *Client
}

// NewServer returns an instantiated Server object
func NewServer() *Server {
	server := &Server{}
	server.clients = make(map[ClientID]*Client)
	server.addCh = make(chan *Client)
	server.delCh = make(chan *Client)
	return server
}

func (s *Server) addClient(c *Client) {
	s.addCh <- c
}

func (s *Server) delClient(c *Client) {
	s.delCh <- c
}

func (s *Server) listen() {
	go s.listenConnects()
	s.listenWrite()
}

func (s *Server) listenConnects() {
	// unlink the socket if it already exists (avoid address in use error)
	if _, err := os.Stat(config.socketPath); !os.IsNotExist(err) {
		conn, err := net.Dial("unix", config.socketPath)
		if err == nil {
			conn.Close()
			log.Fatalln("A process is already listening on", config.socketPath)
		}
		err = os.Remove(config.socketPath)
		if err != nil {
			log.Fatalln("Unable to remove old socket", config.socketPath)
		}
	}

	listen, err := net.Listen("unix", config.socketPath)
	if err != nil {
		log.Fatalf("Listen(): %s\n", err)
	}
	for {
		fd, err := listen.Accept()
		if err != nil {
			log.Printf("listen.Accept(): %s\n", err)
		}
		client := NewClient(s, fd)
		s.addClient(client)
	}
}

func (s *Server) listenWrite() {
	for {
		select {
		case c := <-s.addCh:
			log.Printf("Added new client (id: %d)\n", c.ID)
			s.clients[c.ID] = c
			go c.listen()
		case c := <-s.delCh:
			log.Printf("Lost client (id: %d)\n", c.ID)
			delete(s.clients, c.ID)
		}
	}
}
