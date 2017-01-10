package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	xenAPI "github.com/johnprather/go-xen-api-client"
)

// XenServer a struct to hold xapi server data
type XenServer struct {
	Hostname string
	User     string
	IP       net.IP
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
	xenClient, err := xenAPI.NewClient("https://"+svr.Hostname, nil)
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
	err = secure.SetPassword(svr.IP.String(), pass)
	if err != nil {
		err = fmt.Errorf("SetPassword(): %s", err)
		return
	}

	server = svr
	return
}
