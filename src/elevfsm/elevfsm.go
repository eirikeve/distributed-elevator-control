package elevfsm

import (
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

var elevator et.Elevator
var lastRegisteredFloor int
var timerSignal chan<- bool

func Initialize(timerSignalOutput chan<- bool, e *Elev) {
	timerSignal = timerSignalOutput
	if e == nil {
		elevator = et.Elevator{
			Floor:               4,
			MovementDirection:   et.MD_Stop,
			MovDirFromLastFloor: et.MD_Up,
			State:               et.Initializing,
			ErrorState:          et.FullFunctionality}
		log.WithField("elevator", elevator).Debug("elevfsm Initialize: No ref, reinitialized elevator")
	} else {
		elevator = &e
		log.WithField("elevator", elevator).Debug("elevfsm Initialize: Initialized elevator from ref")
	}
}

// Functions for running the local elevator
/*
func GetMovementDirection() elevtype.MotorDirection {
	return OrderLogicGetMovementDirection(elevator)
}
*/
func GetPanelLights() [et.NumFloors * et.NumButtons]et.ButtonLamp {
	var v [et.NumFloors * et.NumButtons]et.ButtonLamp
	for f := 0; i < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			// @TODO If order is nil this will not work
			// make a get function or something else that returns the value.
			v[f*et.NumButtons+b].Floor = et.ButtonLamp{Floor: f, Button: b, Value: (elevator.Orders[f][b].Status == Accepted)}
		}
	}
	return v
}

func RegisterFloor(floor int) {
	if !(isValidFloor(floor)) {
		log.WithField("floor", floor).Error("elevfsm RegisterFloor: Invalid input")
		return
	}
	log.WithField("floor", floor).Debug("elevfsm RegisterFloor:")
	updateFloor(floor)
	if OrderLogicCheckShouldStopAtFloor(elevator) {
		// need a good way to make sure we start moving after some time
		// if we have new orders
	}

}

func GetState() et.ElevatorState {
	return elevator.State
}

func setState(state et.ElevatorState) {

}

func setDir(dir et.MotorDirection) {

}

func unload() {

}

func idle() {

}

func updateFloor(floor int) {
	lastRegisteredFloor = elevator.Floor
	elevator.Floor = floor
}

func isValidFloor(floor int) bool {
	return (0 <= floor && floor < et.NumFloors)
}

// Functions used when interfacing with elevNetworkHandler
func PushRequestToQueue(order et.ElevOrder) {

}
func RemRequestFromQueue(order et.ElevOrder) {

}

func GetElevator() et.Elevator {
	return elevator
}

func ForceState(state et.ElevatorState) {
	log.WithField("state", state).Warning("elevfsm ForceState: Forced state change")
	elevator.State = state
}
