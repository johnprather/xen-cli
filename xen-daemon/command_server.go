package main

import "fmt"

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
		for _, server := range xenData.servers {
			outStr += fmt.Sprintf("\t%s\n", server.hostname)
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

	}
	serverAdd.help = func(req *Request) {
		outStr := "Command \"server add\" will add a new xapi server.\n\n" +
			"Usage: server add <hostname> <user>\n\n" +
			"Note that if you have not yet configured your xapi password, you will " +
			"need\nto do so in order for this to work.  Use -password.set option " +
			"on the client\nto set the password to be used for new servers."
		req.client.send(outStr)
	}
	server.subCommands[serverAdd.name] = serverAdd
}
