package nethandler

import (
	"strconv"
	"time"

	eval "../elevorderevaluation"
	et "../elevtype"
	ss "../sysstate"
	log "github.com/sirupsen/logrus"
)

var signalNetHandlerToStop chan bool

func StartNetHandler(
	ordersDelegatedFromNetwork chan<- et.GeneralOrder,
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
	ordersDelegatedFromNetwork chan<- et.GeneralOrder,
	buttonPressesToNetwork <-chan et.ButtonEvent,
	elevStateToNetwork <-chan et.Elevator,
) {
	// Start Transmitter and Receiver for sending messages
	//var sendAckNack = make(chan et.AckNackMsg, 6)
	//var recvAckNack = make(chan et.AckNackMsg, 6)
	//var sendRegularUpdates = make(chan et.ElevState, 6)
	//var recvRegularUpdates = make(chan et.ElevState, 6)

	//go b.Transmitter(et.AckHandlerPort, sendAckNack, sendRegularUpdates)
	//go b.Receiver(et.AckHandlerPort, recvAckNack, recvRegularUpdates)
	// Start Heartbeat for
	//go network.StartHeartBeat()
	//defer network.StopHeartBeat()

	// Init netState with backup, if applicable
	// might be best to pass it as an argument to netHandler, which then pushes the necessary orders to the elevhandler?

	// Start ACK service (routine for sending and receiving ACKS)
	// defer stop ACK service

	netHandlerDebugLogMsgTimer := time.Now()
	netHandlerDebugLogMsgFreq := 2 * time.Second

	//timer.StartDelayedFunction("ElevNetHandler Watchdog", time.Second*2, func() { panic("ElevNetHandler Watchdog: timeout") })
	//defer timer.Stop("ElevNetHandler Watchdog")

	for {
		//timer.Update("ElevNetHandler Watchdog", time.Second*3)

		// monitor ACK
		// if order ACK'd by all, update netState to Accepted
		// if order is for this Elev, push order to elevhandler

		// "Regular backup"
		//@TODO should this be called every loop?
		//sb.Backup(ss.GetSystems())
		select {
		// Net Handler Control
		case <-signalNetHandlerToStop:
			return

		case elev := <-elevStateToNetwork:
			ss.UpdateLocalElevator(&elev)
			//log.WithField("e", ss.GetSystemElevators()[0]).Debug("updated local elev:")

		case newOrderButtonPress := <-buttonPressesToNetwork:
			log.WithField("btn", newOrderButtonPress).Debug("nethandler handler: recv button press")
			optSysIndex, err := eval.DelegateOrder(ss.GetSystemElevators(), newOrderButtonPress)
			if err != nil {
				// already existing order
			} else {
				log.WithField("sysid" /*ss.GetSystems()[*/, optSysIndex /*].ID*/).Debug("nethandler netHandler: New order, found optimal sys to take order")
				ordersDelegatedFromNetwork <- et.ElevOrder{
					Id:                strconv.FormatInt(time.Now().Unix(), 16),
					Order:             newOrderButtonPress,
					TimestampReceived: time.Now().Unix(),
					Status:            et.Accepted,
				}
			}

			// Delegate this order and update netState
		default:
		}

		if time.Now().Sub(netHandlerDebugLogMsgTimer) > netHandlerDebugLogMsgFreq {
			netHandlerDebugLogMsgTimer = time.Now()
			log.Debug("nethandler handler: Running")
		}
	}
}
