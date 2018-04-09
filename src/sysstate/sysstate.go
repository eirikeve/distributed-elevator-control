package sysstate

import (
	"strconv"
	"time"

	locIP "../elevnetwork/localip"
	et "../elevtype"
)

var LocalIP string

//var systems map[string]et.ElevState
var systems []et.ElevState
var netstate et.NetState

/*
 * initSysState is automatically called the first time we set the local systems
 */
func initSysState() {
	exsistsInSystems := false
	LocalIP, _ = locIP.LocalIP()
	// Check that systems contains the elevator this program is running at. If not, insert this elevator system.
	for _, element := range systems {
		if element.ID == LocalIP {
			exsistsInSystems = true
			break
		}
	}
	if !exsistsInSystems {
		newElevState := et.ElevState{ID: LocalIP, E: et.EmptyElevator()}
		systems = append(systems, newElevState)
	}

}

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
	}
	return elevList
}


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
		}
	}
	if !exsistsInSystems {
		// log TODO
	}
}

func UpdateSysElevator(sysIndex int, o et.ButtonEvent) {
	systems[sysIndex].CurrentOrders[o.Floor][int(o.Button)] = et.ElevOrder{
		Id:                strconv.FormatInt(time.Now().Unix(), 16),
		Order:             o,
		TimestampReceived: time.Now().Unix(),
		Status:            et.Received,
	}
}

