package main

import (
	"fmt"
	"time"

	ed "./elevdriver"
	eh "./elevhandler"
	et "./elevtype"
	nh "./nethandler"
	re "./recover"
	setup "./setup"
	sb "./sysbackup"
	ss "./sysstate"
	log "github.com/sirupsen/logrus"
)

func main() {
	setup.ParseCmdLineArgs()
	setup.SetupLog()
	run()
}

func run() {
	timeOfStart := time.Now().Unix()
	defer recoverIfPanic()
	stopRunning := make(chan bool, 2)

	re.StartSurveillanceOfPrimary()
	ss.InitSysState()
	systemStates, _ := sb.Recover(time.Now().Add(et.BackupRecoverInterval))

	log.WithField("states", systemStates).Debug("main run: Setup sysstates")
	ss.SetSystemsStates(systemStates)
	localSys := ss.GetLocalSystem()
	elevator := sb.GetBackupElev(timeOfStart, localSys)
	log.WithField("states", ss.GetLocalSystem().E).Debug("main run: Done w/ setup of sysstates")

	orderQueueFromNethandler := make(chan [et.NumFloors][et.NumButtons]et.SimpleOrder, 12)
	buttonLightsFromNethandler := make(chan et.ButtonLamp, 12)
	buttonPressesToNethandler := make(chan et.ButtonEvent, 12)
	elevStateToNethandler := make(chan et.Elevator, 12)

	eh.StartElevatorHandler(orderQueueFromNethandler,
		buttonLightsFromNethandler,
		buttonPressesToNethandler,
		elevStateToNethandler,
		elevator,
	)
	nh.StartNetHandler(orderQueueFromNethandler,
		buttonLightsFromNethandler,
		buttonPressesToNethandler,
		elevStateToNethandler)

	ed.StartStopButtonService(stopRunning)
	defer ed.StopStopButtonService()

	var running = true
	for running == true {
		select {
		//case <-stopRunning:
		//	log.Info("main run: Received shutdown signal")
		//	eh.StopElevatorHandler()
		//	nh.StopNetHandler()
		//	running = false
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
