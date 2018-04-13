package sysstate

import (
	"time"

	def "../elevdef"
	network "../elevnetwork"
	locIP "../elevnetwork/localip"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

var LocalIP string
var LocalID int32
var initialized = false
var backInit = false

//var systems map[string]et.ElevState
var systems = make(map[int32]et.ElevState)
var acksForBroadcasting []et.AckNackMsg

func SysIsInitialized() bool {
	return backInit
}

/*
 * initSysState is automatically called the first time we set the local systems
 */
func initSysState() {

	if initialized {

		return
	}

	LocalIP, _ = locIP.LocalIP()
	LocalID, _ = locIP.LocalID()

	_, localSysExists := systems[LocalID]

	if !localSysExists {
		newElevState := et.ElevState{ID: LocalID, E: et.EmptyElevator(), StartupTime: time.Now().Unix()}
		systems[LocalID] = newElevState
	} else {
		backInit = true
	}

	initialized = true

	log.WithField("localID", LocalID).Warn("sysstate: Initialized")

}

func SetSystemsStates(sys []et.ElevState) {
	systems = make(map[int32]et.ElevState)
	for _, system := range sys {
		systems[system.ID] = system
	}
	// Initialize after the assignment since this guarantees local system being in systems after func call
	if !initialized {
		initSysState()
	}
}

func SetSystemsStatesFromBackup(sys []et.ElevState) {
	SetSystemsStates(sys)
	// mark orders as not sent to elevator, so ensure that local orders are resent to the elevator immediately
	for key, es := range systems {
		for f := 0; f < def.NumFloors; f++ {
			for b := 0; b < def.NumButtons; b++ {
				if es.CurrentOrders[f][b].Id != "" {
					es.CurrentOrders[f][b].SentToAssigneeElevator = false
				}
			}
		}
		systems[key] = es
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

	}
	return sys
}

func GetLocalSystem() et.ElevState {
	return systems[LocalID]
}

func GetUnsentLocalSystemOrders() []et.SimpleOrder {
	var orders []et.SimpleOrder
	// Get orders to delegate
	s, _ := systems[LocalID]
	for f := 0; f < def.NumFloors; f++ {
		for b := 0; b < def.NumButtons; b++ {
			if s.CurrentOrders[f][b].IsAccepted() && !s.CurrentOrders[f][b].SentToAssigneeElevator {
				o := s.CurrentOrders[f][b].ToSimpleOrder()
				orders = append(orders, o)
			}
		}
	}
	// Get orders to remove from local elevator queue
	for f := 0; f < def.NumFloors; f++ {
		for b := 0; b < def.NumButtons; b++ {
			if s.E.Orders[f][b].IsActive() && s.CurrentOrders[f][b].Id == "" && !(et.IsCabButton(s.E.Orders[f][b].Order)) {
				o := s.E.Orders[f][b]
				o.TagRemoveOrder = true
				orders = append(orders, o)
			}
		}
	}
	return orders
}

func MarkOrdersAsSent(orders []et.SimpleOrder) {
	s, _ := systems[LocalID]
	for _, order := range orders {
		if s.CurrentOrders[order.GetFloor()][int(order.GetButton())].Id == order.GetID() {
			s.CurrentOrders[order.GetFloor()][int(order.GetButton())].SentToAssigneeElevator = true
		}
	}
	systems[LocalID] = s
}

func GetSystemElevators() []et.Elevator {
	var elevList []et.Elevator
	for _, system := range systems {
		elevList = append(elevList, system.E)
	}
	return elevList
}
