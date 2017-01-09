package main

import (
	"io"
	"log"
	"os"
	"sync"
)

// LogWriter is a bufio.Writer for our logging
type LogWriter struct {
	file     *os.File
	filename string
	lock     sync.Mutex
}

var logWriter *LogWriter

// NewLogWriter instantiates a new logwriter
func initLogger() {
	if config.debug {
		return
	}
	logWriter = &LogWriter{}
	logWriter.filename = config.logFile
	logWriter.refresh()
	log.SetOutput(io.Writer(logWriter))
}

func (w *LogWriter) refresh() {
	var err error

	w.lock.Lock()
	defer w.lock.Unlock()

	if w.file != nil {
		w.file.Close()
	}

	w.file, err = os.OpenFile(config.logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		// some good irony in logging an error when logging is broken
		log.Fatalln("os.OpenFile():", err)
	}

}

func (w *LogWriter) Write(output []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.file.Write(output)
}
