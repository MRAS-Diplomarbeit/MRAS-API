package logger

import (
	"io"
	"log"
	"os"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func init() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	logmw := io.MultiWriter(os.Stdout, file)
	logerrmw := io.MultiWriter(os.Stderr, file)
	InfoLogger = log.New(logmw, "INFO: ", log.LstdFlags)
	WarningLogger = log.New(logmw, "WARNING: ", log.LstdFlags|log.Lshortfile)
	ErrorLogger = log.New(logerrmw, "ERROR: ", log.LstdFlags|log.Lshortfile)
}
