package main

import (
	"fmt"
	"strings"

	xenAPI "github.com/johnprather/go-xen-api-client"
)

func init() {
	forceStr := "force"

	vmcmd := NewCommand("vm", "view and manage VMs")
	vmcmd.run = func(req *Request) {
		switch {
		case len(req.args) == 1:
			outStr := "The \"vm\" commands are for managing Xen VMs.\n\n" +
				"Available commands:\n"
			for _, cmd := range vmcmd.subCommands {
				outStr += fmt.Sprintf("%10s - %s\n", cmd.name, cmd.desc)
			}
			req.client.send(outStr)
		default:
			if cmd, ok := vmcmd.subCommands[req.args[1].Text]; ok {
				cmd.run(req)
				return
			}
			req.client.send(fmt.Sprintf("%s %s: no such command.\n",
				req.args[0].Text, req.args[1].Text))
		}
	}
	vmcmd.help = func(req *Request) {

		outStr := "The vm command is for handling Xen VMs. It has " +
			"several subcommands.\n\n" +
			"Available SubCommands:\n"
		for _, cmd := range vmcmd.subCommands {
			outStr += fmt.Sprintf("%10s - %s\n", cmd.name, cmd.desc)
		}
		req.client.send(outStr)
	}
	commands[vmcmd.name] = vmcmd

	vmshow := NewCommand("show", "view and manage VMs")
	vmshow.run = func(req *Request) {
		if len(req.args) != 3 {
			req.client.send("Usage: vm show <name-label>\n")
			return
		}
		vmName := req.args[2].Text
		var outStrs []string
		alldata := xenData.getData()
		for _, data := range alldata {
			data.lock.Lock()
			for _, vmRec := range data.vmRecs {
				if vmRec.NameLabel == vmName {
					outStr := fmt.Sprintf("name-label: %s\n", vmRec.NameLabel)
					if len(data.poolRecs) > 0 {
						var poolRec xenAPI.PoolRecord
						for _, poolRec = range data.poolRecs {
							break
						}
						outStr += fmt.Sprintf("pool: %s\n", poolRec.NameLabel)
					}
					outStr += fmt.Sprintf("power-state: %s\n", vmRec.PowerState)
					if vmRec.PowerState == xenAPI.VMPowerStateRunning {
						if hostRec, ok := data.hostRecs[vmRec.ResidentOn]; ok {
							outStr += fmt.Sprintf("resident-on: %s\n", hostRec.Hostname)
						}
					}
					outStrs = append(outStrs, outStr)
				}
			}

			data.lock.Unlock()
		}
		if len(outStrs) == 0 {
			req.client.send(fmt.Sprintf("No VMs with name-label: %s\n", vmName))
			return
		}
		req.client.send(strings.Join(outStrs, "\n"))

	}
	vmshow.help = func(req *Request) {
		outStr := "The \"vm show\" command displays VM information.\n\n" +
			"Usage: vm show <name-label>\n"
		req.client.send(outStr)
	}
	vmcmd.subCommands[vmshow.name] = vmshow

	vmstop := NewCommand("shutdown", "shutdown a vm")
	vmstop.run = func(req *Request) {
		argCount := len(req.args)
		if argCount != 3 && !(argCount == 4 && req.args[3].Text == forceStr) {
			req.client.send("Usage: vm shutdown <name-label> [force]\n")
			return
		}
		nameLabel := req.args[2].Text
		vm, server, err := xenData.findOneVM(nameLabel)
		if err != nil {
			req.client.send(fmt.Sprintf("%s\n", err))
			return
		}
		sessionID, client, err := server.getSessionAndNewClient()
		if err != nil {
			req.client.send(fmt.Sprintf("getSessionAndNewClient(): %s\n", err))
			return
		}
		defer client.Close()
		var stopFunc string
		if argCount == 4 && req.args[3].Text == forceStr {
			err = client.VM.HardShutdown(sessionID, vm)
			stopFunc = "HardShutdown"
		} else {
			err = client.VM.Shutdown(sessionID, vm)
			stopFunc = "Shutdown"
		}
		if err != nil {
			req.client.send(fmt.Sprintf("VM.%s(): %s\n", stopFunc, err))
			return
		}
		req.client.send(fmt.Sprintf("shutdown (%s) complete for %s\n",
			stopFunc, req.args[2].Text))
	}
	vmstop.help = func(req *Request) {
		outStr := "The \"vm shutdown\" command attempts to shutdown a VM.\n\n" +
			"Usage: vm shutdown <name-label> [force]\n\n" +
			"If \"force\" is specified, the VM will be forcibly powered off rather " +
			"than relying\non the OS to gracefully handle the ACPI shutdown request."
		req.client.send(outStr)
	}
	vmcmd.subCommands[vmstop.name] = vmstop

	vmstart := NewCommand("start", "start a vm")
	vmstart.run = func(req *Request) {
		argCount := len(req.args)
		if argCount != 3 {
			req.client.send("Usage: vm start <name-label>\n")
			return
		}
		nameLabel := req.args[2].Text
		vm, server, err := xenData.findOneVM(nameLabel)
		if err != nil {
			req.client.send(fmt.Sprintf("%s\n", err))
			return
		}
		sessionID, client, err := server.getSessionAndNewClient()
		if err != nil {
			req.client.send(fmt.Sprintf("getSessionAndNewClient(): %s\n", err))
			return
		}
		defer client.Close()
		err = client.VM.Start(sessionID, vm, false, false)
		if err != nil {
			req.client.send(fmt.Sprintf("VM.Start(): %s\n", err))
			return
		}
		req.client.send(fmt.Sprintf("start complete for %s\n",
			req.args[2].Text))
	}
	vmstart.help = func(req *Request) {
		outStr := "The \"vm start\" command attempts to start a VM.\n\n" +
			"Usage: vm start <name-label>\n"
		req.client.send(outStr)
	}
	vmcmd.subCommands[vmstart.name] = vmstart

	vmreset := NewCommand("reset", "hard reset a vm")
	vmreset.run = func(req *Request) {
		argCount := len(req.args)
		if argCount != 3 {
			req.client.send("Usage: vm reset <name-label>\n")
			return
		}
		nameLabel := req.args[2].Text
		vm, server, err := xenData.findOneVM(nameLabel)
		if err != nil {
			req.client.send(fmt.Sprintf("%s\n", err))
			return
		}
		sessionID, client, err := server.getSessionAndNewClient()
		if err != nil {
			req.client.send(fmt.Sprintf("getSessionAndNewClient(): %s\n", err))
			return
		}
		defer client.Close()

		err = client.VM.HardReboot(sessionID, vm)
		if err != nil {
			req.client.send(fmt.Sprintf("VM.HardReboot(): %s\n", err))
			return
		}
		req.client.send(fmt.Sprintf("reset complete for %s\n",
			req.args[2].Text))
	}
	vmstart.help = func(req *Request) {
		outStr := "The \"vm reset\" command attempts to reset a VM.\n\n" +
			"Usage: vm reset <name-label>\n"
		req.client.send(outStr)
	}
	vmcmd.subCommands[vmreset.name] = vmreset

}
