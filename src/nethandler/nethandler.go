package nethandler

import (
	"time"

	timer "../elevtimer"
	et "../elevtype"
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
	// Start Heartbeat
	// defer stop Heartbeat

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

		// send "Regular updates"

		// send ACK if needed

		// monitor ACK
		// if order ACK'd by all, update netState to Accepted
		// if order is for this Elev, push order to elevhandler

		// "Regular backup"

		select {
		// Net Handler Control
		case <-signalNetHandlerToStop:
			return

		case _ = <-elevToNetwork:
			// Delegate this order and update netState <- Update netState?
			// [@SUGGESTION]: Send to orderDelegation, to decide which elevator should take the order

		}
		if time.Now().Sub(netHandlerDebugLogMsgTimer) > netHandlerDebugLogMsgFreq {
			netHandlerDebugLogMsgTimer = time.Now()
			log.Debug("elevhandler handler: Running")
		}
	}
}
