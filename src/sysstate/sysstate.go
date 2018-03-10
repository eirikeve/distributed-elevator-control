package sysstate

import (
	"time"

	et "../elevtype"
)

type ElevState struct {
	// ID is the local IP of the computer corresponding to an ElevState
	ID string
	// LastUpdate is the last time this ElevState was updated
	LastUpdate time.Time

	ElevatorFloor      int
	ElevatorErrorState et.ElevatorErrorState
	// ExtendedOrders is the Networkhandler's orders in the system corresponding to an ElevState
	CurrentOrders  [et.NumFloors][et.NumButtons]et.ElevOrder
	FinishedOrders []et.ElevOrder
}

type NetState struct {
	// LastUpdate is the last time this NetState was updated
	LastUpdate time.Time
	// P is the active Peers and Lost peers of this computer
	P et.PeerUpdate
}

var localIP string
var systems map[string]ElevState