package main

// Command is a user command
type Command struct {
	name        string
	desc        string
	run         func(req *Request)
	help        func(req *Request)
	subCommands map[string]*Command
}

var commands map[string]*Command

func init() {
	commands = make(map[string]*Command)
}
