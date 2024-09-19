package utils

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
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

func SuccessfullyRetrievedPodLog(namespace string, podName string) {
	Logger.WithFields(log.Fields{
		"namespace": namespace,
		"podName":   podName,
	}).Errorf("Successfully retrieved pod")
}

func FailedToRetrievePodErrorLog(namespace string, podName string, err error) {
	Logger.WithFields(log.Fields{
		"namespace": namespace,
		"podName":   podName,
		"error":     err,
	}).Errorf("Failed to retrieve pod: %v", err)
}

func UnexpectedTypeErrorLog(r string, w string) {
	Logger.WithFields(log.Fields{
		"CorrectType": r,
		"ActualType":  w,
	}).Errorf("Unexpected type %s", w)
}

// Used for errors as they are not usually capitalized
func Uncapitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(string(s[0])) + s[1:]
}
