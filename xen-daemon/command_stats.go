package main

import "fmt"

func init() {
	stats := NewCommand("stats", "view application statistics")
	stats.run = func(req *Request) {
		counts := xenData.countStats()
		outStr := "Xen Fleet Stats\n"
		var spaceStr string
		for i := 0; i < 26; i++ {
			spaceStr += "-"
		}
		spaceStr += "\n"
		outStr += spaceStr
		outStr += fmt.Sprintf("%20s %d\n", "XAPI Servers", counts.Servers)
		outStr += spaceStr
		outStr += fmt.Sprintf("%20s %d\n", "(Total) Pools", counts.Pools.Total)
		if counts.Pools.Dell > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(Dell) Pools", counts.Pools.Dell)
		}
		if counts.Pools.HP > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(HP) Pools", counts.Pools.HP)
		}
		outStr += spaceStr
		outStr += fmt.Sprintf("%20s %d\n", "(Total) Hosts", counts.Hosts.Total)
		if counts.Hosts.Dell > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(Dell) Hosts", counts.Hosts.Dell)
		}
		if counts.Hosts.HP > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(HP) Hosts", counts.Hosts.HP)
		}
		outStr += spaceStr
		outStr += fmt.Sprintf("%20s %d\n", "(Total) VMs", counts.VMs.Total)
		if counts.VMs.Running > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(Running) VMs", counts.VMs.Running)
		}
		if counts.VMs.Halted > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(Halted) VMs", counts.VMs.Halted)
		}
		if counts.VMs.Paused > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(Paused) VMs", counts.VMs.Paused)
		}
		if counts.VMs.Suspended > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(Suspended) VMs", counts.VMs.Suspended)
		}
		outStr += spaceStr
		outStr += fmt.Sprintf("%20s %d\n", "(Total) SRs", counts.SRs.Total)
		if counts.SRs.NFS > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(NFS) SRs", counts.SRs.NFS)
		}
		if counts.SRs.LVM > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(LVM) SRs", counts.SRs.LVM)
		}
		if counts.SRs.UDev > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(UDev) SRs", counts.SRs.UDev)
		}
		if counts.SRs.ISO > 0 {
			outStr += fmt.Sprintf("%20s %d\n", "(ISO) SRs", counts.SRs.ISO)
		}

		req.client.send(outStr)
	}
	stats.help = func(req *Request) {
		outStr := "The \"stats\" command shows various application statistics.\n\n" +
			"Usage: stats\n"
		req.client.send(outStr)
	}
	commands[stats.name] = stats
}
