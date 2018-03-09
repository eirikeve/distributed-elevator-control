package elevnetworkhandler

import log "github.com/sirupsen/logrus"
import et "../elevtype"

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
	return
}
