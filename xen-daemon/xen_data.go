package main

// XenData a struct to hold all our xen data
type XenData struct {
	servers []*XenServer
}

// XenServer a struct to hold xapi server data
type XenServer struct {
	hostname string
	user     string
}

var xenData *XenData

func init() {
	xenData = &XenData{}
}
