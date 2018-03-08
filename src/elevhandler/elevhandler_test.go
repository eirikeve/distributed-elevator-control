package elevhandler

import (
	"testing"
	"time"

	"../elevlog"
	log "github.com/sirupsen/logrus"
)

func TestElevHandler(t *testing.T) {
	elevlog.InitLog(log.DebugLevel, false)

	log.Info("elevhandler TestElevHandler: Starting test")
	StartElevatorHandler()
	time.Sleep(time.Second * 20)
	StopElevatorHandler()
	time.Sleep(time.Second)

	log.Info("elevhandler TestElevHandler: Done")
}

func TestElevHandlerPanic(t *testing.T) {
	// @BUG this doesn't work correctly...
	// Add time.Sleep(time.Second*10) in elevhandler instance loop, then run this test
	elevlog.InitLog(log.DebugLevel, false)
	defer func() {
		if r := recover(); r != nil {
			log.Info("elevhandler TestElevHandlerPanic: Recovered")
			StopElevatorHandler()
			time.Sleep(time.Second)
			// re-init here
		}
	}()

	log.Info("elevhandler TestElevHandlerPanic: Starting test")
	StartElevatorHandler()
	time.Sleep(time.Second * 10)
	defer func() {
		StopElevatorHandler()
		time.Sleep(time.Second)
	}()

	log.Info("elevhandler TestElevHandler: Done")
}
