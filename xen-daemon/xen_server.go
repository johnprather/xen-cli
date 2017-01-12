package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	xenAPI "github.com/johnprather/go-xen-api-client"
)

// XenServer a struct to hold xapi server data
type XenServer struct {
	Hostname       string
	User           string
	IP             net.IP
	lastUpdate     time.Time
	lastUpdateLock sync.Mutex // for read/write to lastUpdate
	password       string
	session        xenAPI.SessionRef
	sessionLock    sync.Mutex
}

// NewXenServer returns an instantiated XenServer object
func NewXenServer(hostname string) (server *XenServer, err error) {
	hosts, err := net.LookupHost(hostname)
	if err != nil {
		err = fmt.Errorf("net.LookupHost(): %s", err)
		return
	}
	if len(hosts) == 0 {
		err = fmt.Errorf("no host records for %s", hostname)
		return
	}

	log.Println("host:", hosts[0])

	ipList, err := net.LookupIP(hosts[0])
	if err != nil {
		err = fmt.Errorf("net.LookupIP(): %s", err)
		return
	}
	if len(ipList) == 0 {
		err = errors.New("unable to lookup IP for " + hostname)
		return
	}

	log.Println("ip:", ipList[0].String())

	svr := &XenServer{}
	svr.IP = ipList[0]

	for _, existingServer := range xenData.servers {
		if string(svr.IP) == string(existingServer.IP) {
			err = errors.New("server already setup: " + hostname)
			return
		}
	}

	realHostnames, err := net.LookupAddr(svr.IP.String())
	if err != nil {
		err = fmt.Errorf("net.LookupAddr(): %s", err)
		return
	}
	if len(realHostnames) == 0 {
		err = errors.New("unable to lookup hostname for " + hostname)
		return
	}
	log.Println("real hostname:", realHostnames[0])
	svr.Hostname = realHostnames[0]

	svr.User = "root"

	pass, err := secure.GetDefaultPassword()
	if err != nil {
		return
	}
	xenClient, err := xenData.getClient(svr.Hostname)
	if err != nil {
		return
	}
	defer xenClient.Close()
	_, err = xenClient.Session.LoginWithPassword(svr.User, pass, "1.0",
		"xen-cli")
	if err != nil {
		errParts := strings.Split(err.Error(), " ")
		if len(errParts) >= 4 && errParts[2] == "HOST_IS_SLAVE" {
			return NewXenServer(errParts[3])
		}
		err = fmt.Errorf("LoginWithPassword(): %s", err)
		return
	}
	svr.password = pass
	err = secure.SetPassword(svr.IP.String(), pass)
	if err != nil {
		err = fmt.Errorf("SetPassword(): %s", err)
		return
	}

	server = svr
	return
}

func (s *XenServer) hasData() bool {
	return xenData.hasData(s)
}

func (s *XenServer) setData(data *XenDataSet) {
	xenData.setData(s, data)
}

func (s *XenServer) clearData() {
	xenData.clearData(s)
}

func (s *XenServer) getLastUpdate() time.Time {
	s.lastUpdateLock.Lock()
	defer s.lastUpdateLock.Unlock()
	lastUpdate := s.lastUpdate
	return lastUpdate
}

func (s *XenServer) setLastUpdate() {
	s.lastUpdateLock.Lock()
	defer s.lastUpdateLock.Unlock()
	s.lastUpdate = time.Now()
}

func (s *XenServer) getData() *XenDataSet {
	return xenData.getDataForServer(s)
}

func (s *XenServer) getSession() (xenAPI.SessionRef, error) {
	s.sessionLock.Lock()
	defer s.sessionLock.Unlock()
	if s.session == "" {
		log.Println("establishing new session for", s.Hostname)
		client, err := s.NewClient()
		if err != nil {
			return "", fmt.Errorf("server.NewClient(): %s", err)
		}
		defer client.Close()
		session, err := client.Session.LoginWithPassword(s.User, s.password,
			"1.0", "xen-cli")
		if err != nil {
			return "", fmt.Errorf("LoginWithPassword(): %s", err)
		}
		s.session = session
	}
	return s.session, nil
}

// NewClient instantiates a new XenServer client
func (s *XenServer) NewClient() (*xenAPI.Client, error) {
	client, err := xenAPI.NewClient("https://"+s.IP.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("xenAPI.NewClient(): %s", err)
	}
	return client, nil
}

func (s *XenServer) getSessionAndNewClient() (xenAPI.SessionRef,
	*xenAPI.Client, error) {
	sessionID, err := s.getSession()
	if err != nil {
		return "", nil, err
	}
	client, err := s.NewClient()
	if err != nil {
		return "", nil, err
	}
	return sessionID, client, nil
}
