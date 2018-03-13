package sysstate

import (
	locIP "../elevnetwork/localip"
	et "../elevtype"
)

var LocalIP string

//var systems map[string]et.ElevState
var systems []et.ElevState
var netstate et.NetState

func initSysState() {
	exsistsInSystems := false
	LocalIP, _ = locIP.LocalIP()
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

func SetSystems(sys []et.ElevState) {
	systems = sys
	initSysState()
}

func GetSystems() []et.ElevState { return systems }

func GetSystemElevators() []et.Elevator {
	var elevList []et.Elevator
	for _, elev := range systems {
		elevList = append(elevList, elev.E)
	}
	return elevList
}

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
