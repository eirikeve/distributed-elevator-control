package main

import (
	"fmt"
	"time"

	ed "./elevdriver"
	eh "./elevhandler"
	et "./elevtype"
	nh "./nethandler"
	sb "./sysbackup"
	ss "./sysstate"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Recover from a panic from https://github.com/golang/go/wiki/PanicAndRecover

	/*defer func() {
		if r := recover(); r != nil {
			var ok bool
			var err error
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()*/

	parseCmdLineArgs()
	setupLog()
	run()
}

func run() {

	defer recoverIfPanic()
	stopRunning := make(chan bool, 2)
	ed.StartStopButtonService(stopRunning)
	defer ed.StopStopButtonService()

	systemStates, _ := sb.Recover(time.Now().Add(et.BackupRecoverInterval))
	ss.SetSystemsStates(systemStates)

	log.WithField("states", systemStates).Debug("main run: Setup sysstates")

	ordersDelegatedFromNetwork := make(chan et.GeneralOrder, 12)
	buttonPressesToNetwork := make(chan et.ButtonEvent, 12)
	elevStateToNetwork := make(chan et.Elevator, 12)

	eh.StartElevatorHandler(ordersDelegatedFromNetwork,
		buttonPressesToNetwork,
		elevStateToNetwork)
	nh.StartNetHandler(ordersDelegatedFromNetwork,
		buttonPressesToNetwork,
		elevStateToNetwork)

	var running = true
	for running == true {
		select {
		case <-stopRunning:
			log.Info("main run: Received shutdown signal")
			eh.StopElevatorHandler()
			nh.StopNetHandler()
			running = false
		default:
			time.Sleep(time.Millisecond)
		}

	}

	log.Info("main run: Exiting in 3 seconds")
	time.Sleep(time.Second * 1)
	log.Info("main run: Exiting in 2 seconds")
	time.Sleep(time.Second * 1)
	log.Info("main run: Exiting in 1 second")
	time.Sleep(time.Second * 1)
}

func recoverIfPanic() {
	if r := recover(); r != nil {
		var ok bool
		var err error
		err, ok = r.(error)
		if !ok {
			err = fmt.Errorf("pkg: %v", r)
			log.WithError(err).Error("main recoverIfPanic: Could not recover")
			//os.Exit(1)
			run()
		}
		log.WithError(err).Warning("main recoverIfPanic: Recovered from panic")
		run()
	}
}
