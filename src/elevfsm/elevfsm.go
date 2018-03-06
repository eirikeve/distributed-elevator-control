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

	// Start by moving downwards
	setDir(et.MD_Down)
	setState(et.Initializing)
	//@TODO possibly set timer here

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
			// @TODO If order is nil this will not work
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
		log.WithField("floor", floor).Error("elevfsm RegisterFloor: Invalid input")
		return
	}
	log.WithField("floor", floor).Debug("elevfsm RegisterFloor:")
	updateFloor(floor)

	switch elevator.State {
	case et.Initializing:
		//@TODO initialization logic here ? Or is idle() sufficient?
		idle()
	case et.Moving:
		if OrderLogicCheckShouldStopAtFloor(elevator) {
			// Will signal timeout to doorTimeoutSignalOutput
			unload()
		} else if floor == et.BOTTOMFLOOR || floor == et.TOPFLOOR {
			idle()
		}
	case et.Unloading:
		//@TODO check if we can/should clear orders again
		//newDirMovement := OrderLogicGetMovementDirection(elevator)
		return

	case et.Idle:
		//@TODO check if we can/should clear orders again
		newDirMovement := OrderLogicGetMovementDirection(elevator)
		if newDirMovement != et.MD_Stop {
			move(newDirMovement)
		}
	default:
		log.WithField("state", elevator.State).Error("elevfsm RegisterFloor: Unhandled state")
	}

}

func RegisterTimerTimeout() {
	switch elevator.State {
	case et.Unloading:
		newMovementDir := OrderLogicGetMovementDirection(elevator)
		setDir(newMovementDir)
		if newMovementDir != et.MD_Stop {
			elevator.MovDirFromLastFloor = elevator.MovementDirection
		}
	case et.Initializing:
		// @TODO maybe this can be used for stopping the FSM if we fail to initialize
		return
	case et.Idle:
	case et.Moving:
	default:
		log.WithField("state", elevator.State).Error("elevfsm RegisterTimerTimeout: Unhandled state")
	}
}

// Functions used when interfacing with elevNetworkHandler
func PushRequestToQueue(order et.ElevOrder) {
	// Need some logic here!
}
func RemRequestFromQueue(order et.ElevOrder) {
	// Need some logic here!
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
	setState(et.Initializing)
	setDir(et.MD_Down)
}

func unload() {
	timer.Start("stopAtFloorTimer", time.Second*3, doorTimeoutSignalOutput)
	setState(et.Unloading)
	setDir(et.MD_Stop)
	newMovementDir := OrderLogicGetMovementDirection(elevator)
	OrderLogicClearRequestsOnCurrentFloor(&elevator, newMovementDir)
}

func idle() {
	setDir(et.MD_Stop)
	setState(et.Idle)
}

func move(dir et.MotorDirection) {
	if dir == et.MD_Stop {
		idle()
	} else {
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
