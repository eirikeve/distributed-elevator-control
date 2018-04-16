package nethandler

import (
	"time"

	network "../elevnetwork"
	b "../elevnetwork/bcast"
	eval "../elevorderevaluation"
	et "../elevtype"
	sb "../sysbackup"
	ss "../sysstate"
	log "github.com/sirupsen/logrus"
)

/*

nethandler controls communication between systems, order delegation, and order accepting.
It also determines which panel button lamps to light up.

*/

////////////////////////////////
// Module variables
////////////////////////////////

var signalNetHandlerToStop chan bool
var handlerRunning = false

////////////////////////////////
// Interface
////////////////////////////////

/*StartNetHandler creates the nethandler instance.
 * @arg orderQueueFromNetHandler: The queue the local elevator is to carry out is sent here
 * @arg buttonLightsFromNetHandler: Values for the button panel lights
 * @arg buttonPressesToNetHandler: Nethandler receives button presses from the local elevator here
 * @arg elevStateToNetHandler: Nethandler receives the floor/movDir/current action of the local elevator here
 */
func StartNetHandler(
	orderQueueFromNethandler chan<- [et.NumFloors][et.NumButtons]et.SimpleOrder,
	buttonLightsFromNethandler chan<- et.ButtonLamp,
	buttonPressesToNethandler <-chan et.ButtonEvent,
	elevStateToNethandler <-chan et.Elevator,
) {

	if handlerRunning {
		log.Warn("nethandler Start: Already running")
		return
	}

	log.Info("elevnetworkhandler Start: Starting")
	signalNetHandlerToStop = make(chan bool)
	handlerRunning = true

	go netHandler(signalNetHandlerToStop,
		orderQueueFromNethandler,
		buttonLightsFromNethandler,
		buttonPressesToNethandler,
		elevStateToNethandler)
}

/*StopNetHandler signals the nethandler instance to shut down
 */
func StopNetHandler() {
	if !handlerRunning {
		log.Warn("nethandler Stop: Handler was not runnning")
		return
	}
	log.Info("elevnetworkhandler StopNetHandler: Stopping")
	signalNetHandlerToStop <- true
	handlerRunning = false
	return
}

////////////////////////////////
// Auxiliary functions
////////////////////////////////

/*netHandler is the process which controls communication, orders, etc.
 * It communicates with both the local elevator (chans), and remote systems (UDP)
 * It sends and receives updates (messages) from/to other systems, which is how orders, order ACKS, timeouts, etc. are propagated
 * @arg signalNetHandlerToStop: Channel which is used to make the instance shut down
 * @arg orderQueueFromNetHandler: The queue the local elevator is to carry out is sent here
 * @arg buttonLightsFromNetHandler: Values for the button panel lights
 * @arg buttonPressesToNetHandler: Nethandler receives button presses from the local elevator here
 * @arg elevStateToNetHandler: Nethandler receives the floor/movDir/current action of the local elevator here
 */
func netHandler(
	signalNetHandlerToStop <-chan bool,
	orderQueueFromNethandler chan<- [et.NumFloors][et.NumButtons]et.SimpleOrder,
	buttonLightsFromNethandler chan<- et.ButtonLamp,
	buttonPressesToNethandler <-chan et.ButtonEvent,
	elevStateToNethandler <-chan et.Elevator,
) {
	// Used for sending to and recv from Transmitter and Receiver
	var sendRegularUpdates = make(chan et.ElevState, 12)
	var recvRegularUpdates = make(chan et.ElevState, 12)

	// Used for determining when to perform actions / send messages
	netHandlerDebugLogMsgTimer := time.Now()
	netHandlerDebugLogMsgFreq := 2 * time.Second

	netHandlerAutoBackupTimer := time.Now()
	netHandlerAutoBackupFreq := 500 * time.Millisecond

	netHandlerSendRegularUpdateTimer := time.Now()
	netHandlerSendRegularUpdateFreq := 100 * time.Millisecond

	netHandlerSendElevatorQueueTimer := time.Now()
	netHandlerSendElevatorQueueFreq := 50 * time.Millisecond

	netHandlerSendElevatorLightsTimer := time.Now()
	netHandlerSendElevatorLightsFreq := 200 * time.Millisecond

	// Regular update transmitter
	go b.Transmitter(et.AckHandlerPort, sendRegularUpdates)
	// Regular update receiver
	go b.Receiver(et.AckHandlerPort, recvRegularUpdates)

	// HearBeat for monitoring active systems
	go network.StartHeartBeat()
	defer network.StopHeartBeat()

	for {
		////////
		//// Receive messages
		////////
		select {
		// Net Handler Control
		case <-signalNetHandlerToStop:
			return
		// Updates from the local elevator
		case elev := <-elevStateToNethandler:
			ss.PushLocalElevatorUpdate(&elev)
		case newOrderButtonPress := <-buttonPressesToNethandler:
			log.WithField("btn", newOrderButtonPress).Debug("nethandler handler: recv button press")
			optSysID, err := eval.FindOptimalSystem(ss.GetActiveSystemsStates(), newOrderButtonPress)

			if err != nil {
				// already existing order, or <= 1 active systems
			} else {
				ss.PushButtonEvent(optSysID, newOrderButtonPress)
			}
		// Regular Update received over UDP
		case remoteElevStateUpdate := <-recvRegularUpdates:
			if remoteElevStateUpdate.ID != 0 && remoteElevStateUpdate.ID != ss.LocalID {

			}
			ss.HandleRegularUpdate(remoteElevStateUpdate)
		default:
		}

		////////
		//// Perform actions
		////////

		ss.CheckForAndHandleOrderTimeouts()

		// Backup the system on file. Also done each time we accept an order.
		if time.Now().Sub(netHandlerAutoBackupTimer) > netHandlerAutoBackupFreq {
			netHandlerAutoBackupTimer = time.Now()
			sb.Backup(ss.GetSystemsStates())
		}

		// Logs a message regularly, helpful to see when the system is running
		if time.Now().Sub(netHandlerDebugLogMsgTimer) > netHandlerDebugLogMsgFreq {
			netHandlerDebugLogMsgTimer = time.Now()
			log.Debug("nethandler handler: Running")
		}

		////////
		//// Send messages
		////////

		// Send regular updates over UDP
		if time.Now().Sub(netHandlerSendRegularUpdateTimer) > netHandlerSendRegularUpdateFreq {
			netHandlerSendRegularUpdateTimer = time.Now()
			select {
			case sendRegularUpdates <- ss.GetLocalSystem():
			default:
				log.Warn("nethandler Handler: Could not send regular update")
			}
		}

		// Send order queue to local elevator
		if time.Now().Sub(netHandlerSendElevatorQueueTimer) > netHandlerSendElevatorQueueFreq {
			netHandlerSendElevatorQueueTimer = time.Now()
			orders := ss.GetLocalSystemQueue()
			select {
			case orderQueueFromNethandler <- orders:
			default:
				log.Warn("nethandler Handler: failed to send order queue to elevator")
			}
		}

		// Send buttom lamp lights to local elevator
		if time.Now().Sub(netHandlerSendElevatorLightsTimer) > netHandlerSendElevatorLightsFreq {
			netHandlerSendElevatorLightsTimer = time.Now()
			lights := ss.GetPanelLights()
			for f := 0; f < et.NumFloors; f++ {
				for b := 0; b < et.NumButtons; b++ {
					select {
					case buttonLightsFromNethandler <- lights[f][b]:
					default:

					}
				}
			}
		}

	}
}
