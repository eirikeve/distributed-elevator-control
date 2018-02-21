package elevlog

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

/*
TestLog tests basic log functionality, and shows how to use the log system
*/
func TestLog(*testing.T) {

	// initLog sets logrus' parameters.

	initLog(log.DebugLevel, false)
	var b = 5
	var a [10]byte
	a[2] = 255

	/*
		General log call:
		log.WithFields(log.Fields{
			"event": event,
			"topic": topic,
			"key": key,
		}).Debug("This is a message of level log.DebugLevel")

		Levels:
		log.Debug("Useful debugging information.")
		log.Info("Something noteworthy happened!")
		log.Warn("You should probably take a look at this.")
		log.Error("Something failed but I'm not quitting.")
		// Calls os.Exit(1) after logging
		log.Fatal("Bye.")
		// Calls panic() after logging
		log.Panic("I'm bailing.")
	*/

	log.WithFields(log.Fields{
		"variable b": 10,
		"variable a": 10,
	}).Debug("This will usually not show up. This is debug.")

	log.WithFields(log.Fields{
		"variable b": b,
		"variable a": a,
	}).Info("This was logged on purpose. This is info.")

	log.WithFields(log.Fields{
		"variable b": a,
		"variable a": b,
	}).Warning("This is a warning.")

	log.WithFields(log.Fields{
		"variable b": a,
		"variable a": b,
	}).Error("This is an error.")

	// This is a logger which always logs "Log level" and "a"
	contextLogger := log.WithFields(log.Fields{
		"Log level": log.InfoLevel,
		"a":         &a,
		"b":         b,
	})

	contextLogger.Info("Here's some info from a context logger")
	a[3] = 128
	b = 6
	contextLogger.Warning("Warning: We changed a and b, but only a is updated since it was by reference")

}
