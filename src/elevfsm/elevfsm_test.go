package elevfsm

import (
	"testing"
	"time"

	"../elevlog"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

func TestFSM(t *testing.T) {
	elevlog.InitLog(log.DebugLevel, false)
	var timeoutChan = make(chan bool)

	var case1 fsmTestScenario
	case1.TimeStepLength = 100 * time.Millisecond
	case1.TimeoutChan = timeoutChan

	InitFSM(timeoutChan, nil)

}

func simulateTestInstance() {
	// for timesteps:
	// read state [i]
	// push events [i]

}

func logEvents(timeStep int, elevator *et.Elevator, testScenario *fsmTestScenario) {
	currentTime := float64(timeStep) * testScenario.TimeStepLength.Seconds()
	log.WithFields(
		log.Fields{
			"step":           timeStep,
			"time":           currentTime,
			"floorRegister:": testScenario.FloorRegisters[timeStep],
			"timeout":        testScenario.Timeouts[timeStep],
			"orderRecv":      testScenario.Orders[timeStep],
		}).Debug("Events at step " + string(timeStep))
}

func logState(timeStep int, elevator *et.Elevator, testScenario *fsmTestScenario) {
	currentTime := float64(timeStep) * testScenario.TimeStepLength.Seconds()
	log.WithFields(
		log.Fields{
			"step":    timeStep,
			"time":    currentTime,
			"eState":  elevator.State,
			"eFloor:": elevator.Floor,
		}).Debug("States at step " + string(timeStep))
}

func initCase1(case1 *fsmTestScenario, timeoutChan chan bool) {
	var c *fsmTestScenario
	case1 = c

	case1.TimeStepLength = 100 * time.Millisecond
	case1.TimeoutChan = timeoutChan

	*(case1.ExpectedState[0]) = et.Initializing
	*(case1.FloorRegisters[1]) = 0
	*(case1.ExpectedState[2]) = et.Idle
	case1.Orders[3] = &et.SimpleOrder{"Order1", et.ButtonEvent{2, et.BT_HallUp}}
	*(case1.FloorRegisters[3]) = 0
	*(case1.ExpectedMD[4]) = et.MD_Up
	*(case1.FloorRegisters[5]) = -1
	*(case1.FloorRegisters[6]) = -1
	*(case1.FloorRegisters[7]) = 1
	*(case1.ExpectedState[8]) = et.Moving
	*(case1.FloorRegisters[9]) = -1
	*(case1.FloorRegisters[10]) = 2
	*(case1.ExpectedState[11]) = et.Unloading
	case1.Orders[11] = &et.SimpleOrder{"Order2", et.ButtonEvent{3, et.BT_HallDown}}
	*(case1.Timeouts[12]) = true
	*(case1.ExpectedState[13]) = et.Moving
	case1.Orders[10] = &et.SimpleOrder{"Order3", et.ButtonEvent{3, et.BT_Cab}}

	case1.Orders[15] = &et.SimpleOrder{"Order4", et.ButtonEvent{1, et.BT_Cab}}

}

type fsmTestScenario struct {
	TimeStepLength time.Duration
	FloorRegisters [100]*int
	Timeouts       [100]*bool
	Orders         [100]*et.SimpleOrder
	ExpectedMD     [100]*et.MotorDirection
	ExpectedState  [100]*et.ElevatorState
	TimeoutChan    chan bool
}
