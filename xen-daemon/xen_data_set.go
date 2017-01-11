package main

import (
	"sync"

	xenAPI "github.com/johnprather/go-xen-api-client"
)

// XenDataSet struct to hold data for a specific host
type XenDataSet struct {
	consoleRecs   map[xenAPI.ConsoleRef]xenAPI.ConsoleRecord
	hostRecs      map[xenAPI.HostRef]xenAPI.HostRecord
	messageRecs   map[xenAPI.MessageRef]xenAPI.MessageRecord
	pbdRecs       map[xenAPI.PBDRef]xenAPI.PBDRecord
	pifRecs       map[xenAPI.PIFRef]xenAPI.PIFRecord
	poolRecs      map[xenAPI.PoolRef]xenAPI.PoolRecord
	srRecs        map[xenAPI.SRRef]xenAPI.SRRecord
	taskRecs      map[xenAPI.TaskRef]xenAPI.TaskRecord
	vbdRecs       map[xenAPI.VBDRef]xenAPI.VBDRecord
	vdiRecs       map[xenAPI.VDIRef]xenAPI.VDIRecord
	vifRecs       map[xenAPI.VIFRef]xenAPI.VIFRecord
	vmRecs        map[xenAPI.VMRef]xenAPI.VMRecord
	vmMetricsRecs map[xenAPI.VMMetricsRef]xenAPI.VMMetricsRecord

	lock sync.Mutex
}
