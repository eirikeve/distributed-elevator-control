package elevfsm

import (
	"../elevtype"
)

var elevator elevtype.Elevator
var lastRegisteredFloor int

// Functions for running the local elevator
/*
func GetMovementDirection() elevtype.MotorDirection {
	return OrderLogicGetMovementDirection(elevator)
}
*/
func GetPanelLights() []elevtype.ButtonLamp {
	var v []elevtype.ButtonLamp

	return v
}

func RegisterFloor() {

}

func GetState() elevtype.ElevatorState {
	v := elevtype.Idle

	return v

}

func setState(state elevtype.ElevatorState) {

}

func setDir(dir elevtype.MotorDirection) {

}

func unload() {

}

func idle() {

}

// Functions used when interfacing with elevNetworkHandler
func PushRequestToQueue(order elevtype.ElevOrder) {

}
func RemRequestFromQueue(order elevtype.ElevOrder) {

}

func GetElevator() elevtype.Elevator {
	var v elevtype.Elevator

	return v

}
