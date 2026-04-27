package main

import (
	"fmt"
	"log"
	"os"
)

type processLogger struct {
	prefix string
	file   *os.File
	logger *log.Logger
}

func (p processLogger) LogInfo(message string) {
	p.prefix = "INFO: "
	p.logger.Println(message)
}

func (p processLogger) LogError(message string) {
	p.prefix = "ERROR: "
	p.logger.Println(message)
}

func (p processLogger) Close() {
	p.file.Close()
}

func initProcessLogger(filePath string) processLogger {
	prefix := ""

	processLogFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("error intitating process log: %s\n", err.Error())
	}

	logger := log.New(processLogFile, prefix, log.Ldate|log.Ltime)

	return processLogger{
		file:   processLogFile,
		prefix: prefix,
		logger: logger,
	}

}
