package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	xenAPI "github.com/johnprather/go-xen-api-client"
)

// XenData a struct to hold all our xen data
type XenData struct {
	servers     []*XenServer
	serversLock sync.Mutex
	data        map[*XenServer]*XenDataSet
	pollers     map[*XenServer]*XenPoller
}

// XenDataSet struct to hold data for a specific host
type XenDataSet struct {
	poolRecs []*xenAPI.PoolRecord
}

var xenData *XenData

func init() {
	xenData = &XenData{}
	xenData.data = make(map[*XenServer]*XenDataSet)
	xenData.pollers = make(map[*XenServer]*XenPoller)
}

func (x *XenData) addServer(hostname string) (*XenServer, error) {
	svr, err := NewXenServer(hostname)
	if err != nil {
		return nil, err
	}
	x.serversLock.Lock()
	defer x.serversLock.Unlock()
	x.servers = append(x.servers, svr)
	x.pollers[svr] = NewXenPoller(svr)
	x.saveXenServers()
	return svr, nil
}

func (x *XenData) delServer(hostname string) (string, error) {
	x.serversLock.Lock()
	defer x.serversLock.Unlock()
	var deleted bool
	var deletedHostname string
	for key, svr := range x.servers {
		if svr.Hostname == hostname || svr.Hostname == hostname+"." {
			x.servers = append(x.servers[0:key], x.servers[key+1:]...)
			if _, ok := x.pollers[svr]; ok {
				// it may take the poller time to notice doneCh, don't wait
				// (we have a lock for x.servers, gotta be quick)
				go func() {
					x.pollers[svr].doneCh <- true
					delete(x.pollers, svr)
				}()
			}
			if _, ok := x.data[svr]; ok {
				delete(x.data, svr)
			}
			err := secure.DelPassword(svr.IP.String())
			if err != nil {
				log.Println("DelPassword():", err)
			}
			deleted = true
			deletedHostname = svr.Hostname
			break
		}
	}
	if !deleted {
		return "", fmt.Errorf("server not found: %s", hostname)
	}
	return deletedHostname, nil
}

func (x *XenData) saveXenServers() {
	data, err := json.Marshal(xenData.servers)
	if err != nil {
		log.Fatalln("Unable to marshal json for xenservers save:", err)
	}
	err = ioutil.WriteFile(config.serversFile, data, 0644)
	if err != nil {
		log.Fatalln("Unable to save xenservers file:", err)
	}
}

func (x *XenData) loadXenServers() {
	_, err := os.Stat(config.serversFile)
	if os.IsNotExist(err) {
		return
	}
	data, err := ioutil.ReadFile(config.serversFile)
	if err != nil {
		log.Fatalln("Unable to read xenservers file:", err)
	}
	var servers []*XenServer
	err = json.Unmarshal(data, &servers)
	if err != nil {
		log.Fatalln("Unable to unmarshal xenservers data:", err)
	}
	x.servers = servers
}

func (x *XenData) launchPollers() {
	x.serversLock.Lock()
	currentServers := x.servers
	x.serversLock.Unlock()

	for _, server := range currentServers {
		x.launchPoller(server)
	}
}

func (x *XenData) launchPoller(server *XenServer) {
	if _, ok := x.pollers[server]; ok {
		return
	}
	x.pollers[server] = NewXenPoller(server)
}
