package main

import (
	"fmt"
	"log"
	"os"
)

type processLogger struct {
	file   *os.File
	logger *log.Logger
}

type logError struct {
	errMessage string
	logMessage string
}

func (l *logError) Error() string {
	return l.errMessage
}

func (l *logError) LogError() string {
	return l.logMessage
}

func (p processLogger) LogInfo(message string) {
	prefix := "INFO: "
	p.logger.Println(prefix + message)
}

func (p processLogger) LogError(message string) {
	prefix := "ERROR: "
	p.logger.Println(prefix + message)
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
		logger: logger,
	}

}
