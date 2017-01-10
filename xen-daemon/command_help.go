package main

import (
	"fmt"
	"sort"
)

func init() {
	help := &Command{}
	help.name = "help"
	help.desc = "provides command usage information"
	help.run = func(req *Request) {
		switch {
		case len(req.args) == 1:
			outStr := "Use \"help <command>\" for help with <command>.\n\n" +
				"Available commands:\n\n"
			var cmdList []string
			for name := range commands {
				cmdList = append(cmdList, name)
			}
			sort.Strings(cmdList)
			for _, name := range cmdList {
				cmd := commands[name]
				outStr += fmt.Sprintf("%10s - %s\n", name, cmd.desc)
			}
			outStr += "\n"
			req.client.send(outStr)
		default:
			cmdList := commands
			argList := req.args[1:]
			var command *Command

			for {
				cmd, ok := cmdList[argList[0].Text]
				if !ok {
					break
				}
				command = cmd
				if len(argList) > 1 {
					argList = argList[1:]
					cmdList = command.subCommands
				} else {
					break
				}
			}
			if command == nil {
				req.client.send("Help unavailable, unrecognized command.\n")
				return
			}
			command.help(req)
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
