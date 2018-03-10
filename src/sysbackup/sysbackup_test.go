package sysbackup

import (
	"testing"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

func TestBackup(t *testing.T) {
	log.Info("sysbackup Test: Starting Backup Test")
	SetupSysBackup()
	states := setupStates()

	backupTestWrite("something\n")
	Backup(states)

}

func TestRecover(t *testing.T) {
	states, err := Recover(time.Now().Add(-time.Second * 20))
	if err != nil {
		// @todo
		return
	}
	for _, s := range states {
		log.WithFields(log.Fields{
			"ID":         s.ID,
			"Floor":      s.E.Floor,
			"LastUpdate": s.LastUpdate,
		}).Info("sysbackup Recover: Result")
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
	state[0].FinishedOrders = make([]et.ElevOrder, 0)
	state[1].FinishedOrders = make([]et.ElevOrder, 0)
	state[2].FinishedOrders = make([]et.ElevOrder, 0)
	return state
}
