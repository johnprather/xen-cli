package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"

	xenAPI "github.com/johnprather/go-xen-api-client"
)

// XenData a struct to hold all our xen data
type XenData struct {
	servers     []*XenServer
	serversLock sync.Mutex
	data        map[*XenServer]*XenDataSet
	dataLock    sync.Mutex
	pollers     map[*XenServer]*XenPoller
	pollersLock sync.Mutex
}

var xenData *XenData

func init() {
	xenData = &XenData{}
	xenData.data = make(map[*XenServer]*XenDataSet)
	xenData.pollers = make(map[*XenServer]*XenPoller)
}

func (x *XenData) getTransport() *http.Transport {
	xapiTransport := &http.Transport{}
	xapiTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return xapiTransport
}

func (x *XenData) getClient(hostname string) (*xenAPI.Client, error) {
	return xenAPI.NewClientTimeout("https://"+hostname, x.getTransport(), config.requestTimeout)
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
	x.saveXenServers()
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
	for _, server := range servers {
		pass, err := secure.GetPassword(server.IP.String())
		if err != nil {
			log.Println("secure.GetPassword():", server.Hostname, err)
			pass, err = secure.GetDefaultPassword()
			if err != nil {
				log.Fatalln("secure.GetDefaultPassword():", err)

			}
		}
		server.password = pass
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

func (x *XenData) hasData(server *XenServer) bool {
	x.dataLock.Lock()
	defer x.dataLock.Unlock()
	if _, ok := x.data[server]; ok {
		return true
	}
	return false
}

func (x *XenData) setData(server *XenServer, data *XenDataSet) {
	log.Printf("setting data for %s\n", server.Hostname)
	defer log.Printf("done setting data for %s\n", server.Hostname)
	x.dataLock.Lock()
	x.data[server] = data
	x.dataLock.Unlock()
}

func (x *XenData) clearData(server *XenServer) {
	x.dataLock.Lock()
	defer x.dataLock.Unlock()
	if _, ok := x.data[server]; ok {
		delete(x.data, server)
	}
}

func (x *XenData) getData() map[*XenServer]*XenDataSet {
	x.dataLock.Lock()
	defer x.dataLock.Unlock()
	data := make(map[*XenServer]*XenDataSet)
	for k, v := range x.data {
		data[k] = v
	}
	return data
}

func (x *XenData) getDataForServer(server *XenServer) *XenDataSet {
	x.dataLock.Lock()
	defer x.dataLock.Unlock()
	if data, ok := x.data[server]; ok {
		return data
	}
	return nil
}

func (x *XenData) findVMs(nameLabel string) map[xenAPI.VMRef]*XenServer {
	vms := make(map[xenAPI.VMRef]*XenServer)
	alldata := x.getData()
	for svr, data := range alldata {
		data.lock.Lock()
		for vm, vmRec := range data.vmRecs {
			if vmRec.NameLabel == nameLabel {
				vms[vm] = svr
			}
		}
		data.lock.Unlock()
	}
	return vms
}

func (x *XenData) findOneVM(nameLabel string) (xenAPI.VMRef, *XenServer, error) {
	vms := x.findVMs(nameLabel)
	switch {
	case len(vms) == 0:
		return "", nil, fmt.Errorf("No VMs found with name-label %s", nameLabel)
	case len(vms) > 1:
		return "", nil, fmt.Errorf("Ambiguous name-label %s used by %d VMs",
			nameLabel, len(vms))
	}
	var vm xenAPI.VMRef
	var server *XenServer
	for vm, server = range vms {
		break
	}
	return vm, server, nil
}

func (x *XenData) countServers() int {
	x.serversLock.Lock()
	defer x.serversLock.Unlock()
	return len(x.servers)
}

// StatCounts a struct for vm counts
type StatCounts struct {
	Servers   int
	Pools     StatCountsPools
	Hosts     StatCountsHosts
	VMs       StatCountsVMs
	SRs       StatCountsSRs
	Templates int
}

// StatCountsVMs is a struct for vm-based counts
type StatCountsVMs struct {
	Total     int
	Running   int
	Halted    int
	Paused    int
	Suspended int
}

// StatCountsSRs is a struct for sr-based counts
type StatCountsSRs struct {
	Total int
	NFS   int
	ISO   int
	LVM   int
	UDev  int
}

// StatCountsHosts is a struct for host-based counts
type StatCountsHosts struct {
	Total int
	HP    int
	Dell  int
	Live  int
	Dead  int
}

// StatCountsPools is a struct for pool-based counts
type StatCountsPools struct {
	Total int
	HP    int
	Dell  int
}

func (x *XenData) countStats() (counts *StatCounts) {
	counts = &StatCounts{}

	x.serversLock.Lock()
	counts.Servers = len(x.servers)
	x.serversLock.Unlock()

	alldata := x.getData()
	for _, data := range alldata {
		data = data.dup()

		countedPoolManus := false
		counts.Pools.Total += len(data.poolRecs)
		counts.Hosts.Total += len(data.hostRecs)
		for _, hostRec := range data.hostRecs {
			if hostMetricsRec, ok := data.hostMetricsRecs[hostRec.Metrics]; ok &&
				hostMetricsRec.Live {
				counts.Hosts.Live++
			} else {
				counts.Hosts.Dead++
			}
			if manu, ok := hostRec.BiosStrings["system-manufacturer"]; ok {
				if match, err := regexp.MatchString("Dell", manu); err == nil && match {
					counts.Hosts.Dell++
					if !countedPoolManus {
						counts.Pools.Dell++
						countedPoolManus = true
					}

				} else if match, err := regexp.MatchString("(HP)|(Hewlett ?Packard)", manu); err == nil && match {
					counts.Hosts.HP++
					if !countedPoolManus {
						counts.Pools.HP++
						countedPoolManus = true
					}
				}
			}

		}
		for _, vmRec := range data.vmRecs {
			if vmRec.IsATemplate {
				counts.Templates++
			} else {
				counts.VMs.Total++
				switch vmRec.PowerState {
				case xenAPI.VMPowerStateRunning:
					counts.VMs.Running++
				case xenAPI.VMPowerStateHalted:
					counts.VMs.Halted++
				case xenAPI.VMPowerStatePaused:
					counts.VMs.Paused++
				case xenAPI.VMPowerStateSuspended:
					counts.VMs.Suspended++
				}
			}
		}
		for _, srRec := range data.srRecs {
			counts.SRs.Total++
			switch srRec.Type {
			case "nfs":
				counts.SRs.NFS++
			case "iso":
				counts.SRs.ISO++
			case "lvm":
				counts.SRs.LVM++
			case "udev":
				counts.SRs.UDev++
			}

		}
	}
	return
}
