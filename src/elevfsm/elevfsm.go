package elevfsm

import (
	"time"

	timer "../elevtimer"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

var elevator et.Elevator
var lastRegisteredFloor int
var doorTimeoutSignalOutput chan bool
var errorCount int

const initFailTimeout time.Duration = 5 * time.Second

func Initialize(doorTimeoutSignal chan bool, e *et.Elevator) {
	doorTimeoutSignalOutput = doorTimeoutSignal
	if e == nil {
		elevator = et.Elevator{
			Floor:               4,
			MovementDirection:   et.MD_Stop,
			MovDirFromLastFloor: et.MD_Up,
			State:               et.Initializing,
			ErrorState:          et.FullFunctionality}
		log.WithField("elevator", elevator).Debug("elevfsm Initialize: No ref, reinitialized elevator")
	} else {
		elevator := &e
		log.WithField("elevator", elevator).Debug("elevfsm Initialize: Initialized elevator from ref")
	}
	timer.Start("Initialization", initFailTimeout, doorTimeoutSignalOutput)
	// Start by moving downwards
	initialize()
}

// Functions for running the local elevator
/*
func GetMovementDirection() elevtype.MotorDirection {
	return OrderLogicGetMovementDirection(elevator)
}
*/
func GetPanelLights() [et.NumFloors * et.NumButtons]et.ButtonLamp {
	var lights [et.NumFloors * et.NumButtons]et.ButtonLamp
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			// @TODO If order is nil this will not work //@BUG
			// make a get function or something else that returns the value.
			lights[f*et.NumButtons+b] = et.ButtonLamp{Floor: f, Button: et.ButtonType(b), Value: (elevator.Orders[f][b].Status == et.Accepted)}
		}
	}
	return lights
}

func GetDoorOpenLight() bool {
	return (elevator.State == et.Unloading)
}

func GetState() et.ElevatorState {
	return elevator.State
}

func GetMotorDir() et.MotorDirection {
	return elevator.MovementDirection
}

func RegisterFloor(floor int) {
	if !(isValidFloor(floor)) {
		log.WithField("floor", floor).Error("elevfsm RegisterFloor: Ignoring invalid input")
		return
	}
	log.WithField("floor", floor).Debug("elevfsm RegisterFloor:")
	updateFloor(floor)

	switch elevator.State {
	case et.Initializing:
		idle()
	case et.Moving:
		if OrderLogicCheckShouldStopAtFloor(elevator) {
			unload()
		} else if floor == et.BOTTOMFLOOR || floor == et.TOPFLOOR {
			idle()
		}
	case et.Unloading:
		if OrderLogicCheckShouldStopAtFloor(elevator) { //@TODO is this sufficient?
			unload()
		}
		return
	case et.Idle:
		if OrderLogicCheckShouldStopAtFloor(elevator) {
			unload()
		} else {
			newDirMovement := OrderLogicGetMovementDirection(elevator)
			if newDirMovement != et.MD_Stop {
				elevator.MovDirFromLastFloor = newDirMovement
				move(newDirMovement)
			}
		}

	default:
		log.WithField("state", elevator.State).Error("elevfsm RegisterFloor: Unhandled state")
	}

}

func RegisterTimerTimeout() {
	switch elevator.State {
	case et.Unloading:
		idle()
	case et.Initializing:
		switch elevator.MovementDirection {
		case et.MD_Down: // normal initialization
			setDir(et.MD_Up)
			errorCount++
			timer.Start("InitializationRetry", initFailTimeout, doorTimeoutSignalOutput)
			log.Warning("elevfsm RegisterTimerTimeout: Retrying Init, moving up")

		case et.MD_Up:
			setDir(et.MD_Stop)
			errorCount++
			log.Error("elevfsm RegisterTimerTimeout: Error, could not init, restarting")
			//@TODO restart here
		}
		return
	case et.Idle:
		fallthrough
	case et.Moving:
		fallthrough
	default:
		log.WithField("state", elevator.State).Error("elevfsm RegisterTimerTimeout: Unhandled state")
		errorCount++ //@TODO maybe change to an internal variable in elevator?
	}
}

// Functions used when interfacing with elevNetworkHandler
func PushRequestToQueue(order et.ElevOrder) {
	// Need some logic here! @TODO
}
func RemRequestFromQueue(order et.ElevOrder) {
	// Need some logic here! @TODO
}

func GetElevator() et.Elevator {
	return elevator
}

func ForceState(state et.ElevatorState) {
	log.WithField("state", state).Warning("elevfsm ForceState: Forced state change")
	elevator.State = state
}

func setState(state et.ElevatorState) {
	log.WithField("state", state).Debug("elevfsm setState: State change")
	elevator.State = state
}

func setDir(dir et.MotorDirection) {
	log.WithField("movementDir", dir).Debug("elevfsm setDir: Dir change")
	elevator.MovementDirection = dir
}

func initialize() {
	log.Debug("elevfsm initialize: Initializing")
	setState(et.Initializing)
	setDir(et.MD_Down)
}

func unload() {
	log.WithField("floor", elevator.Floor).Debug("elevfsm unload: Unloading")
	timer.Start("UnloadTimer", time.Second*3, doorTimeoutSignalOutput)
	setState(et.Unloading)
	setDir(et.MD_Stop)
	newMovementDir := OrderLogicGetMovementDirection(elevator)
	// @TODO maybe rewrite this so that we don't need to reassign elevator.
	elevator = OrderLogicClearRequestsOnCurrentFloor(elevator, newMovementDir)
}

func idle() {
	log.WithField("floor", elevator.Floor).Debug("elevfsm idle: Idling")
	setDir(et.MD_Stop)
	setState(et.Idle)
}

func move(dir et.MotorDirection) {
	if dir == et.MD_Stop {
		idle()
	} else {
		log.WithFields(log.Fields{"floor": elevator.Floor, "dir": dir}).Debug("elevfsm move: Moving")
		setState(et.Moving)
		setDir(dir)
	}

}

func updateFloor(floor int) {
	lastRegisteredFloor = elevator.Floor
	elevator.Floor = floor
}

func isValidFloor(floor int) bool {
	return (0 <= floor && floor < et.NumFloors)
}
