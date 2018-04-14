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
	ordersDelegatedFromNetwork chan<- [et.NumFloors][et.NumButtons]et.SimpleOrder,
	buttonPressesToNetwork <-chan et.ButtonEvent,
	elevStateToNetwork <-chan et.Elevator,
) {
	signalNetHandlerToStop = make(chan bool)
	go netHandler(signalNetHandlerToStop,
		ordersDelegatedFromNetwork,
		buttonPressesToNetwork,
		elevStateToNetwork)
}
func StopNetHandler() {
	log.Info("elevnetworkhandler StopNetHandler: Stopping")
	//@BUG this does not send
	signalNetHandlerToStop <- true
	log.Debug("elevnetworkhandler StopNetHandler: Stop signalled")
	return
}

func netHandler(
	signalNetHandlerToStop <-chan bool,
	ordersDelegatedFromNetwork chan<- [et.NumFloors][et.NumButtons]et.SimpleOrder,
	buttonPressesToNetwork <-chan et.ButtonEvent,
	elevStateToNetwork <-chan et.Elevator,
) {
	// Start Transmitter and Receiver for sending messages
	var sendAckNack = make(chan et.AckNackMsg, 6)
	var recvAckNack = make(chan et.AckNackMsg, 6)
	var sendRegularUpdates = make(chan et.ElevState, 12)
	var recvRegularUpdates = make(chan et.ElevState, 12)

	go b.Transmitter(et.AckHandlerPort, sendAckNack, sendRegularUpdates)
	go b.Receiver(et.AckHandlerPort, recvAckNack, recvRegularUpdates)
	// Start Heartbeat
	go network.StartHeartBeat()
	defer network.StopHeartBeat()

	// Start ACK service (routine for sending and receiving ACKS)
	// defer stop ACK service

	netHandlerDebugLogMsgTimer := time.Now()
	netHandlerDebugLogMsgFreq := 2 * time.Second

	netHandlerAutoBackupTimer := time.Now()
	netHandlerAutoBackupFreq := 500 * time.Millisecond

	netHandlerSendRegularUpdateTimer := time.Now()
	netHandlerSendRegularUpdateFreq := 100 * time.Millisecond

	netHandlerSendElevatorQueueTimer := time.Now()
	netHandlerSendElevatorQueueFreq := 100 * time.Millisecond

	//timer.StartDelayedFunction("ElevNetHandler Watchdog", time.Second*2, func() { panic("ElevNetHandler Watchdog: timeout") })
	//defer timer.Stop("ElevNetHandler Watchdog")

	for {
		// Receive messages
		select {
		// Net Handler Control
		case <-signalNetHandlerToStop:
			return

		case elev := <-elevStateToNetwork:
			ss.UpdateLocalElevator(&elev)
			//log.WithField("e", ss.GetSystemElevators()[0]).Debug("updated local elev:")
		case newOrderButtonPress := <-buttonPressesToNetwork:
			log.WithField("btn", newOrderButtonPress).Debug("nethandler handler: recv button press")

			optSysID, err := eval.FindOptimalSystem(ss.GetActiveSystemsStates(), newOrderButtonPress)

			if err != nil {
				// already existing order
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
			case ordersDelegatedFromNetwork <- orders:
			default:
				log.Warn("nethandler Handler: failed to send order queue to elevator")
			}
			// There maay be a bug due to this not being called for sent orders, but I don't think so.
			//ss.MarkOrdersAsSent(sentOrders)

		}

		if time.Now().Sub(netHandlerDebugLogMsgTimer) > netHandlerDebugLogMsgFreq {
			netHandlerDebugLogMsgTimer = time.Now()
			// @TODO not correct
			log.Debug("nethandler handler: Running")
		}
		if time.Now().Sub(netHandlerAutoBackupTimer) > netHandlerAutoBackupFreq {
			netHandlerAutoBackupTimer = time.Now()
			sb.Backup(ss.GetSystemsStates())
		}

	}
}
