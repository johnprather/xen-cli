package main

import (
	"sync"
	"time"

	xenAPI "github.com/johnprather/go-xen-api-client"
)

// XenDataSet struct to hold data for a specific host
type XenDataSet struct {
	hostRecs              map[xenAPI.HostRef]xenAPI.HostRecord
	messageRecs           map[xenAPI.MessageRef]xenAPI.MessageRecord
	pbdRecs               map[xenAPI.PBDRef]xenAPI.PBDRecord
	pifRecs               map[xenAPI.PIFRef]xenAPI.PIFRecord
	poolRecs              map[xenAPI.PoolRef]xenAPI.PoolRecord
	srRecs                map[xenAPI.SRRef]xenAPI.SRRecord
	taskRecs              map[xenAPI.TaskRef]xenAPI.TaskRecord
	vbdRecs               map[xenAPI.VBDRef]xenAPI.VBDRecord
	vdiRecs               map[xenAPI.VDIRef]xenAPI.VDIRecord
	vifRecs               map[xenAPI.VIFRef]xenAPI.VIFRecord
	vmRecs                map[xenAPI.VMRef]xenAPI.VMRecord
	hostMetricsRecs       map[xenAPI.HostMetricsRef]xenAPI.HostMetricsRecord
	lastHostMetricsUpdate time.Time
	lock                  sync.Mutex
}

func (d *XenDataSet) dup() *XenDataSet {
	dup := &XenDataSet{}
	dup.hostRecs = make(map[xenAPI.HostRef]xenAPI.HostRecord)
	dup.messageRecs = make(map[xenAPI.MessageRef]xenAPI.MessageRecord)
	dup.pbdRecs = make(map[xenAPI.PBDRef]xenAPI.PBDRecord)
	dup.pifRecs = make(map[xenAPI.PIFRef]xenAPI.PIFRecord)
	dup.poolRecs = make(map[xenAPI.PoolRef]xenAPI.PoolRecord)
	dup.srRecs = make(map[xenAPI.SRRef]xenAPI.SRRecord)
	dup.taskRecs = make(map[xenAPI.TaskRef]xenAPI.TaskRecord)
	dup.vbdRecs = make(map[xenAPI.VBDRef]xenAPI.VBDRecord)
	dup.vdiRecs = make(map[xenAPI.VDIRef]xenAPI.VDIRecord)
	dup.vifRecs = make(map[xenAPI.VIFRef]xenAPI.VIFRecord)
	dup.vmRecs = make(map[xenAPI.VMRef]xenAPI.VMRecord)
	dup.hostMetricsRecs = make(map[xenAPI.HostMetricsRef]xenAPI.HostMetricsRecord)
	d.lock.Lock()
	defer d.lock.Unlock()
	for k, v := range d.hostRecs {
		dup.hostRecs[k] = v
	}
	for k, v := range d.messageRecs {
		dup.messageRecs[k] = v
	}
	for k, v := range d.pbdRecs {
		dup.pbdRecs[k] = v
	}
	for k, v := range d.pifRecs {
		dup.pifRecs[k] = v
	}
	for k, v := range d.poolRecs {
		dup.poolRecs[k] = v
	}
	for k, v := range d.srRecs {
		dup.srRecs[k] = v
	}
	for k, v := range d.taskRecs {
		dup.taskRecs[k] = v
	}
	for k, v := range d.vbdRecs {
		dup.vbdRecs[k] = v
	}
	for k, v := range d.vdiRecs {
		dup.vdiRecs[k] = v
	}
	for k, v := range d.vifRecs {
		dup.vifRecs[k] = v
	}
	for k, v := range d.vmRecs {
		dup.vmRecs[k] = v
	}
	for k, v := range d.hostMetricsRecs {
		dup.hostMetricsRecs[k] = v
	}
	dup.lastHostMetricsUpdate = d.lastHostMetricsUpdate
	return dup
}
