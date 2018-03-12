package nethandler

import (
	"time"

	network "../elevnetwork"
	b "../elevnetwork/bcast"
	eval "../elevorderevaluation"
	timer "../elevtimer"
	et "../elevtype"
	sb "../sysbackup"
	ss "../sysstate"
	log "github.com/sirupsen/logrus"
)

var signalNetHandlerToStop chan bool

func StartNetHandler(
	networkToElev chan<- et.GeneralOrder,
	elevToNetwork <-chan et.ButtonEvent,
) {
	signalNetHandlerToStop = make(chan bool)
	go netHandler(signalNetHandlerToStop, networkToElev, elevToNetwork)
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
	networkToElev chan<- et.GeneralOrder,
	elevToNetwork <-chan et.ButtonEvent,
) {
	// Recover system backup
	backup, err := sb.Recover(time.Now().Add(-time.Second * 20))
	if err != nil {
		log.Info("nethandler netHandler: No backup to recover")
		backup = make([]et.ElevState, 1)
		// @TODO init state here
	}
	ss.SetSystems(backup)

	// Start Transmitter and Receiver for sending messages
	var sendAckNack = make(chan et.AckNackMsg, 6)
	var recvAckNack = make(chan et.AckNackMsg, 6)
	var sendRegularUpdates = make(chan et.ElevState, 6)
	var recvRegularUpdates = make(chan et.ElevState, 6)

	go b.Transmitter(et.AckHandlerPort, sendAckNack, sendRegularUpdates)
	go b.Receiver(et.AckHandlerPort, recvAckNack, recvRegularUpdates)
	// Start Heartbeat for
	go network.StartHeartBeat()
	defer network.StopHeartBeat()

	// Init netState with backup, if applicable
	// might be best to pass it as an argument to netHandler, which then pushes the necessary orders to the elevhandler?

	// Start ACK service (routine for sending and receiving ACKS)
	// defer stop ACK service

	netHandlerDebugLogMsgTimer := time.Now()
	netHandlerDebugLogMsgFreq := 2 * time.Second

	timer.StartDelayedFunction("ElevNetHandler Watchdog", time.Second*2, func() { panic("ElevHandler Watchdog: timeout") })
	defer timer.Stop("ElevHandler Watchdog")

	for {
		timer.Update("ElevNetHandler Watchdog", time.Second*3)

		// monitor ACK
		// if order ACK'd by all, update netState to Accepted
		// if order is for this Elev, push order to elevhandler

		// "Regular backup"
		//@TODO should this be called every loop?
		sb.Backup(ss.GetSystems())

		select {
		// Net Handler Control
		case <-signalNetHandlerToStop:
			return

		case newOrderButtonPress := <-elevToNetwork:
			optSysIndex := eval.DelegateOrder(ss.GetSystemElevators(), newOrderButtonPress)
			log.WithField("sysid", ss.GetSystems()[optSysIndex].ID).Debug("nethandler netHandler: New order, found optimal sys to take order")
			// Delegate this order and update netState

		}
		if time.Now().Sub(netHandlerDebugLogMsgTimer) > netHandlerDebugLogMsgFreq {
			netHandlerDebugLogMsgTimer = time.Now()
			log.Debug("elevhandler handler: Running")
		}
	}
}
