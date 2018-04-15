package elevhandler

import (
	"errors"
	"time"

	driver "../elevdriver"
	fsm "../elevfsm"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

var signalHandlerToStop chan bool
var fsmTimeoutSignal chan bool

func StartElevatorHandler(
	orderQueueFromNethandler <-chan [et.NumFloors][et.NumButtons]et.SimpleOrder,
	buttonLightsFromNethandler <-chan et.ButtonLamp,
	buttonPressesToNethandler chan<- et.ButtonEvent,
	elevatorFSMToNethandler chan<- et.Elevator,
	elevatorInitialState *et.Elevator,
) {
	log.Info("elevhandler StartElevatorHandler: Starting")

	signalHandlerToStop = make(chan bool, 2)
	fsmTimeoutSignal = make(chan bool, 2)
	fsm.InitFSM(fsmTimeoutSignal, elevatorInitialState)

	go handler(signalHandlerToStop,
		orderQueueFromNethandler,
		buttonLightsFromNethandler,
		buttonPressesToNethandler,
		elevatorFSMToNethandler)

	log.Error("elevhandler StartElevatorHandler: Started")

}

func StopElevatorHandler() {
	log.Info("elevhandler StopElevatorHandler: Stopping")
	signalHandlerToStop <- true
	log.Debug("elevhandler StopElevatorHandler: Stop signalled")
	return
}

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

	//@TODO program loop
	handlerDebugLogMsgTimer := time.Now()
	handlerDebugLogMsgFreq := 2 * time.Second
	sendFSMUpdatesFreq := 100 * time.Millisecond
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
		if time.Now().Sub(sendFSMUpdatesTimer) > sendFSMUpdatesFreq {
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

		if time.Now().Sub(handlerDebugLogMsgTimer) > handlerDebugLogMsgFreq {
			handlerDebugLogMsgTimer = time.Now()
			log.Debug("elevhandler handler: Running")
		}
	}
}

func getElevatorBackup() (*et.Elevator, error) {
	return nil, errors.New("Backup not yet implemented")
}
