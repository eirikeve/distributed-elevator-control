package sysstate

import (
	et "../elevtype"
)

var localIP string

//var systems map[string]et.ElevState
var systems []et.ElevState
var netstate et.NetState

func Getsystems()([]et.ElevState){
	return systems
}

func GetSystemElevators()([]et.Elevator){
	var elevList []et.Elevator
	for _,elev :=range systems{
		elevList = append(elevList,elev.E)
	}
	return elevList
}