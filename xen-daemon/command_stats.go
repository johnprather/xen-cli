package main

import (
	"fmt"
	"strings"
)

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
		outStr += fmt.Sprintf("XAPI Servers: %d\n", counts.Servers)
		outStr += fmt.Sprintf("Pools: %d", counts.Pools.Total)
		var poolManus []string
		if counts.Pools.Dell > 0 {
			poolManus = append(poolManus, fmt.Sprintf("Dell: %d", counts.Pools.Dell))
		}
		if counts.Pools.HP > 0 {
			poolManus = append(poolManus, fmt.Sprintf("HP: %d", counts.Pools.HP))
		}
		if len(poolManus) > 0 {
			outStr += fmt.Sprintf(" - %s", strings.Join(poolManus, ", "))
		}
		outStr += "\n"
		outStr += fmt.Sprintf("Hosts: %d", counts.Hosts.Total)
		var hostLives []string
		if counts.Hosts.Live > 0 {
			hostLives = append(hostLives, fmt.Sprintf("Live: %d", counts.Hosts.Live))
		}
		if counts.Hosts.Dead > 0 {
			hostLives = append(hostLives, fmt.Sprintf("Dead: %d", counts.Hosts.Dead))
		}
		if len(hostLives) > 0 {
			outStr += fmt.Sprintf(" - %s", strings.Join(hostLives, ", "))
		}
		var hostManus []string
		if counts.Hosts.Dell > 0 {
			hostManus = append(hostManus, fmt.Sprintf("Dell: %d", counts.Hosts.Dell))
		}
		if counts.Hosts.HP > 0 {
			hostManus = append(hostManus, fmt.Sprintf("HP: %d", counts.Hosts.HP))
		}
		if len(hostManus) > 0 {
			outStr += fmt.Sprintf(" - %s", strings.Join(hostManus, ", "))
		}
		outStr += "\n"
		outStr += fmt.Sprintf("VMs: %d", counts.VMs.Total)
		var vmPows []string
		if counts.VMs.Running > 0 {
			vmPows = append(vmPows, fmt.Sprintf("Running: %d", counts.VMs.Running))
		}
		if counts.VMs.Halted > 0 {
			vmPows = append(vmPows, fmt.Sprintf("Halted: %d", counts.VMs.Halted))
		}
		if counts.VMs.Paused > 0 {
			vmPows = append(vmPows, fmt.Sprintf("Paused: %d", counts.VMs.Paused))
		}
		if counts.VMs.Suspended > 0 {
			vmPows = append(vmPows, fmt.Sprintf("Suspended: %d", counts.VMs.Suspended))
		}
		if len(vmPows) > 0 {
			outStr += fmt.Sprintf(" - %s", strings.Join(vmPows, ", "))
		}
		outStr += "\n"
		outStr += fmt.Sprintf("SRs: %d", counts.SRs.Total)
		var srTypes []string
		if counts.SRs.NFS > 0 {
			srTypes = append(srTypes, fmt.Sprintf("NFS: %d", counts.SRs.NFS))
		}
		if counts.SRs.LVM > 0 {
			srTypes = append(srTypes, fmt.Sprintf("LVM: %d", counts.SRs.LVM))
		}
		if counts.SRs.UDev > 0 {
			srTypes = append(srTypes, fmt.Sprintf("UDev: %d", counts.SRs.UDev))
		}
		if counts.SRs.ISO > 0 {
			srTypes = append(srTypes, fmt.Sprintf("ISO: %d", counts.SRs.ISO))
		}
		if len(srTypes) > 0 {
			outStr += fmt.Sprintf(" - %s", strings.Join(srTypes, ", "))
		}
		outStr += "\n"
		req.client.send(outStr)
	}
	stats.help = func(req *Request) {
		outStr := "The \"stats\" command shows various application statistics.\n\n" +
			"Usage: stats\n"
		req.client.send(outStr)
	}
	commands[stats.name] = stats
}
