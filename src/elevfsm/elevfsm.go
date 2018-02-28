package elevfsm

import et "../elevtype"

var elevator et.Elevator
var lastRegisteredFloor int

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
		return
	}
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
	// log here
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
	var v elevtype.Elevator

	return v

}
