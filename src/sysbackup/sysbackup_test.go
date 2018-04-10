package sysbackup

import (
	"testing"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

func TestBackup(t *testing.T) {
	log.Info("sysbackup Test: Starting Backup Test")
	states := setupStates()

	backupTestWrite("something\n")

	Backup(states)

	states = setupStates2()

	Backup(states)

}

func TestRecover(t *testing.T) {
	states, err := Recover(time.Now().Add(-time.Second * 20))
	if err != nil {
		log.WithField("err", err.Error()).Info("sysbackup Test: Could not recover")
		// @todo
		return
	}
	if len(states) == 0 {
		log.Info("sysbackup Test: Nothing to recover")
		return
	} else {
		log.WithField("num", len(states)).Info("sysbackup Test: Recovered states")
	}
	for _, s := range states {
		log.WithFields(log.Fields{
			"ID":         s.ID,
			"Floor":      s.E.Floor,
			"LastUpdate": s.LastUpdate,
		}).Info("sysbackup Recover: Result")
		if s.ID == "Elevator0" {
			log.WithField("Finished Order", s.FinishedOrders).Info("Elevator 0 Finished Orders")
		}
	}
}

func setupStates() []et.ElevState {
	state := make([]et.ElevState, 3)
	state[0].ID = "Elevator0"
	state[1].ID = "Elevator1"
	state[2].ID = "Elevator2"
	state[0].LastUpdate = time.Now()
	state[1].LastUpdate = time.Now().Add(-time.Second)
	state[2].LastUpdate = time.Now().Add(time.Second)
	state[0].E.Floor = 0
	state[1].E.Floor = 1
	state[2].E.Floor = 2
	state[0].FinishedOrders = make([]et.ElevOrder, 1)
	state[0].FinishedOrders[0] = et.ElevOrder{
		Id:                "Order of elev 0",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_Cab},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       "Elevator1",
	}
	state[1].FinishedOrders = make([]et.ElevOrder, 0)
	state[2].FinishedOrders = make([]et.ElevOrder, 0)

	return state
}

func setupStates2() []et.ElevState {
	state := make([]et.ElevState, 3)
	state[0].ID = "Elevator0"
	state[1].ID = "Elevator1"
	state[2].ID = "Elevator2"
	state[0].LastUpdate = time.Now()
	state[1].LastUpdate = time.Now().Add(-time.Second)
	state[2].LastUpdate = time.Now().Add(time.Second)
	state[0].E.Floor = 1
	state[1].E.Floor = 2
	state[2].E.Floor = 3
	state[0].FinishedOrders = make([]et.ElevOrder, 1)
	state[0].FinishedOrders[0] = et.ElevOrder{
		Id:                "Order of elev 0",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_Cab},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       "Elevator1",
	}
	state[1].FinishedOrders = make([]et.ElevOrder, 0)
	state[2].FinishedOrders = make([]et.ElevOrder, 0)

	return state
}
