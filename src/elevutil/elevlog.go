package elevutil

/*go get github.com/op/go-logging


log "github.com/sirupsen/logrus"*/

import (
	/*"os"
	"strconv"
	"time"*/

	log "github.com/sirupsen/logrus"
)

func main() {
	initLog()
}

func initLog() {
	// Set name of current log output to the current UNIX timestamp
	// Can also write to syslog if we want, but doubt it is neecessary
	/*filename := strconv.FormatInt(time.Now().Unix(), 10) + ".elevlog"
	logFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		println("Could not open log file destination file")
	}
	log.SetOutput(logFile)
	// Only log info or higher to file*/

	log.SetLevel(log.DebugLevel)

	/*
			General log call:
			log.WithFields(log.Fields{
		  		"event": event,
		  		"topic": topic,
		  		"key": key,
			}).Fatal("Failed to send event")

			Levels:

			log.Debug("Useful debugging information.")
			log.Info("Something noteworthy happened!")
			log.Warn("You should probably take a look at this.")
			log.Error("Something failed but I'm not quitting.")
			// Calls os.Exit(1) after logging
			log.Fatal("Bye.")
			// Calls panic() after logging
			log.Panic("I'm bailing.")


			Important concent: contextLogger
	*/

	//
}
