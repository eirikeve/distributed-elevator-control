package sysbackup

import (
	"testing"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

func TestBackup(t *testing.T) {
	log.SetLevel(log.InfoLevel)
	log.Info("sysbackup Test: Starting Backup Test")
	states := setupStates()
	//setupSysBackup()
	log.Info("sysbackup Test: Setup states")

	Backup(states)
	log.Info("sysbackup Test: Backed up first states")

	states = setupStates2()
	log.Info("sysbackup Test: Setup states 2")

	Backup(states)
	log.Info("sysbackup Test: Backed up states 2")

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
			"ID":    s.ID,
			"Floor": s.E.Floor,
		}).Info("sysbackup Recover: Result")

		for f := 0; f < et.NumFloors; f++ {
			for b := 0; b < et.NumButtons; b++ {
				if s.CurrentOrders[f][b].Id != "" {
					log.WithFields(log.Fields{
						"sysID":         s.ID,
						"f":             f,
						"b":             b,
						"orderId":       s.CurrentOrders[f][b].Id,
						"orderAssignee": s.CurrentOrders[f][b].Assignee,
					}).Info("sysbackup Test: Recovered an order in the currentOrder queue")
				}
			}
		}
		for i := 0; i < len(s.FinishedOrders); i++ {
			if s.FinishedOrders[i].Id != "" {
				log.WithFields(log.Fields{
					"sysID":         s.ID,
					"i":             i,
					"orderId":       s.FinishedOrders[i].Id,
					"orderAssignee": s.FinishedOrders[i].Assignee,
				}).Info("sysbackup Test: Recovered an order in the finishedOrder queue")
			}
		}

	}
}

func setupStates() []et.ElevState {
	state := make([]et.ElevState, 4)
	state[1].ID = 1
	state[2].ID = 2
	state[3].ID = 3
	state[1].E.Floor = 0
	state[2].E.Floor = 1
	state[3].E.Floor = 2

	state[1].CurrentOrders[0][0] = et.ElevOrder{
		Id:                "Active Order 0 of elev 0",
		Order:             et.ButtonEvent{Floor: 0, Button: et.BT_HallUp},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Received,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[1].CurrentOrders[1][0] = et.ElevOrder{
		Id:                "Active Order 1 of elev 0",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallUp},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[1].CurrentOrders[1][1] = et.ElevOrder{
		Id:                "Active Order 2 of elev 0",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallDown},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       2,
	}
	state[2].CurrentOrders[1][0] = et.ElevOrder{
		Id:                "Active Order 1 of elev 1",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallUp},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[2].CurrentOrders[1][1] = et.ElevOrder{
		Id:                "Active Order 2 of elev 1",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallDown},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       2,
	}
	state[3].CurrentOrders[1][0] = et.ElevOrder{
		Id:                "Active Order 1 of elev 2",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallUp},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[3].CurrentOrders[1][1] = et.ElevOrder{
		Id:                "Active Order 2 of elev 2",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallDown},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       2,
	}

	state[1].FinishedOrders[0] = et.ElevOrder{
		Id:                "Finished Order 1 of elev 0",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_Cab},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Finished,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[1].FinishedOrders[1] = et.ElevOrder{
		Id:                "Finished Order 2 of elev 0",
		Order:             et.ButtonEvent{Floor: 2, Button: et.BT_Cab},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Finished,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[2].FinishedOrders[5] = et.ElevOrder{
		Id:                "Finished Order of elev 1,2",
		Order:             et.ButtonEvent{Floor: 3, Button: et.BT_HallDown},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Finished,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       3,
	}
	state[3].FinishedOrders[5] = et.ElevOrder{
		Id:                "Finished Order of elev 1,2",
		Order:             et.ButtonEvent{Floor: 3, Button: et.BT_HallDown},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Finished,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       3,
	}

	return state
}

func setupStates2() []et.ElevState {
	state := make([]et.ElevState, 4)
	state[1].ID = 1
	state[2].ID = 2
	state[3].ID = 3
	state[1].E.Floor = 0
	state[2].E.Floor = 1
	state[3].E.Floor = 2

	state[1].CurrentOrders[0][0] = et.ElevOrder{
		Id:                "Active Order 0 of elev 0",
		Order:             et.ButtonEvent{Floor: 0, Button: et.BT_HallUp},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Received,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[1].CurrentOrders[1][0] = et.ElevOrder{
		Id:                "Active Order 1 of elev 0",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallUp},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[1].CurrentOrders[1][1] = et.ElevOrder{
		Id:                "Active Order 2 of elev 0",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallDown},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       2,
	}
	state[2].CurrentOrders[1][0] = et.ElevOrder{
		Id:                "Active Order 1 of elev 1",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallUp},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[2].CurrentOrders[1][1] = et.ElevOrder{
		Id:                "Active Order 2 of elev 1",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallDown},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       2,
	}
	state[3].CurrentOrders[1][0] = et.ElevOrder{
		Id:                "Active Order 1 of elev 2",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallUp},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[3].CurrentOrders[1][1] = et.ElevOrder{
		Id:                "Active Order 2 of elev 2",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallDown},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Accepted,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       2,
	}
	state[1].FinishedOrders[0] = et.ElevOrder{
		Id:                "Finished Order 1 of elev 0",
		Order:             et.ButtonEvent{Floor: 1, Button: et.BT_Cab},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Finished,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}
	state[2].FinishedOrders[1] = et.ElevOrder{
		Id:                "Finished Order 2 of elev 0",
		Order:             et.ButtonEvent{Floor: 2, Button: et.BT_Cab},
		TimestampReceived: time.Now().Add(time.Second * -10).Unix(),
		Status:            et.Finished,
		TimestampLastOrderStatusChange: time.Now().Unix(),
		Assignee:                       1,
	}

	return state
}
