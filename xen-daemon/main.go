package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	checkSecure()

	initFlags()
	validateConfig()

	if !config.debug {
		daemonize()
	}

	initSignalHUP()

	log.Printf("%20s: %s\n", "base.dir", config.baseDir)
	log.Printf("%20s: %s\n", "socket.path", config.socketPath)

	xenData.loadXenServers()
	xenData.launchPollers()

	server := NewServer()
	server.listen()

}

func daemonize() {

	//	time.Sleep(1 * time.Second)

	if os.Getppid() != 1 {
		// i am an adult?
		binary, err := exec.LookPath(os.Args[0])
		if err != nil {
			log.Fatalln("exec.LookPath():", err)
		}
		_, err = os.StartProcess(binary, os.Args, &os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
		if err != nil {
			log.Fatalln("os.StartProcess():", err)
		}
		os.Exit(0)
	} else {
		// I am a child?
		_, err := syscall.Setsid()
		if err != nil {
			log.Fatalln("syscall.Setsid():", err)
		}
		file, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
		if err != nil {
			log.Fatalln("os.OpenFile(\"/dev/null\", ...):", err)
		}
		devNullFD := int(file.Fd())
		syscall.Dup2(devNullFD, int(os.Stdin.Fd()))
		syscall.Dup2(devNullFD, int(os.Stdout.Fd()))
		syscall.Dup2(devNullFD, int(os.Stdout.Fd()))
		file.Close()
		initLogger()
	}
}
