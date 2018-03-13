package nethandler

import (
	"strconv"
	"time"

	network "../elevnetwork"
	b "../elevnetwork/bcast"
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
	var sendAckNack = make(chan et.AckNackMsg, 6)
	var recvAckNack = make(chan et.AckNackMsg, 6)
	var sendRegularUpdates = make(chan et.ElevState, 12)
	var recvRegularUpdates = make(chan et.ElevState, 12)

	go b.Transmitter(et.AckHandlerPort, sendAckNack, sendRegularUpdates)
	go b.Receiver(et.AckHandlerPort, recvAckNack, recvRegularUpdates)
	// Start Heartbeat
	go network.StartHeartBeat()
	defer network.StopHeartBeat()

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
				ss.UpdateSysElevator(optSysIndex, newOrderButtonPress)
				//log.WithField("sysid" /*ss.GetSystems()[*/, optSysIndex /*].ID*/).Debug("nethandler netHandler: New order, found optimal sys to take order")
				/*ordersDelegatedFromNetwork <- et.ElevOrder{
					Id:                strconv.FormatInt(time.Now().Unix(), 16),
					Order:             newOrderButtonPress,
					TimestampReceived: time.Now().Unix(),
					Status:            et.Accepted,
				}*/
			}
		case state := <-recvRegularUpdates:
			//'TODO
			//ss.SetSystems(states)
			log.Info("Recv regular update! :)")
			for floor_index := 0; floor_index < et.NumFloors; floor_index++ {
				for btn_index := 0; btn_index < 3; btn_index++ {
					if state.CurrentOrders[floor_index][btn_index].Id != "" {
						//@TODO actual logic. Here we dont even check if accepted
						ordersDelegatedFromNetwork <- state.CurrentOrders[floor_index][btn_index]
					}
					log.WithField("order", state.CurrentOrders[floor_index][btn_index]).Debug("nethandler handler: recv order")

				}

			}
			// Delegate this order and update netState
		default:
		}

		if time.Now().Sub(netHandlerDebugLogMsgTimer) > netHandlerDebugLogMsgFreq {
			netHandlerDebugLogMsgTimer = time.Now()
			log.Debug("nethandler handler: Running")
			s := ss.GetSystems()[0]
			s.CurrentOrders[0][0] = et.ElevOrder{
				Id:                strconv.FormatInt(time.Now().Unix(), 16),
				Order:             et.ButtonEvent{Floor: 1, Button: et.BT_HallDown},
				TimestampReceived: time.Now().Unix(),
				Status:            et.Accepted,
			}
			select {
			case sendRegularUpdates <- s:
				log.Info("Sent regular updates")
			case sendAckNack <- et.AckNackMsg{et.MsgACK, "hi"}:
			case a := <-recvAckNack:
				log.WithField("msg", a.MsgData).Info("recv ack")

			default:
			}
		}
	}
}
