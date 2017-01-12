package main

import (
	"fmt"
	"time"
)

func init() {
	server := NewCommand("server", "manage xapi servers")
	server.run = func(req *Request) {
		switch {
		case len(req.args) == 1:
			outStr := "The \"server\" commands are for managing xapi servers.\n\n" +
				"Available commands:\n"
			for _, cmd := range server.subCommands {
				outStr += fmt.Sprintf("%10s - %s\n", cmd.name, cmd.desc)
			}
			req.client.send(outStr)
		default:
			if cmd, ok := server.subCommands[req.args[1].Text]; ok {
				cmd.run(req)
				return
			}
			req.client.send(fmt.Sprintf("%s %s: no such command.\n",
				req.args[0].Text, req.args[1].Text))
		}
	}
	server.help = func(req *Request) {
		outStr := "The server command is for handling xapi servers. It has " +
			"several subcommands\n" +
			"for adding, removing, and otherwise managing xapi servers.\n\n" +
			"Available SubCommands:\n"
		for _, cmd := range server.subCommands {
			outStr += fmt.Sprintf("%10s - %s\n", cmd.name, cmd.desc)
		}

		req.client.send(outStr)
	}

	commands[server.name] = server

	serverList := NewCommand("list", "list configured xapi servers")
	serverList.run = func(req *Request) {
		outStr := "XAPI Servers:\n"
		xenData.serversLock.Lock()
		currentServers := xenData.servers
		xenData.serversLock.Unlock()
		for _, server := range currentServers {
			var state string
			if server.hasData() {
				lastUpdate := server.getLastUpdate()
				seconds := time.Now().Unix() - lastUpdate.Unix()
				state = fmt.Sprintf("updated %d seconds ago", seconds)
			} else {
				state = "no data"
			}

			outStr += fmt.Sprintf("\t%-40s - %s\n", server.Hostname, state)
		}
		req.client.send(outStr)
	}
	serverList.help = func(req *Request) {
		outStr := "Command \"server list\" will provide a list of the currently " +
			"configured\nXAPI servers.\n"
		req.client.send(outStr)
	}
	server.subCommands[serverList.name] = serverList

	serverAdd := NewCommand("add", "add xapi server")
	serverAdd.run = func(req *Request) {
		if len(req.args) != 3 {
			req.client.send("Usage: server add <hostname>\n")
			return
		}
		svr, err := xenData.addServer(req.args[2].Text)
		if err != nil {
			req.client.send(fmt.Sprintf("Error adding server %s: %s\n",
				req.args[2].Text, err))
			return
		}
		req.client.send(fmt.Sprintf("Server added: %s\n", svr.Hostname))
	}
	serverAdd.help = func(req *Request) {
		outStr := "Command \"server add\" will add a new xapi server.\n\n" +
			"Usage: server add <hostname>\n\n" +
			"Note that if you have not yet configured your xapi password, you will " +
			"need\nto do so in order for this to work.  Use -password.set option " +
			"on the client\nto set the password to be used for new servers.\n"
		req.client.send(outStr)
	}
	server.subCommands[serverAdd.name] = serverAdd

	serverDel := NewCommand("remove", "remove xapi server")
	serverDel.run = func(req *Request) {
		if len(req.args) != 3 {
			req.client.send("Usage: server remove <hostname>\n")
			return
		}
		hostname, err := xenData.delServer(req.args[2].Text)
		if err != nil {
			req.client.send(fmt.Sprintf("Error removing server server %s: %s\n",
				req.args[2].Text, err))
			return
		}
		req.client.send(fmt.Sprintf("Server removed: %s\n", hostname))
	}
	serverDel.help = func(req *Request) {
		outStr := "Command \"server remove\" will remove an xapi server.\n\n" +
			"Usage: server remove <hostname>\n"
		req.client.send(outStr)
	}
	server.subCommands[serverDel.name] = serverDel

	serverClear := NewCommand("clearall", "remove all xapi servers")
	serverClear.run = func(req *Request) {
		if len(req.args) != 2 {
			req.client.send("Usage: server clearall\n")
			return
		}
		var outStr string
		var svrs []*XenServer
		xenData.serversLock.Lock()
		for _, svr := range xenData.servers {
			svrs = append(svrs, svr)
		}
		xenData.serversLock.Unlock()

		for _, svr := range svrs {
			hostname, err := xenData.delServer(svr.Hostname)
			if err != nil {
				outStr += fmt.Sprintf("Error removing %s: %s\n", svr.Hostname, err)
			} else {
				outStr += fmt.Sprintf("Removed server %s\n", hostname)
			}
		}
		req.client.send(outStr)
	}
	serverClear.help = func(req *Request) {
		outStr := "Command \"server clearall\" will remove all xapi servers.\n\n" +
			"Usage: server clearall\n"
		req.client.send(outStr)
	}
	server.subCommands[serverClear.name] = serverClear
}
