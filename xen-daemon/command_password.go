package main

import "fmt"

func init() {
	password := NewCommand("password", "set default xapi server password")
	password.run = func(req *Request) {
		if len(req.args) != 2 {
			outStr := "Usage: password <password>"
			req.client.send(outStr)
			return
		}
		err := secure.SetDefaultPassword(req.args[1].Text)
		if err != nil {
			req.client.send(fmt.Sprintf("secure.SetDefaultPassword: %s\n", err))
			return
		}
		req.client.send("Password set.\n")
	}
	password.help = func(req *Request) {
		outStr := "The password command is for setting the default xapi server " +
			"password.\n\nNOTE YOU SHOULD PROBABLY USE THE -password ARGUMENT ON " +
			"THE CLIENT INSTEAD\nIN ORDER TO ENTER A PASSWORD SECURELY.\n"
		req.client.send(outStr)
	}

	commands[password.name] = password
}
