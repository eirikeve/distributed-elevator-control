package sysstate

import (
	et "../elevtype"
)

var localIP string
var systems map[string]et.ElevState
var netstate et.NetState
