package elevfsm

import (
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

var elevator et.Elevator
var lastRegisteredFloor int
var timerSignal chan<- bool

func Initialize(timerSignalOutput chan<- bool, e * Elev) {
	timerSignal = timerSignalOutput;
	if (e == nil) {
		elevator = et.Elevator{
			Floor:              4,
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
func GetPanelLights() []et.ButtonLamp {
	var v []et.ButtonLamp

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
	v := et.Idle

	return v

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
	var v et.Elevator

	return v

}

isTimeout()
