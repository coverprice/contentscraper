package toolbox

import (
	"flag"
	"os"
	"runtime"

	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

var (
	logLevelFlag string
)

func init() {
	flag.StringVar(&logLevelFlag, "log-level", "INFO", "One of DEBUG, INFO, WARN, ERROR, FATAL, PANIC")
}

func initLogOutput(logFilename string) {
	if logFilename != "" {
		// If the file doesn't exist, create it, or append to the file
		filehandle, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		log.SetFormatter(&log.TextFormatter{DisableColors: true})
		log.SetOutput(filehandle)
		return
	}

	// Log to stdout
	if runtime.GOOS == "windows" {
		// This fixes up the terminal colors in Windows
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(colorable.NewColorableStdout())
	}
	return
}

// InitLogging initializes the logging for the application.
// logFilename is an optional location for the file. If not specified, the logs go to stdout.
func InitLogging(logFilename string) {
	initLogOutput(logFilename)

	logLevel, err := log.ParseLevel(logLevelFlag)
	if err != nil {
		log.Fatalf("Invalid log level: '%s'", logLevel)
	}
	log.SetLevel(logLevel)
}
