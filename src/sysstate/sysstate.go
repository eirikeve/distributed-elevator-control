package sysstate

import (
	"time"

	network "../elevnetwork"
	locIP "../elevnetwork/localip"
	et "../elevtype"
)

var LocalIP string
var initialized = false

//var systems map[string]et.ElevState
var systems = make(map[string]et.ElevState)
var netstate et.NetState
var acksForBroadcasting []et.AckNackMsg

/*
 * initSysState is automatically called the first time we set the local systems
 */
func initSysState() {
<<<<<<< HEAD
	exsistsInSystems := false
	LocalIP, _ = locIP.LocalIP()
	// Check that systems contains the elevator this program is running at. If not, insert this elevator system.
	for _, element := range systems {
		if element.ID == LocalIP {
			exsistsInSystems = true
			break
		}
=======

	if initialized {
		return
>>>>>>> d597c8aac8f40ee07545d414c695c77d43dac16f
	}

	LocalIP, _ = locIP.LocalIP()

	_, localSysExists := systems[LocalIP]

	if !localSysExists {
		newElevState := et.ElevState{ID: LocalIP, E: et.EmptyElevator(), StartupTime: time.Now()}
		systems[LocalIP] = newElevState
	}

	initialized = true

}

<<<<<<< HEAD
/*SetSystems assigns a slice to systems.
 *
 */
func SetSystems(sys []et.ElevState) {
	systems = sys
	initSysState()
}

/*GetSystems is a get-function for systems
 *
 */
func GetSystems() []et.ElevState { return systems }

/*GetSystemElevators is a get-function for the elevators in systems
 *
 */
func GetSystemElevators() []et.Elevator {
	var elevList []et.Elevator
	for _, elev := range systems {
		elevList = append(elevList, elev.E)
=======
func SetSystemsStates(sys []et.ElevState) {
	systems = make(map[string]et.ElevState)
	for _, system := range sys {
		systems[system.ID] = system
	}
	// Initialize after the assignment since this guarantees local system being in systems after func call
	if !initialized {
		initSysState()
	}
}

func GetSystemsStates() []et.ElevState {
	var sys []et.ElevState
	for _, system := range systems {
		sys = append(sys, system)
	}
	return sys
}

func GetActiveSystemsStates() []et.ElevState {
	activeSys := network.GetSystemsInNetwork()
	var sys []et.ElevState
	for _, system := range systems {
		if contains(activeSys, system.ID) {
			sys = append(sys, system)
		}

>>>>>>> d597c8aac8f40ee07545d414c695c77d43dac16f
	}
	return sys
}

func GetLocalSystem() et.ElevState {
	return systems[LocalIP]
}

<<<<<<< HEAD

/*UpdateLocalElevator updates the stored local elevator system with its current state.append
 * The argument e is the newest update from elevhandler, i.e. the current state of the local FSM
 */
func UpdateLocalElevator(e *et.Elevator) {
	exsistsInSystems := false
	for index, element := range systems {
		if element.ID == LocalIP {
			exsistsInSystems = true
			systems[index].E = *e
			break
=======
func GetUnsentLocalSystemOrders() []et.ElevOrder {
	var orders []et.ElevOrder
	s, _ := systems[LocalIP]
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if s.CurrentOrders[f][b].IsAccepted() && s.CurrentOrders[f][b].Assignee == LocalIP && !s.CurrentOrders[f][b].SentToAssigneeElevator {
				orders = append(orders, s.CurrentOrders[f][b])
			}
>>>>>>> d597c8aac8f40ee07545d414c695c77d43dac16f
		}
	}
	return orders
}

func MarkOrdersAsSent(orders []et.ElevOrder) {
	s, _ := systems[LocalIP]
	for _, order := range orders {
		if s.CurrentOrders[order.GetFloor()][int(order.GetButton())].Id == order.Id {
			s.CurrentOrders[order.GetFloor()][int(order.GetButton())].SentToAssigneeElevator = true
		}
	}
	systems[LocalIP] = s
}

func GetSystemElevators() []et.Elevator {
	var elevList []et.Elevator
	for _, system := range systems {
		elevList = append(elevList, system.E)
	}
	return elevList
}

