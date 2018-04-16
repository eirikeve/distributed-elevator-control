package elevfsm

import (
	"time"

	timer "../elevtimer"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

/*
  The elevfsm module contains information and logic for the elevator, which is used to determine its actions.
  It is used by elevhandler to run, share and store information about the local elevator.
  The module elevorderevaluation uses the FSM for simulation of the elevator.
*/

////////////////////////////////
// Module variables
////////////////////////////////

var elevator et.Elevator
var lastRegisteredFloor int
var doorTimeoutSignalOutput chan bool

const initFailTimeout time.Duration = 5 * time.Second

////////////////////////////////
// Interface
////////////////////////////////

/*InitFSM intializes the elevator. Either by using stored backup information, if available,
 * or to defualt.
 * @arg doorTimeoutSignal:
 * @arg e: An elevator with information from backup or nil pointer representing a new elevator
 */
func InitFSM(doorTimeoutSignal chan bool, e *et.Elevator) {
	doorTimeoutSignalOutput = doorTimeoutSignal
	// When we initialize the elevator from backup, e will have a non-zero value.
	if e == nil {
		elevator = et.Elevator{
			Floor:               et.BOTTOMFLOOR,
			MovementDirection:   et.MD_Down,
			MovDirFromLastFloor: et.MD_Down,
			State:               et.Initializing,
			ErrorState:          et.FullFunctionality}
		log.WithField("elevator", elevator).Debug("elevfsm Initialize: No ref, reinitialized elevator")
	} else {
		elevator := *e
		log.WithField("elevator", elevator).Debug("elevfsm Initialize: Initialized elevator from ref")
	}

	runCurrentStateFunction(elevator)
}

/*HandleOrders (.), when in Idle: decides if the elevator should start moving, unload or
 * remain in Idle by calulating a new movement direction and checking current orders.
 * Else do nothing.
 */
func HandleOrders() {
	switch elevator.State {
	// Only go to moving from idle.
	case et.Idle:
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

/*RegisterFloor (.), updates the current floor of the elevator and consideres to change state based on given input floor, current
 * state and exsisting orders. Function is called when a change in floor is detected.
 * @arg floor: Inputs a new registered floor
 */
func RegisterFloor(floor int) {
	if !(isValidFloor(floor)) {
		log.WithField("floor", floor).Error("elevfsm RegisterFloor: Ignoring invalid input")
		return
	}
	log.WithField("floor", floor).Debug("elevfsm RegisterFloor:")
	updateFloor(floor)

	switch elevator.State {
	case et.Initializing:
		timer.Stop("Initialization") // No need to signal timeout, since we reached a floor
		idle()
	case et.Moving:

		if OrderLogicCheckShouldStopAtFloor(elevator) {
			unload() //@TODO add extra condition for unloading. In case of timeout.
		} else if floor == et.BOTTOMFLOOR && elevator.MovementDirection == et.MD_Down ||
			floor == et.TOPFLOOR && elevator.MovementDirection == et.MD_Up {
			idle()
		}
	case et.Unloading:
		if OrderLogicCheckShouldStopAtFloor(elevator) {
			unload()
		}
		return
	case et.Idle:
		// Do nothing

	default:
		log.WithField("state", elevator.State).Error("elevfsm RegisterFloor: Unhandled state")
	}

}

/*RegisterTimerTimeout (.), called when the FSM-timer has timed out.
 * If the elevator is unloading, it returns to Idle. Meaning unloading is completed.
 * If called when the elevator is initializing, it the elevator has failed to initialize
 * and will attept to reinitialize.
 */
func RegisterTimerTimeout() {
	switch elevator.State {
	case et.Unloading:
		log.Info("elevfsm RegisterTimeOut: Going from Unloading to Idle")
		idle()
	case et.Initializing:
		atteptToReinitialize()
	case et.Idle:
		fallthrough
	case et.Moving:
		fallthrough
	default:
		log.WithField("state", elevator.State).Error("elevfsm RegisterTimerTimeout: Unhandled state")
	}
}

/*GetDoorOpenLight (.)
 * @return: true if unloading, else false
 */
func GetDoorOpenLight() bool {
	return (elevator.State == et.Unloading)
}

/*GetState (.)
 * @return: The current elevator state
 */
func GetState() et.ElevatorState {
	return elevator.State
}

/*GetMotorDir (.)
 * @return: The current movement direction
 */
func GetMotorDir() et.MotorDirection {
	return elevator.MovementDirection
}

/*GetFloor (.)
 * @return: The current floor
 */
func GetFloor() int {
	return elevator.Floor
}

/*PushQueue (.) takes a set of orders and pushes them to the local queue
 * @arg orders: Orders received from the nethandler
 */
func PushQueue(orders [et.NumFloors][et.NumButtons]et.SimpleOrder) {
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			elevator.Orders[f][b] = orders[f][b]
		}
	}
}

/*RemoveOrderFromQueue (.) Removes the given order from the local elevator queue
 */
func RemOrderFromQueue(order et.GeneralOrder) {
	floor := order.GetFloor()
	button := int(order.GetButton())
	elevator.Orders[floor][button] = et.SimpleOrder{} // Default ID is "" which evaluates to Empty
}

/*GetElevator (.)
 * @returns: The local elevator
 */
func GetElevator() et.Elevator {
	return elevator
}

/*MarkElevatorSentTONetHandler (.)Marks that the elevator was successfully passed to the chan to nethandler.
 *Avoids duplicate registrations of finished orders.
 */
func MarkElevatorSentToNetHandler() {
	elevator.FinishedOrders = nil
}

////////////////////////////////
// Auxiliary
////////////////////////////////

/*setState (.)
 * @arg state: Sets the elevator state to given state
 */
func setState(state et.ElevatorState) {
	log.WithField("state", state).Debug("elevfsm setState: State change")
	elevator.State = state
}

/*setDir (.)
 * @arg dir: Sets the elevators movements direction to dir
 */
func setDir(dir et.MotorDirection) {
	log.WithField("movementDir", dir).Debug("elevfsm setDir: Dir change")
	elevator.MovementDirection = dir
}

/*initialize (.) sets the elevator state to initialize
 * and movement direction downwards.
 */
func initialize() {
	log.Debug("elevfsm initialize: Initializing")
	setState(et.Initializing)
	setDir(et.MD_Down)
}

/*atteptToReinitialize (.), called if initialization has failed to detect floor within given time limit.
 * First time called the movement direction of the elevator upwards and restarts timer.
 * Second time, the elevator will stop and remain silent.
 */
func atteptToReinitialize() {
	switch elevator.MovementDirection {
	case et.MD_Down: // normal initialization
		setDir(et.MD_Up)
		timer.Start("InitializationRetry", initFailTimeout, doorTimeoutSignalOutput)
		log.Warning("elevfsm RegisterTimerTimeout: Retrying Init, moving up")

	case et.MD_Up:
		setDir(et.MD_Stop)
		log.Error("elevfsm RegisterTimerTimeout: Error, could not init, restarting")
		//@TODO restart here

	}
}

/*unload (.) Called when serving a order at a floor. Starts timer, which will later signalize when
 * unloading is completed. Sets the elevator state to unlading and movement direction to stop.
 * Appends the served order to the finished order queue.
 */
func unload() {
	log.WithField("floor", elevator.Floor).Debug("elevfsm unload: Unloading")
	timer.Start("UnloadTimer", time.Second*3, doorTimeoutSignalOutput)

	setState(et.Unloading)
	setDir(et.MD_Stop)

	elevator.FinishedOrders = append(elevator.FinishedOrders, OrderLogicGetRequestsWeCanClearOnCurrentFloor(elevator, elevator.MovDirFromLastFloor)...)

}

/*idle (.) Sets the elevators movement direction to stop and changes
 * the state to Idle.
 */
func idle() {
	log.WithField("floor", elevator.Floor).Debug("elevfsm idle: Idling")
	setDir(et.MD_Stop)
	setState(et.Idle)
}

/*move (.)  Changes the elevator state to Moving, unless given direction is stop. Then the elevator
 * will enter state Idle.
 * @arg dir: If dir equal stop go to idle, else movement direction is sat to dir
 */
func move(dir et.MotorDirection) {
	if dir == et.MD_Stop {
		idle()
	} else {
		log.WithFields(log.Fields{"floor": elevator.Floor, "dir": dir}).Debug("elevfsm move: Moving")
		setState(et.Moving)
		setDir(dir)
	}

}

/*updateFloor (.) Updates current elevator floor,
 * and sets lastRegistered floor
 */
func updateFloor(floor int) {
	lastRegisteredFloor = elevator.Floor
	elevator.Floor = floor
}

/*isValidFloor (.) inserts a floor and returns
 * true if it is within valid floors, else returns false.
 */
func isValidFloor(floor int) bool {
	return (0 <= floor && floor < et.NumFloors)
}

/*runCurrentStateFunction (.) executes the state function
 * corresponding to the elevators current state.
 * @arg elev: Elevator containing the present state
 */
func runCurrentStateFunction(elev et.Elevator) {
	switch elev.State {
	case et.Idle:
		idle()
	case et.Moving:
		move(elev.MovementDirection)
	case et.Unloading:
		unload()
	default:
		initialize()
	}
}
