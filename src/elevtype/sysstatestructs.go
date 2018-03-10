package elevtype

import "time"

type ElevState struct {
	// ID is the local IP of the computer corresponding to an ElevState
	ID string
	// LastUpdate is the last time this ElevState was updated
	LastUpdate time.Time

	E Elevator
	// ExtendedOrders is the Networkhandler's orders in the system corresponding to an ElevState
	CurrentOrders  [NumFloors][NumButtons]ElevOrder
	FinishedOrders []ElevOrder
}

type NetState struct {
	// LastUpdate is the last time this NetState was updated
	LastUpdate time.Time
	// P is the active Peers and Lost peers of this computer
	P PeerUpdate
}
