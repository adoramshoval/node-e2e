package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
    // Set JSON formatter for log output
    log.SetFormatter(&log.JSONFormatter{
        TimestampFormat: "2006-01-02T15:04:05Z07:00",
        PrettyPrint:     false,
    })

    // Log to stdout
    log.SetOutput(os.Stdout)

    // Enable logging of the file and line number
    log.SetReportCaller(true)

    // Set log level
    log.SetLevel(log.InfoLevel)
}

// Export the logger instance
var Logger = log.StandardLogger()

