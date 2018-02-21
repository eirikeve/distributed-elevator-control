package elevutil

/*go get github.com/op/go-logging


log "github.com/sirupsen/logrus"*/

import (
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
 * Setup for logrus
 * @arg level: is the level of severity which is registered in the log. Goes from 0 (PanicLevel; log ~nothing) to 5 (DebugLevel; log everything)
 * @arg toFile: If true, log to a file with filename <current UNIX timestamp>.elevlog, if false, log to stdout
 */
func initLog(level log.Level, toFile bool) {
	if toFile {
		// Set name of current log output to the current UNIX timestamp
		filename := strconv.FormatInt(time.Now().Unix(), 10) + ".elevlog"
		logFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			// Could not open file
			log.WithFields(log.Fields{
				"Error": err.Error(),
			}).Error("initLog: Could not open log output file. Defaulting to bash output.")
		} else {
			log.SetOutput(logFile)
		}
	}
	if log.PanicLevel <= level && level <= log.DebugLevel {
		log.SetLevel(level)
	} else {
		// Invalid level input
		log.SetLevel(log.InfoLevel)
		log.WithFields(log.Fields{
			"Log level": level,
		}).Warning("initLog: Invalid log level input. Defaulting to log.InfoLevel")
	}
	// @ Todo make some way to close the log file

}
