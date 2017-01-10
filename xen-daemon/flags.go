package main

import (
	"flag"
	"log"
	"os"
	"path"
)

func initFlags() {

	// if we can determine HOME, use it to set a default base.dir
	var defaultBase string
	if homeDir := os.Getenv("HOME"); homeDir != "" {
		defaultBase = path.Join(homeDir, ".xen-cli")
	}

	// our flags to parse from commandline
	baseDir := flag.String("base.dir", defaultBase,
		"base dir used by xen-cli for its files")
	socketPath := flag.String("socket.path", "",
		"socket path used by xen-cli for interprocess communication")
	debug := flag.Bool("debug", false, "do not fork to background")
	logFile := flag.String("log.file", "", "file to use for logging")
	serversFile := flag.String("servers.file", "", "file for storing server list")
	flag.Parse()

	if baseDir != nil && *baseDir == "" {
		log.Fatalln("Unable to determine baseDir via arguments or HOME env var.")
	}

	config.baseDir = *baseDir

	// default socket path is based on determined base.dir, unless specified
	if socketPath != nil && *socketPath != "" {
		config.socketPath = *socketPath
	} else {
		config.socketPath = path.Join(config.baseDir, "xen-daemon.sock")
	}

	if logFile != nil && *logFile != "" {
		config.logFile = *logFile
	} else {
		config.logFile = path.Join(config.baseDir, "xen-daemon.log")
	}

	if serversFile != nil && *serversFile != "" {
		config.serversFile = *serversFile
	} else {
		config.serversFile = path.Join(config.baseDir, "servers.json")
	}

	if debug != nil {
		config.debug = *debug
	}

}
