package logger

import (
	"io"
	"log"
	"os"
)

// Global logger variable
var Log *log.Logger

// Setup initializes the logger
func Setup(logDir string) error {
	// Create the specified log directory if it doesn't exist
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return err
	}

	logFileName := logDir + "/log.txt"

	// Open the log file for appending
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	// Create a multi-writer that writes to both the log file and stdout
	multiWriter := io.MultiWriter(logFile, os.Stdout)

	// Set log output to the multi-writer
	Log = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}
