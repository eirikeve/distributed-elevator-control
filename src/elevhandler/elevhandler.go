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
) {
	log.Info("elevhandler StartElevatorHandler: Starting")
	// @TODO Finish initialization
	signalHandlerToStop = make(chan bool, 2)
	fsmTimeoutSignal = make(chan bool, 2)
	fsm.InitFSM(fsmTimeoutSignal, nil)
	go handler(signalHandlerToStop,
		orderQueueFromNethandler,
		buttonLightsFromNethandler,
		buttonPressesToNethandler,
		elevatorFSMToNethandler)

	log.Error("elevhandler StartElevatorHandler: Started")

}

func StopElevatorHandler() {
	log.Info("elevhandler StopElevatorHandler: Stopping")
	//@BUG this does not send
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
	//buttonLampInput := make(chan et.ButtonLamp, 12)
	floorIndicatorInput := make(chan int, 2)
	doorOpenLampInput := make(chan bool, 2)
	//buttonPressSensorOut := elevToNetwork //make(chan et.ButtonEvent, 12)
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

	//@TODO REMOVE. ONLY FOR DEBUGGING
	//fsm.RegisterFloor(0)

	//@TODO program loop
	handlerDebugLogMsgTimer := time.Now()
	handlerDebugLogMsgFreq := 2 * time.Second
	sendFSMUpdatesFreq := 100 * time.Millisecond
	sendFSMUpdatesTimer := time.Now()
	//timer.StartDelayedFunction("ElevHandler Watchdog", time.Second*2, func() { panic("ElevHandler Watchdog: timeout") })
	//defer timer.Stop("ElevHandler Watchdog")
	for {
		//timer.Update("ElevHandler Watchdog", time.Second*4)
		// Make elevator move
		fsm.HandleOrders()
		motorDir := fsm.GetMotorDir()
		select {
		case motorDirectionInput <- motorDir:
		default:
		}

		// Get values to be sent

		if time.Now().Sub(sendFSMUpdatesTimer) > sendFSMUpdatesFreq {
			sendFSMUpdatesTimer = time.Now()

			elev := fsm.GetElevator()
			//buttonLamps := fsm.GetPanelLights()
			doorOpenLamp := fsm.GetDoorOpenLight()
			floor := fsm.GetFloor()

			select {
			case floorIndicatorInput <- floor:
			case doorOpenLampInput <- doorOpenLamp:
			case elevatorFSMToNethandler <- elev:
				fsm.MarkElevatorSentToNetHandler()
			default:

			}
		}

		select {
		// Elevator Handler Control
		case <-signalHandlerToStop:
			return
			// Pushing button presses to the Network Handler done by driver

		//log.WithField("button", b).Warning("elevhandler elevHandlerInstance: Registered btn, sending not implemented")

		// Pushing elevator state to Network Handler

		// Receiving orders from the Network Handler
		case o := <-orderQueueFromNethandler:
			fsm.PushQueue(o)
		// Checking floor, registering in FSM
		case f := <-floorSensorOut:
			println("\n\n\n Reading floor \n\n\n")
			if et.BOTTOMFLOOR <= f && f <= et.TOPFLOOR {
				fsm.RegisterFloor(f)
			}
		// Checking timer timeout, registering in FSM
		case <-fsmTimeoutSignal:
			log.Debug("elevhandler handler: Before")
			fsm.RegisterTimerTimeout()
			log.Debug("elevhandler handler: After")
		// Pushing motor direction to Driver

		default:
			// // nothing
		}

		if time.Now().Sub(handlerDebugLogMsgTimer) > handlerDebugLogMsgFreq {
			handlerDebugLogMsgTimer = time.Now()
			log.Debug("elevhandler handler: Running")
		}
		//log.Error("elevhandler handler: Running")
	}
}

func getElevatorBackup() (*et.Elevator, error) {
	return nil, errors.New("Backup not yet implemented")
}
