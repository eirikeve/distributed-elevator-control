package elevhandler

import (
	"errors"
	"time"

	driver "../elevdriver"
	fsm "../elevfsm"
	timer "../elevtimer"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

var signalHandlerToStop chan bool
var fsmTimeoutSignal chan bool

func StartElevatorHandler() {
	log.Info("elevhandler StartElevatorHandler: Starting")
	// @TODO Finish initialization
	signalHandlerToStop = make(chan bool, 2)
	fsmTimeoutSignal = make(chan bool, 2)

	elevator, err := getElevatorBackup()
	if err != nil {
		elevator = nil
	}

	fsm.StartFSM(fsmTimeoutSignal, elevator)

	go elevatorHandlerInstance(signalHandlerToStop)

}

func StopElevatorHandler() {
	log.Info("elevhandler StopElevatorHandler: Stopping")
	//@BUG this does not send
	signalHandlerToStop <- true
	log.Info("elevhandler StopElevatorHandler: Stop signalled")
	return
}

func elevatorHandlerInstance(signalHandlerToStop <-chan bool) {
	log.Debug("elevhandler elevatorHandlerInstance: Starting")
	loopCount := 0

	motorDirectionInput := make(chan et.MotorDirection, 2)
	buttonLampInput := make(chan et.ButtonLamp, 12)
	floorIndicatorInput := make(chan int, 2)
	doorOpenLampInput := make(chan bool)
	buttonPressSensorOut := make(chan et.ButtonEvent, 12)
	floorSensorOut := make(chan int, 2)

	driver.StartDriver(
		et.NumFloors,
		motorDirectionInput,
		buttonLampInput,
		floorIndicatorInput,
		doorOpenLampInput,
		buttonPressSensorOut,
		floorSensorOut,
	)
	defer driver.StopDriver()

	//@TODO REMOVE. ONLY FOR DEBUGGING
	fsm.RegisterFloor(0)

	//@TODO program loop
	timer.StartDelayedFunction("ElevHandler Watchdog", time.Second*2, func() { panic("ElevHandler Watchdog: timeout") })
	defer timer.Stop("ElevHandler Watchdog")
	for {
		timer.Update("ElevHandler Watchdog", time.Second*2)
		select {
		// Handler Control
		case <-signalHandlerToStop:
			return
		// Reading Signals from FSM
		case f := <-floorSensorOut:
			if et.BOTTOMFLOOR <= f && f <= et.TOPFLOOR {
				fsm.RegisterFloor(f)
			}

		// Pushing Signals to FSM

		default:

		}
		time.Sleep(time.Millisecond * 100)
		loopCount++
		if loopCount > 20 {
			loopCount = 0
			log.Debug("elevhandler elevatorHandlerInstance: Running")
		}
	}
}

func getElevatorBackup() (*et.Elevator, error) {
	return nil, errors.New("Backup not yet implemented")
}
