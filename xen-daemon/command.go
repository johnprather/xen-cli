package main

// Command is a user command
type Command struct {
	name        string
	desc        string
	run         func(req *Request)
	help        func(req *Request)
	subCommands map[string]*Command
}

// NewCommand returns an instantiated Command object
func NewCommand(name string, desc string) *Command {
	cmd := &Command{name: name, desc: desc}
	cmd.subCommands = make(map[string]*Command)
	return cmd
}

var commands map[string]*Command

func init() {
	commands = make(map[string]*Command)
}
