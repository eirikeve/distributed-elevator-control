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

var signalNetHandlerToStop chan bool

func StartNetHandler(
	orderQueueFromNethandler chan<- [et.NumFloors][et.NumButtons]et.SimpleOrder,
	buttonLightsFromNethandler chan<- et.ButtonLamp,
	buttonPressesToNethandler <-chan et.ButtonEvent,
	elevStateToNethandler <-chan et.Elevator,
) {
	signalNetHandlerToStop = make(chan bool)
	go netHandler(signalNetHandlerToStop,
		orderQueueFromNethandler,
		buttonLightsFromNethandler,
		buttonPressesToNethandler,
		elevStateToNethandler)
}
func StopNetHandler() {
	log.Info("elevnetworkhandler StopNetHandler: Stopping")
	signalNetHandlerToStop <- true
	log.Debug("elevnetworkhandler StopNetHandler: Stop signalled")
	return
}

func netHandler(
	signalNetHandlerToStop <-chan bool,
	orderQueueFromNethandler chan<- [et.NumFloors][et.NumButtons]et.SimpleOrder,
	buttonLightsFromNethandler chan<- et.ButtonLamp,
	buttonPressesToNethandler <-chan et.ButtonEvent,
	elevStateToNethandler <-chan et.Elevator,
) {

	var sendRegularUpdates = make(chan et.ElevState, 12)
	var recvRegularUpdates = make(chan et.ElevState, 12)

	go b.Transmitter(et.AckHandlerPort, sendRegularUpdates)
	go b.Receiver(et.AckHandlerPort, recvRegularUpdates)
	go network.StartHeartBeat()
	defer network.StopHeartBeat()

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

	for {
		// Receive messages
		select {
		// Net Handler Control
		case <-signalNetHandlerToStop:
			return

		case elev := <-elevStateToNethandler:
			ss.UpdateLocalElevator(&elev)
		case newOrderButtonPress := <-buttonPressesToNethandler:
			log.WithField("btn", newOrderButtonPress).Debug("nethandler handler: recv button press")

			optSysID, err := eval.FindOptimalSystem(ss.GetActiveSystemsStates(), newOrderButtonPress)

			if err != nil {
				// already existing order, or <= 1 active systems
			} else {
				ss.PushButtonEvent(optSysID, newOrderButtonPress)
			}
		case remoteElevStateUpdate := <-recvRegularUpdates:
			if remoteElevStateUpdate.ID != 0 && remoteElevStateUpdate.ID != ss.LocalID {

			}

			ss.HandleRegularUpdate(remoteElevStateUpdate)
		default:
		}

		ss.CheckForAndHandleOrderTimeouts()

		// Send messages
		if time.Now().Sub(netHandlerSendRegularUpdateTimer) > netHandlerSendRegularUpdateFreq {
			netHandlerSendRegularUpdateTimer = time.Now()
			select {
			case sendRegularUpdates <- ss.GetLocalSystem():
			default:
				log.Warn("nethandler Handler: Could not send regular update")
			}
		}
		if time.Now().Sub(netHandlerSendElevatorQueueTimer) > netHandlerSendElevatorQueueFreq {
			netHandlerSendElevatorQueueTimer = time.Now()
			orders := ss.GetLocalSystemOrders()
			select {
			case orderQueueFromNethandler <- orders:
			default:
				log.Warn("nethandler Handler: failed to send order queue to elevator")
			}
		}

		if time.Now().Sub(netHandlerDebugLogMsgTimer) > netHandlerDebugLogMsgFreq {
			netHandlerDebugLogMsgTimer = time.Now()
			log.Debug("nethandler handler: Running")
		}
		if time.Now().Sub(netHandlerAutoBackupTimer) > netHandlerAutoBackupFreq {
			netHandlerAutoBackupTimer = time.Now()
			sb.Backup(ss.GetSystemsStates())
		}

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
