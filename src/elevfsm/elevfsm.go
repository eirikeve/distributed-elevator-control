package elevfsm

import (
	"time"

	timer "../elevtimer"
	et "../elevtype"
	sb "../sysbackup"
	ss "../sysstate"
	log "github.com/sirupsen/logrus"
)

var elevator et.Elevator
var lastRegisteredFloor int
var doorTimeoutSignalOutput chan bool
var errorCount int

const initFailTimeout time.Duration = 5 * time.Second

func InitFSM(doorTimeoutSignal chan bool, e *et.Elevator) {
	doorTimeoutSignalOutput = doorTimeoutSignal
	if e == nil {
		elevator = et.Elevator{
			Floor:               et.BOTTOMFLOOR,
			MovementDirection:   et.MD_Down,
			MovDirFromLastFloor: et.MD_Down,
			State:               et.Initializing,
			ErrorState:          et.FullFunctionality}
		log.WithField("elevator", elevator).Debug("elevfsm Initialize: No ref, reinitialized elevator")
	} else {
		elevator := &e
		log.WithField("elevator", elevator).Debug("elevfsm Initialize: Initialized elevator from ref")
	}
	// If recovered from Backup, elevator is set to equal backUp data
	if sb.IsInitializedFromBackup() && ss.SysIsInitialized() {
		elevator = ss.GetLocalSystem().E
	} else {
		//timer.Start("Initialization", initFailTimeout, doorTimeoutSignalOutput)
		// Start by moving downwards

		initialize()
	}
}

// Functions for running the local elevator
/*
func GetMovementDirection() elevtype.MotorDirection {
	return OrderLogicGetMovementDirection(elevator)
}
*/
func GetPanelLights() [et.NumFloors][et.NumButtons]et.ButtonLamp {
	var lights [et.NumFloors][et.NumButtons]et.ButtonLamp
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			// @TODO If order is nil this will not work //@BUG
			// make a get function or something else that returns the value.
			lights[f][b] = et.ButtonLamp{Floor: f, Button: et.ButtonType(b), Value: (elevator.Orders[f][b].GetID() != "")}
		}
	}
	return lights
}

func GetDoorOpenLight() bool {
	return (elevator.State == et.Unloading)
}

func GetState() et.ElevatorState {
	return elevator.State
}

func GetMotorDir() et.MotorDirection {
	return elevator.MovementDirection
}

func GetFloor() int {
	return elevator.Floor
}
func HandleOrders() {
	switch elevator.State {
	// Only go to moving from idle.
	case et.Idle:
		//log.WithField("shouldStop", OrderLogicCheckShouldStopAtFloor(elevator)).Debug("shouldStop")
		newDirMovement := OrderLogicGetMovementDirection(elevator)
		if newDirMovement != et.MD_Stop {
			elevator.MovDirFromLastFloor = newDirMovement
			move(newDirMovement)
		} else if OrderLogicCheckIfRequestsAtCurrentFloor(elevator) {
			unload()
		}
	default:

	}
}
func RegisterFloor(floor int) {
	if !(isValidFloor(floor)) {
		log.WithField("floor", floor).Error("elevfsm RegisterFloor: Ignoring invalid input")
		return
	}
	log.WithField("floor", floor).Debug("elevfsm RegisterFloor:")
	updateFloor(floor)

	switch elevator.State {
	case et.Initializing:
		//timer.Stop("Initialization") // No need to signal timeout, since we reached a floor
		idle()
	case et.Moving:
		if OrderLogicCheckShouldStopAtFloor(elevator) {
			unload()
		} else if floor == et.BOTTOMFLOOR && elevator.MovementDirection == et.MD_Down ||
			floor == et.TOPFLOOR && elevator.MovementDirection == et.MD_Up {
			idle()
		}
	case et.Unloading:
		if OrderLogicCheckShouldStopAtFloor(elevator) { //@TODO is this sufficient?
			unload()
		}
		return
	case et.Idle:
		// Do nothing

	default:
		log.WithField("state", elevator.State).Error("elevfsm RegisterFloor: Unhandled state")
	}

}

func RegisterTimerTimeout() {
	switch elevator.State {
	case et.Unloading:
		log.Info("elevfsm RegisterTimeOut: Going from Unloading to Idle")
		idle()
	case et.Initializing:
		switch elevator.MovementDirection {
		case et.MD_Down: // normal initialization
			setDir(et.MD_Up)
			errorCount++
			//timer.Start("InitializationRetry", initFailTimeout, doorTimeoutSignalOutput)
			log.Warning("elevfsm RegisterTimerTimeout: Retrying Init, moving up")

		case et.MD_Up:
			setDir(et.MD_Stop)
			errorCount++
			log.Error("elevfsm RegisterTimerTimeout: Error, could not init, restarting")
			//@TODO restart here
		}
		return
	case et.Idle:
		fallthrough
	case et.Moving:
		fallthrough
	default:
		log.WithField("state", elevator.State).Error("elevfsm RegisterTimerTimeout: Unhandled state")
		errorCount++ //@TODO maybe change to an internal variable in elevator?
	}
}

// Functions used when interfacing with elevNetworkHandler
func PushOrderToQueue(order et.GeneralOrder) {
	floor := order.GetFloor()
	var button = int(order.GetButton())
	elevator.Orders[floor][button] = order.ToSimpleOrder()
	// Need some logic here! @TODO
	log.WithField("btnEvent", order.GetOrder()).Info("elevfsm PushOrderToQueue: Recv")
	log.WithFields(log.Fields{
		"registeredOrder": elevator.Orders[floor][button],
		"floor":           floor,
		"button":          button,
	}).Info("elevfsm PushOrderToQueue: Added to queue")
}
func RemOrderFromQueue(order et.ElevOrder) {
	floor := order.GetFloor()
	var button = int(order.GetButton())
	elevator.Orders[floor][button] = et.SimpleOrder{} // Default ID is "" which evaluates to Empty
}

func GetElevator() et.Elevator {
	return elevator
}

func ForceState(state et.ElevatorState) {
	log.WithField("state", state).Warning("elevfsm ForceState: Forced state change")
	elevator.State = state
}

func setState(state et.ElevatorState) {
	log.WithField("state", state).Debug("elevfsm setState: State change")
	elevator.State = state
}

func setDir(dir et.MotorDirection) {
	log.WithField("movementDir", dir).Debug("elevfsm setDir: Dir change")
	elevator.MovementDirection = dir
}

func initialize() {
	log.Debug("elevfsm initialize: Initializing")
	setState(et.Initializing)
	setDir(et.MD_Down)
}

func unload() {
	log.WithField("floor", elevator.Floor).Debug("elevfsm unload: Unloading")
	timer.Start("UnloadTimer", time.Second*3, doorTimeoutSignalOutput)
	setState(et.Unloading)
	setDir(et.MD_Stop)
	elevator = OrderLogicClearRequestsOnCurrentFloor(elevator, elevator.MovDirFromLastFloor)
	//newMovementDir := OrderLogicGetMovementDirection(elevator)
	// @TODO maybe rewrite this so that we don't need to reassign elevator.

}

func idle() {
	log.WithField("floor", elevator.Floor).Debug("elevfsm idle: Idling")
	setDir(et.MD_Stop)
	setState(et.Idle)
}

func move(dir et.MotorDirection) {
	if dir == et.MD_Stop {
		idle()
	} else {
		log.WithFields(log.Fields{"floor": elevator.Floor, "dir": dir}).Debug("elevfsm move: Moving")
		setState(et.Moving)
		setDir(dir)
	}

}

func updateFloor(floor int) {
	lastRegisteredFloor = elevator.Floor
	elevator.Floor = floor
}

func isValidFloor(floor int) bool {
	return (0 <= floor && floor < et.NumFloors)
}
