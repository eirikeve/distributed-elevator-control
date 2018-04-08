package sysstate

import (
	locIP "../elevnetwork/localip"
	et "../elevtype"
)

var LocalIP string
var initialized = false

//var systems map[string]et.ElevState
var systems = make(map[string]et.ElevState)
var netstate et.NetState

func initSysState() {

	if initialized {
		return
	}

	LocalIP, _ = locIP.LocalIP()

	_, localSysExists := systems[LocalIP]

	if !localSysExists {
		newElevState := et.ElevState{ID: LocalIP, E: et.EmptyElevator()}
		systems[LocalIP] = newElevState
	}

	initialized = true

}

func SetSystems(sys []et.ElevState) {
	systems = make(map[string]et.ElevState)
	for _, system := range sys {
		systems[system.ID] = system
	}
	// Initialize after the assignment since this guarantees local system being in systems after func call
	if !initialized {
		initSysState()
	}

}

func GetSystems() []et.ElevState {
	var sys []et.ElevState
	for _, system := range systems {
		sys = append(sys, system)
	}
	return sys
}

func GetSystemElevators() []et.Elevator {
	var elevList []et.Elevator
	for _, system := range systems {
		elevList = append(elevList, system.E)
	}
	return elevList
}
