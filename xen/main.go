package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

var conn net.Conn
var sockReader *bufio.Reader

func main() {
	initFlags()

	switch {
	case setPassword:
		fmt.Printf("Enter default XAPI password: ")
		pass, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println("")
		if err != nil {
			fmt.Println("terminal.ReadPassword():", err)
			os.Exit(1)
		}
		connect()
		defer disconnect()
		runCommand("password " + string(pass))
	case len(os.Args) > 1:
		var cmdString string
		for _, arg := range os.Args[1:] {
			cmdString += fmt.Sprintf("\"%s\" ",
				strings.Replace(arg, "\"", "\\\"", -1))
		}
		connect()
		defer disconnect()
		runCommand(cmdString)
	default:
		reader := bufio.NewReader(os.Stdin)
		connect()
		defer disconnect()
		for {
			fmt.Print("xen> ")
			cmdString, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("")
				if err == io.EOF {
					os.Exit(0)
				}
				fmt.Println("error reading from stdin")
				os.Exit(1)
			}
			cmdString = strings.TrimSpace(cmdString)
			if cmdString == "exit" {
				break
			}
			runCommand(cmdString)
		}
	}
}

func runCommand(cmd string) {
	conn.Write([]byte(cmd + "\n"))
	for {
		buf, err := sockReader.ReadString('\n')
		if err != nil {
			fmt.Println("sockReader.ReadString():", err)
			os.Exit(1)
		}
		bufString := string(buf)
		if bufString == ".\n" {
			break
		}
		fmt.Print(bufString)
	}
}

func connect() {
	var err error
	conn, err = net.Dial("unix", config.socketPath)
	if err != nil {
		fmt.Println("net.Dial():", err)
		os.Exit(1)
	}
	sockReader = bufio.NewReader(conn)
}

func disconnect() {
	conn.Close()
}
