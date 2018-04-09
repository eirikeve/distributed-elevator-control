package elevtype

import "time"

type ElevState struct {
	// ID is the local IP of the computer corresponding to an ElevState
	ID string `json: "ID"`
	// LastUpdate is the last time this ElevState was updated
	LastUpdate  time.Time `json: "LastUpdate"`
	StartupTime time.Time `json: "StartupTime"`

	E Elevator `json: "E"`
	// ExtendedOrders is the Networkhandler's orders in the system corresponding to an ElevState
	CurrentOrders  [NumFloors][NumButtons]ElevOrder `json: "CurrentOrders"`
	FinishedOrders []ElevOrder                      `json: "FinishedOrders"`
}

type NetState struct {
	// LastUpdate is the last time this NetState was updated
	LastUpdate time.Time
	// P is the active Peers and Lost peers of this computer
	P PeerUpdate
}
