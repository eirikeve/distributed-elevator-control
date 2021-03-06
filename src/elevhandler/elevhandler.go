package elevhandler

import (
	"time"

	driver "../elevdriver"
	fsm "../elevfsm"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

/*
 * Module for communication between the network system and the local system.
 * Sends updates from the local modules to the network, while receiving updates from the network and
 * distributes that information to the other local system modules.
 */

////////////////////////////////
// Module variables
////////////////////////////////

var signalHandlerToStop chan bool
var fsmTimeoutSignal chan bool
var handlerRunning = false

////////////////////////////////
// Interface
////////////////////////////////

/*StartElevatorHandler (.) initializes the FSM, module variables and starts the
 * handler itself.

 * @arg orderQueueFromNetHandler: Input channel, receiving order queue from the nethandler
 * @arg buttonLightsFromNethandler: Input channel, receiving the entire systems active orders, used to set button lights
 * @arg buttonPressesToNethandler: Output channel, sending registered buttonPresses to nethandler, used for setting orders
 * @arg elevatorFSMToNethandler: Output channel,sending finished orders from to the Nethandler
 * @arg elevatorInitialState: Contains either a elevator initailized from backup, or nil hence a new elevator is shall be initialzed
 */
func StartElevHandler(
	orderQueueFromNethandler <-chan [et.NumFloors][et.NumButtons]et.SimpleOrder,
	buttonLightsFromNethandler <-chan et.ButtonLamp,
	buttonPressesToNethandler chan<- et.ButtonEvent,
	elevatorFSMToNethandler chan<- et.Elevator,
	elevatorInitialState *et.Elevator,
) {

	if handlerRunning {
		log.Warn("elevhandler Start: Already running")
		return
	}

	signalHandlerToStop = make(chan bool, 2)
	fsmTimeoutSignal = make(chan bool, 2)
	fsm.InitFSM(fsmTimeoutSignal, elevatorInitialState)
	handlerRunning = true

	go handler(signalHandlerToStop,
		orderQueueFromNethandler,
		buttonLightsFromNethandler,
		buttonPressesToNethandler,
		elevatorFSMToNethandler)

	log.Info("elevhandler StartElevatorHandler: Started")

}

/*StopElevatorHandler (.) is called to signal the Handler
 * to stop, buy writing true to signalHandlerToStop channel
 */
func StopElevHandler() {
	if !handlerRunning {
		log.Warn("elevhandler Stop: Handler was not runnning")
		return
	}
	log.Info("elevhandler StopElevatorHandler: Stopping")
	signalHandlerToStop <- true
	handlerRunning = false
	return
}

////////////////////////////////
// Auxiliary
////////////////////////////////

/*handler (.) starts the elevdriver, then starts running the FSM while receiving and sending updates to the
 * nethandler by communicating with the channel parameters.

 * @arg orderQueueFromNetHandler: Input channel, receiving order queue from the nethandler
 * @arg buttonLightsFromNethandler: Input channel, receiving the entire systems active orders, used to set button lights
 * @arg buttonPressesToNethandler: Output channel, sending registered buttonPresses to nethandler, used for setting orders
 * @arg elevatorFSMToNethandler: Output channel, sending finished orders from to the Nethandler
 */
func handler(
	signalHandlerToStop <-chan bool,
	orderQueueFromNethandler <-chan [et.NumFloors][et.NumButtons]et.SimpleOrder,
	buttonLightsFromNethandler <-chan et.ButtonLamp,
	buttonPressesToNethandler chan<- et.ButtonEvent,
	elevatorFSMToNethandler chan<- et.Elevator) {
	log.Debug("elevhandler handler: Starting")

	motorDirectionInput := make(chan et.MotorDirection, 2)
	floorIndicatorInput := make(chan int, 2)
	doorOpenLampInput := make(chan bool, 2)
	floorSensorOut := make(chan int, 5)

	driver.StartDriver(
		et.NumFloors,
		motorDirectionInput,
		buttonLightsFromNethandler,
		floorIndicatorInput,
		doorOpenLampInput,
		buttonPressesToNethandler,
		floorSensorOut,
	)
	defer driver.StopDriver()

	//[@todo] where to place these?
	handlerDebugLogMsgTimer := time.Now()

	sendFSMUpdatesTimer := time.Now()

	for {

		// Make elevator move
		fsm.HandleOrders()
		motorDir := fsm.GetMotorDir()

		select {

		// Pushing motor direction to Driver
		case motorDirectionInput <- motorDir:

		default:
		}

		// Get values to be sent
		if time.Now().Sub(sendFSMUpdatesTimer) > et.HandlerSendFSMUpdatesFreq {
			sendFSMUpdatesTimer = time.Now()

			elev := fsm.GetElevator()
			doorOpenLamp := fsm.GetDoorOpenLight()
			floor := fsm.GetFloor()

			select {
			case floorIndicatorInput <- floor:

			case doorOpenLampInput <- doorOpenLamp:

			// Pushing elevator to Network Handler
			case elevatorFSMToNethandler <- elev:
				fsm.MarkElevatorSentToNetHandler()
			default:
			}
		}

		select {
		// Elevator Handler Control
		case <-signalHandlerToStop:
			return

		// Receiving orders from the Network Handler
		case o := <-orderQueueFromNethandler:
			fsm.PushQueue(o)
		// Checking floor, registering in FSM
		case f := <-floorSensorOut:
			if et.BOTTOMFLOOR <= f && f <= et.TOPFLOOR {
				fsm.RegisterFloor(f)
			}
		// Checking timer timeout, registering in FSM
		case <-fsmTimeoutSignal:
			fsm.RegisterTimerTimeout()

		default:
			// Do nothing
		}

		if time.Now().Sub(handlerDebugLogMsgTimer) > et.HandlerDebugLogMsgFreq {
			handlerDebugLogMsgTimer = time.Now()
			log.Debug("elevhandler handler: Running")
		}
	}
}
