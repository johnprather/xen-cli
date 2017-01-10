package main

import "fmt"

func init() {
	help := &Command{}
	help.name = "help"
	help.desc = "provides command usage information"
	help.run = func(req *Request) {
		switch {
		case len(req.args) == 1:
			outStr := "Use \"help <command>\" for help with <command>.\n\n" +
				"Available commands:\n"
			for _, cmd := range commands {
				outStr += fmt.Sprintf("%10s - %s\n", cmd.name, cmd.desc)
			}
			req.client.send(outStr)
		default:
			cmdList := commands
			argList := req.args[1:]
			var cmd *Command
			var ok bool
			for cmd, ok = cmdList[argList[0].Text]; ok; {
				if len(argList) > 1 {
					argList = argList[1:]
				} else {
					break
				}
			}
			if cmd == nil {
				req.client.send("Help unavailable, unrecognized command.\n")
				return
			}
			cmd.help(req)
		}
	}
	help.help = func(req *Request) {
		outStr := "The help command is for getting usage information about the " +
			"available commands.\n\n" +
			"Use \"help\" for a list of commands, \"help <command> ...\" for a " +
			"specific command.\n"
		req.client.send(outStr)
	}

	commands[help.name] = help
}
