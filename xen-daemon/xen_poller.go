package main

import (
	"log"
	"time"

	xenAPI "github.com/johnprather/go-xen-api-client"
)

// XenPoller struct for manager of polling a host for data
type XenPoller struct {
	server  *XenServer
	doneCh  chan bool
	client  *xenAPI.Client
	session xenAPI.SessionRef
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
	log.Printf("starting polling loop for %s", p.server.Hostname)
	defer log.Printf("returning from polling loop for %s", p.server.Hostname)

	for {
		select {
		case <-p.doneCh:
			if p.client != nil {
				p.client.Close()
			}
			return
		default:
			pollDelay := 2
			p.gatherData()
			if !p.server.hasData() {
				pollDelay = 15
			}
			<-time.After(time.Duration(pollDelay) * time.Second)
		}
	}
}

func (p *XenPoller) gatherData() {
	if !p.server.hasData() {
		p.gatherAllData()
	} else {
		p.gatherEvents()
	}
}

func (p *XenPoller) gatherAllData() {
	defer log.Printf("%s: finsihed gatherAllData\n", p.server.Hostname)
	data := &XenDataSet{}
	xenClient, err := p.server.getClient()
	if err != nil {
		log.Println("getClient():", err)
		p.fail()
		return
	}

	sessionID, err := xenClient.Session.LoginWithPassword(p.server.User,
		p.server.password, "1.0", "xen-cli")
	if err != nil {
		log.Println("LoginWithPassword():", p.server.Hostname, err)
		p.fail()
		return
	}

	err = xenClient.Event.Register(sessionID, []string{"*"})
	if err != nil {
		log.Println("Event.Register():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.poolRecs, err = xenClient.Pool.GetAllRecords(sessionID)
	if err != nil {
		log.Println("Pool.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.hostRecs, err = xenClient.Host.GetAllRecords(sessionID)
	if err != nil {
		log.Println("Host.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}
	log.Printf("%s: loaded %d host records\n", p.server.Hostname,
		len(data.hostRecs))

	data.vmRecs, err = xenClient.VM.GetAllRecords(sessionID)
	if err != nil {
		log.Println("VM.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.vbdRecs, err = xenClient.VBD.GetAllRecords(sessionID)
	if err != nil {
		log.Println("VBD.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.vdiRecs, err = xenClient.VDI.GetAllRecords(sessionID)
	if err != nil {
		log.Println("VDI.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.srRecs, err = xenClient.SR.GetAllRecords(sessionID)
	if err != nil {
		log.Println("SR.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.consoleRecs, err = xenClient.Console.GetAllRecords(sessionID)
	if err != nil {
		log.Println("Console.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.pbdRecs, err = xenClient.PBD.GetAllRecords(sessionID)
	if err != nil {
		log.Println("PBD.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.pifRecs, err = xenClient.PIF.GetAllRecords(sessionID)
	if err != nil {
		log.Println("PIF.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.vifRecs, err = xenClient.VIF.GetAllRecords(sessionID)
	if err != nil {
		log.Println("VIF.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.messageRecs, err = xenClient.Message.GetAllRecords(sessionID)
	if err != nil {
		log.Println("VIF.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	data.taskRecs, err = xenClient.Task.GetAllRecords(sessionID)
	if err != nil {
		log.Println("Task.GetAllRecords():", p.server.Hostname, err)
		p.fail()
		return
	}

	p.client = xenClient
	p.session = sessionID
	p.server.setData(data)
	p.server.setLastUpdate()
}

func (p *XenPoller) gatherEvents() {
	var (
		host    = "host"
		message = "message"
		pool    = "pool"
		sr      = "sr"
		task    = "task"
		vbd     = "vbd"
		vdi     = "vdi"
		vm      = "vm"
	)
	data := p.server.getData()
	if data == nil {
		log.Println("data is nil for", p.server.Hostname)
		p.fail()
		return
	}

	events, err := p.client.Event.Next(p.session)
	if err != nil {
		log.Println("Event.From():", err)
		p.fail()
		return
	}
	for _, event := range events {
		var err error
		var method string
		switch {
		case event.Class == host && event.Operation == xenAPI.EventOperationAdd:
			err = p.addHost(xenAPI.HostRef(event.Ref))
		case event.Class == host && event.Operation == xenAPI.EventOperationDel:
			p.delHost(xenAPI.HostRef(event.Ref))
		case event.Class == host && event.Operation == xenAPI.EventOperationMod:
			err = p.modHost(xenAPI.HostRef(event.Ref))
		case event.Class == message && event.Operation == xenAPI.EventOperationAdd:
			err = p.addMessage(xenAPI.MessageRef(event.Ref))
		case event.Class == message && event.Operation == xenAPI.EventOperationDel:
			p.delMessage(xenAPI.MessageRef(event.Ref))
		case event.Class == message && event.Operation == xenAPI.EventOperationMod:
			err = p.modMessage(xenAPI.MessageRef(event.Ref))
		case event.Class == pool && event.Operation == xenAPI.EventOperationAdd:
			err = p.addPool(xenAPI.PoolRef(event.Ref))
		case event.Class == pool && event.Operation == xenAPI.EventOperationDel:
			p.delPool(xenAPI.PoolRef(event.Ref))
		case event.Class == pool && event.Operation == xenAPI.EventOperationMod:
			err = p.modPool(xenAPI.PoolRef(event.Ref))
		case event.Class == sr && event.Operation == xenAPI.EventOperationAdd:
			err = p.addSR(xenAPI.SRRef(event.Ref))
		case event.Class == sr && event.Operation == xenAPI.EventOperationDel:
			p.delSR(xenAPI.SRRef(event.Ref))
		case event.Class == sr && event.Operation == xenAPI.EventOperationMod:
			err = p.modSR(xenAPI.SRRef(event.Ref))
		case event.Class == task && event.Operation == xenAPI.EventOperationAdd:
			err = p.addTask(xenAPI.TaskRef(event.Ref))
		case event.Class == task && event.Operation == xenAPI.EventOperationDel:
			p.delTask(xenAPI.TaskRef(event.Ref))
		case event.Class == task && event.Operation == xenAPI.EventOperationMod:
			err = p.modTask(xenAPI.TaskRef(event.Ref))
		case event.Class == vbd && event.Operation == xenAPI.EventOperationAdd:
			err = p.addVBD(xenAPI.VBDRef(event.Ref))
		case event.Class == vbd && event.Operation == xenAPI.EventOperationDel:
			p.delVBD(xenAPI.VBDRef(event.Ref))
		case event.Class == vbd && event.Operation == xenAPI.EventOperationMod:
			err = p.modVBD(xenAPI.VBDRef(event.Ref))
		case event.Class == vdi && event.Operation == xenAPI.EventOperationAdd:
			err = p.addVDI(xenAPI.VDIRef(event.Ref))
		case event.Class == vdi && event.Operation == xenAPI.EventOperationDel:
			p.delVDI(xenAPI.VDIRef(event.Ref))
		case event.Class == vdi && event.Operation == xenAPI.EventOperationMod:
			err = p.modVDI(xenAPI.VDIRef(event.Ref))
		case event.Class == vm && event.Operation == xenAPI.EventOperationAdd:
			err = p.addVM(xenAPI.VMRef(event.Ref))
		case event.Class == vm && event.Operation == xenAPI.EventOperationDel:
			p.delVM(xenAPI.VMRef(event.Ref))
		case event.Class == vm && event.Operation == xenAPI.EventOperationMod:
			err = p.modVM(xenAPI.VMRef(event.Ref))
		default:
			log.Printf("%s: %s > %s > %s\n", p.server.Hostname, event.Class,
				event.Operation, event.Ref)
		}

		if err != nil && err.Error()[0:26] != "API Error: HANDLE_INVALID " {
			log.Printf("%s: %s(): %s\n", p.server.Hostname, method, err)
		}
		p.server.setLastUpdate()
	}
}

func (p *XenPoller) fail() {
	p.server.clearData()
}

func (p *XenPoller) addTask(task xenAPI.TaskRef) (err error) {
	return p.setTask(task, "added")
}

func (p *XenPoller) modTask(task xenAPI.TaskRef) (err error) {
	return p.setTask(task, "modified")
}

func (p *XenPoller) setTask(task xenAPI.TaskRef, action string) (err error) {
	taskRec, err := p.client.Task.GetRecord(p.session, task)
	if err != nil {
		return
	}
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	data.taskRecs[task] = taskRec
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	log.Printf("task %s: %s (%s)\n", action, taskRec.NameLabel,
		poolRec.NameLabel)
	return
}

func (p *XenPoller) delTask(task xenAPI.TaskRef) {
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	if taskRec, ok := data.taskRecs[task]; ok {
		delete(data.taskRecs, task)
		log.Printf("task removed: %s (%s)\n", taskRec.NameLabel, poolRec.NameLabel)
	}
}

func (p *XenPoller) addPool(pool xenAPI.PoolRef) (err error) {
	return p.setPool(pool, "added")
}
func (p *XenPoller) modPool(pool xenAPI.PoolRef) (err error) {
	return p.setPool(pool, "modified")
}

func (p *XenPoller) setPool(pool xenAPI.PoolRef, action string) (err error) {
	poolRec, err := p.client.Pool.GetRecord(p.session, pool)
	if err != nil {
		return
	}
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	data.poolRecs[pool] = poolRec
	log.Printf("pool %s: %s\n", action, poolRec.NameLabel)
	return
}

func (p *XenPoller) delPool(pool xenAPI.PoolRef) {
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	if poolRec, ok := data.poolRecs[pool]; ok {
		delete(data.poolRecs, pool)
		log.Printf("pool removed: %s\n", poolRec.NameLabel)
	}
}

func (p *XenPoller) addHost(host xenAPI.HostRef) (err error) {
	return p.setHost(host, "added")
}
func (p *XenPoller) modHost(host xenAPI.HostRef) (err error) {
	return p.setHost(host, "modified")
}

func (p *XenPoller) setHost(host xenAPI.HostRef, action string) (err error) {
	hostRec, err := p.client.Host.GetRecord(p.session, host)
	if err != nil {
		return
	}
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	data.hostRecs[host] = hostRec
	log.Printf("host %s: %s (%s)\n", action, hostRec.NameLabel,
		poolRec.NameLabel)
	return
}

func (p *XenPoller) delHost(host xenAPI.HostRef) {
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	if hostRec, ok := data.hostRecs[host]; ok {
		delete(data.hostRecs, host)
		log.Printf("host removed: %s (%s)\n", hostRec.NameLabel, poolRec.NameLabel)
	}
}

func (p *XenPoller) addVM(vm xenAPI.VMRef) (err error) {
	return p.setVM(vm, "added")
}
func (p *XenPoller) modVM(vm xenAPI.VMRef) (err error) {
	return p.setVM(vm, "modified")
}

func (p *XenPoller) setVM(vm xenAPI.VMRef, action string) (err error) {
	vmRec, err := p.client.VM.GetRecord(p.session, vm)
	if err != nil {
		return
	}
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	data.vmRecs[vm] = vmRec
	log.Printf("vm %s: %s (%s)\n", action, vmRec.NameLabel, poolRec.NameLabel)
	return
}

func (p *XenPoller) delVM(vm xenAPI.VMRef) {
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	if vmRec, ok := data.vmRecs[vm]; ok {
		delete(data.vmRecs, vm)
		log.Printf("vm removed: %s (%s)\n", vmRec.NameLabel, poolRec.NameLabel)
	}
}

func (p *XenPoller) addSR(sr xenAPI.SRRef) (err error) {
	return p.setSR(sr, "added")
}
func (p *XenPoller) modSR(sr xenAPI.SRRef) (err error) {
	return p.setSR(sr, "modified")
}

func (p *XenPoller) setSR(sr xenAPI.SRRef, action string) (err error) {
	srRec, err := p.client.SR.GetRecord(p.session, sr)
	if err != nil {
		return
	}
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	data.srRecs[sr] = srRec
	log.Printf("sr %s: %s (%s)\n", action, srRec.NameLabel, poolRec.NameLabel)
	return
}

func (p *XenPoller) delSR(sr xenAPI.SRRef) {
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	if srRec, ok := data.srRecs[sr]; ok {
		delete(data.srRecs, sr)
		log.Printf("sr removed: %s (%s)\n", srRec.NameLabel, poolRec.NameLabel)
	}
}

func (p *XenPoller) addMessage(message xenAPI.MessageRef) (err error) {
	return p.setMessage(message, "added")
}
func (p *XenPoller) modMessage(message xenAPI.MessageRef) (err error) {
	return p.setMessage(message, "modified")
}

func (p *XenPoller) setMessage(message xenAPI.MessageRef, action string) (err error) {
	messageRec, err := p.client.Message.GetRecord(p.session, message)
	if err != nil {
		return
	}
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	data.messageRecs[message] = messageRec
	log.Printf("message %s: %s (%s)\n", action, messageRec.Body, poolRec.NameLabel)
	return
}

func (p *XenPoller) delMessage(message xenAPI.MessageRef) {
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	if messageRec, ok := data.messageRecs[message]; ok {
		delete(data.messageRecs, message)
		log.Printf("message removed: %s (%s)\n", messageRec.Body, poolRec.NameLabel)
	}
}

func (p *XenPoller) addVBD(vbd xenAPI.VBDRef) (err error) {
	return p.setVBD(vbd, "added")
}
func (p *XenPoller) modVBD(vbd xenAPI.VBDRef) (err error) {
	return p.setVBD(vbd, "modified")
}

func (p *XenPoller) setVBD(vbd xenAPI.VBDRef, action string) (err error) {
	vbdRec, err := p.client.VBD.GetRecord(p.session, vbd)
	if err != nil {
		return
	}
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	data.vbdRecs[vbd] = vbdRec
	vmRec := data.vmRecs[vbdRec.VM]
	device := vbdRec.Device
	if vbdRec.Userdevice != "" {
		device = vbdRec.Userdevice
	}
	log.Printf("vbd %s: %s:%s\n", action, vmRec.NameLabel, device)
	return
}

func (p *XenPoller) delVBD(vbd xenAPI.VBDRef) {
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	if vbdRec, ok := data.vbdRecs[vbd]; ok {
		delete(data.vbdRecs, vbd)
		vmRec := data.vmRecs[vbdRec.VM]
		device := vbdRec.Device
		if vbdRec.Userdevice != "" {
			device = vbdRec.Userdevice
		}
		log.Printf("vbd removed: %s:%s\n", vmRec.NameLabel, device)
	}
}

func (p *XenPoller) addVDI(vdi xenAPI.VDIRef) (err error) {
	return p.setVDI(vdi, "added")
}
func (p *XenPoller) modVDI(vdi xenAPI.VDIRef) (err error) {
	return p.setVDI(vdi, "modified")
}

func (p *XenPoller) setVDI(vdi xenAPI.VDIRef, action string) (err error) {
	vdiRec, err := p.client.VDI.GetRecord(p.session, vdi)
	if err != nil {
		return
	}
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	data.vdiRecs[vdi] = vdiRec
	var poolRec xenAPI.PoolRecord
	for _, poolRec = range data.poolRecs {
		break
	}
	log.Printf("vdi %s: %s (%s)\n", action, vdiRec.NameLabel, poolRec.NameLabel)
	return
}

func (p *XenPoller) delVDI(vdi xenAPI.VDIRef) {
	data := p.server.getData()
	data.lock.Lock()
	defer data.lock.Unlock()
	if vdiRec, ok := data.vdiRecs[vdi]; ok {
		delete(data.vdiRecs, vdi)
		var poolRec xenAPI.PoolRecord
		for _, poolRec = range data.poolRecs {
			break
		}
		log.Printf("vdi removed: %s (%s)\n", vdiRec.NameLabel, poolRec.NameLabel)
	}
}
