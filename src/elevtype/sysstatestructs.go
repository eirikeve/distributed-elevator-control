package elevtype

import def "../elevdef"

type ElevState struct {
	// ID is the local IP of the computer corresponding to an ElevState
	ID int32 `json: "ID"`
	// LastUpdate is the last time this ElevState was updated
	//LastUpdate  time.Time `json: "LastUpdate"`
	StartupTime int64    `json: "StartupTime"`
	E           Elevator `json: "E"`
	// ExtendedOrders is the Networkhandler's orders in the system corresponding to an ElevState
	CurrentOrders  [def.NumFloors][def.NumButtons]ElevOrder `json: "CurrentOrders"`
	FinishedOrders [12]ElevOrder                            `json: "FinishedOrders"`
}
