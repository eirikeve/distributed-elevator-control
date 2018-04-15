package sysstate

import (
	"time"

	network "../elevnetwork"
	locIP "../elevnetwork/localip"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

/*

sysstate.go contains methods for initializing, setting, and getting the system's state.
The system state is the overview of all orders (finished and current),
along with the current physical position, dir, etc. of the local elevator.

*/

////////////////////////////////
// Module variables
////////////////////////////////

// This system's unique ID, used as the local key in the systems map. All values except 0 are used as IDs.
var LocalID int32

// Contains the local ElevState (complete overview of the sys) along with the most recent updates from other systems.
var systems = make(map[int32]et.ElevState)
var initialized = false

////////////////////////////////
// Interface
////////////////////////////////

/*initSysState performs a basic setup of systems and LocalID.
 * It's automatically called the first time we call any functions in the sysstate module interface.
 */
func InitSysState() {

	LocalID, _ = locIP.LocalID()

	_, localSysExists := systems[LocalID]

	if !localSysExists {
		newElevState := et.ElevState{ID: LocalID, E: et.EmptyElevator(), StartupTime: time.Now().Unix()}

		systems[LocalID] = newElevState
	}

	initialized = true
	log.WithField("localID", LocalID).Info("sysstate: Initialized")

}

/*SetSystemsStates assigns a new value to the systems variable.
 * @arg sys: Slice of ElevState variables, all assumed to have unique IDs (otherwise some will be overwritten when assigning)
 */
func SetSystemsStates(sys []et.ElevState) {
	for _, system := range sys {
		if system.ID != 0 {
			systems[system.ID] = system
		}
	}

}

/*GetSystemsStates returns the ElevStates in the systems variable.
 * @return: ElevState of local system, and all other elevators we have communicated with
 */
func GetSystemsStates() []et.ElevState {

	var sys []et.ElevState
	for _, system := range systems {
		sys = append(sys, system)
	}
	return sys
}

/*GetActiveSystemsStates returns the currently active ElevStates in the systems variable.
 * This is determined from checking heartbeat.
 * @return: ElevState of active systems
 */
func GetActiveSystemsStates() []et.ElevState {

	activeSys := network.GetSystemsInNetwork()
	var sys []et.ElevState
	for _, system := range systems {
		if contains(activeSys, system.ID) {
			sys = append(sys, system)
		}

	}
	return sys
}

/*GetLocalSystem returns the ElevState representing the local system.
 * @return: ElevState of local system (with ID == LocalID)
 */
func GetLocalSystem() et.ElevState {
	localSys := systems[LocalID]
	return localSys
}

/*GetLocalSystemQueue returns the queue.
 * It's used for sending the queue to elevatorhandler.
 * @return: 2d array where each position is an order (they can be empty).
 */
func GetLocalSystemQueue() [et.NumFloors][et.NumButtons]et.SimpleOrder {

	var orders [et.NumFloors][et.NumButtons]et.SimpleOrder
	// Get new orders to delegate
	s, _ := systems[LocalID]
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if s.CurrentOrders[f][b].IsAccepted() && s.CurrentOrders[f][b].Assignee == LocalID {
				o := s.CurrentOrders[f][b].ToSimpleOrder()
				orders[f][b] = o
			}
		}
	}
	return orders
}

/*GetSystemElevators returns all the elevators from the systems variable
 * It's used for choosing a system to delegate an order to in orderlogic.go
 * @return: slice of elevators, one for each ElevState in systems.
 */
func GetSystemElevators() []et.Elevator {

	var elevList []et.Elevator
	for _, system := range systems {
		elevList = append(elevList, system.E)
	}
	return elevList
}

/*GetPanelLights is used to set the lights of the elevator buttons.
 * @return: 2d array where each position is an ButtonLamp value (floor, button, on/off).
 */
func GetPanelLights() [et.NumFloors][et.NumButtons]et.ButtonLamp {

	s, _ := systems[LocalID]
	var lights [et.NumFloors][et.NumButtons]et.ButtonLamp
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			lights[f][b] = et.ButtonLamp{Floor: f, Button: et.ButtonType(b), Value: s.CurrentOrders[f][b].IsAccepted()}
		}
	}
	return lights
}
