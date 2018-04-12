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
	ordersDelegatedFromNetwork <-chan et.GeneralOrder,
	buttonPressesToNetwork chan<- et.ButtonEvent,
	elevStateToNetwork chan<- et.Elevator,
) {
	log.Info("elevhandler StartElevatorHandler: Starting")
	// @TODO Finish initialization
	signalHandlerToStop = make(chan bool, 2)
	fsmTimeoutSignal = make(chan bool, 2)
	fsm.InitFSM(fsmTimeoutSignal, nil)
	go handler(signalHandlerToStop,
		ordersDelegatedFromNetwork,
		buttonPressesToNetwork,
		elevStateToNetwork)

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
	ordersDelegatedFromNetwork <-chan et.GeneralOrder,
	buttonPressesToNetwork chan<- et.ButtonEvent,
	elevStateToNetwork chan<- et.Elevator) {
	log.Debug("elevhandler handler: Starting")

	motorDirectionInput := make(chan et.MotorDirection, 2)
	buttonLampInput := make(chan et.ButtonLamp, 12)
	floorIndicatorInput := make(chan int, 2)
	doorOpenLampInput := make(chan bool, 2)
	//buttonPressSensorOut := elevToNetwork //make(chan et.ButtonEvent, 12)
	floorSensorOut := make(chan int, 2)

	driver.StartDriver(
		et.NumFloors,
		motorDirectionInput,
		buttonLampInput,
		floorIndicatorInput,
		doorOpenLampInput,
		buttonPressesToNetwork,
		floorSensorOut,
	)
	defer driver.StopDriver()

	//@TODO REMOVE. ONLY FOR DEBUGGING
	//fsm.RegisterFloor(0)

	//@TODO program loop
	handlerDebugLogMsgTimer := time.Now()
	handlerDebugLogMsgFreq := 2 * time.Second
	sendFSMUpdatesFreq := 10 * time.Millisecond
	sendFSMUpdatesTimer := time.Now()
	//timer.StartDelayedFunction("ElevHandler Watchdog", time.Second*2, func() { panic("ElevHandler Watchdog: timeout") })
	//defer timer.Stop("ElevHandler Watchdog")
	for {
		//timer.Update("ElevHandler Watchdog", time.Second*4)
		// Make elevator move
		fsm.HandleOrders()

		// Get values to be sent

		if time.Now().Sub(sendFSMUpdatesTimer) > sendFSMUpdatesFreq {
			sendFSMUpdatesTimer = time.Now()

			elev := fsm.GetElevator()
			buttonLamps := fsm.GetPanelLights()
			doorOpenLamp := fsm.GetDoorOpenLight()
			motorDir := fsm.GetMotorDir()
			floor := fsm.GetFloor()

			select {
			case motorDirectionInput <- motorDir:
			case floorIndicatorInput <- floor:
			case doorOpenLampInput <- doorOpenLamp:
			case elevStateToNetwork <- elev:
			default:
			}

			// Push button lamps
			for i := 0; i < et.NumFloors; i++ {
				for j := 0; j < et.NumButtons; j++ {
					select {
					// try to send buttonLamp inputs
					case buttonLampInput <- buttonLamps[i][j]:
						//log.Debug("Sent lamp")
					}
				}
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
		case o := <-ordersDelegatedFromNetwork:
			fsm.PushOrderToQueue(o)
		// Checking floor, registering in FSM
		case f := <-floorSensorOut:
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
