package main

import (
	"log"
	"os"
)

type configClass struct {
	baseDir    string
	debug      bool
	socketPath string
	logFile    string
}

var config = &configClass{}

func validateConfig() {
	baseDirStat, err := os.Stat(config.baseDir)
	switch {
	case os.IsNotExist(err):
		log.Printf("creating %s", config.baseDir)
		err = os.Mkdir(config.baseDir, os.FileMode(0700))
		if err != nil {
			log.Fatalf("error creating %s: %s", config.baseDir, err)
		}
	case !baseDirStat.IsDir():
		log.Fatalf("base dir %s is not a directory", config.baseDir)
	}
}
