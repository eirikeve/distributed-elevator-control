package sysstate

import (
	"errors"
	"strconv"
	"time"

	network "../elevnetwork"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

const ackRetainTime = 30 * time.Second

/*UpdateLocalElevator contains logic for updating the state & orders of the local system
 */
func UpdateLocalElevator(e *et.Elevator) {
	//log.Debug("sysstate Update: Update local elevator")

	if !initialized {
		initSysState()
	}

	system, _ := systems[LocalID]
	system.E = *e
	systems[LocalID] = system

	updateFinishedOrders()
}

/*PushButtonEvent creates an order to handle the received btn event, but only if it is not yet an active order.
 *
 */
func PushButtonEvent(sysID int32, btn et.ButtonEvent) {

	if !initialized {
		initSysState()
	}

	t := time.Now().Unix()
	o := et.ElevOrder{
		Id:                LocalIP + strconv.FormatInt(int64(btn.Floor), 10) + strconv.FormatInt(int64(btn.Button), 10) + strconv.FormatInt(time.Now().Unix(), 30),
		Order:             btn,
		TimestampReceived: t,
		Status:            et.Received,
		TimestampLastOrderStatusChange: t,
		Assignee:                       sysID,
		SentToAssigneeElevator:         false,
	}

	if !isOrderAlreadyActive(btn) {
		if o.IsCabOrder() {
			o.Status = et.Accepted
			o.Acks = append(o.Acks, LocalID)
			o.Assignee = LocalID

			system, _ := systems[LocalID]
			system.CurrentOrders[btn.Floor][int(btn.Button)] = o
			systems[LocalID] = system

		} else {

			activeSystems := network.GetSystemsInNetwork()
			if len(activeSystems) < 2 {
				log.Warn("sysstate Push: Order rejected. Cannot guarantee completion")
				return
			}
			log.Debug("sysstate Push: Is new order!")

			o.Acks = append(o.Acks, LocalID)

			// Note that sysIndex says which system should perform the order.
			// But, since THIS system is the one delegating,
			// The order is stored in THIS system|s ElevState!
			// Other system's elevstate is only used for updating which orders
			// we will perform locally, we don't modify them

			// Verify that the system the order is assigned to exists
			_, ok := systems[sysID]
			// The order is saved in the currentOrders of the local elevator
			system, _ := systems[LocalID]
			if ok {
				system.CurrentOrders[btn.Floor][int(btn.Button)] = o
				systems[LocalID] = system
				log.WithField("o", o).Debug("sysstate Update: Set received order")
			} else {
				log.WithField("sysID", sysID).Error("sysstate Push: Order assigned to unknown system")
			}
		}

	}
	// @TODO check if only one sys is active, if so, do NOT accept order!!!
}

/*func RegisterAckNackMsg(msg et.AckNackMsg) {
	if msg.MsgType == et.MsgNACK {
		rejectOrder(msg.MsgData)
	} else {
		o, err := findOrder(msg.MsgData)
		if err != nil {
			log.WithError(err).Error("sysstate RegisterAck: Could not find any corresponding orders")
			return
		}
		switch o.Status {
		case et.Received:
			updateOrderAcksAfterMsg(o, msg)
			if canGuaranteeOrderCompletion(o) {

			}
			system, _ := systems[LocalIP]
			system.CurrentOrders[o.GetFloor()][int(o.GetButton())] = o
			systems[LocalIP] = system
			return
		case et.Accepted:
			fallthrough
		case et.Finished:
			return
		case et.Timeout:
			log.WithField("o", o).Warn("sysstate RegisterAck: Order has previously timed out!")
		case et.Unknown:
			fallthrough
		default:
			log.WithField("o", o).Error("sysstate RegisterAck: Unknown order status")
		}
	}
}*/

func HandleRegularUpdate(es et.ElevState) {

	if es.ID == LocalID || es.ID == 0 {
		return
	}

	oldEs, existsInSystems := systems[es.ID]

	if existsInSystems {
		if oldEs.StartupTime != es.StartupTime {
			notifySystemOfBackup(oldEs)
			// same backup
		} else {
			ok := verifyOrderChangesOk(es, oldEs)
			if !ok {
				// handle!!
				log.Error("sysstate HandleRegularUpdate: Can not verify order changes")
			}
		}
	}

	systems[es.ID] = es
	//localSys := systems[LocalIP]

	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			/*if localSys.CurrentOrders[f][b].Id != "" || es.CurrentOrders[f][b].Id != "" {
				log.WithFields(log.Fields{
					"f":                       f,
					"b":                       b,
					"LocalOrder":              localSys.CurrentOrders[f][b].Id,
					"LocalOrderLastTimeChng":  localSys.CurrentOrders[f][b].TimestampLastOrderStatusChange,
					"RemoteOrder":             es.CurrentOrders[f][b].Id,
					"RemoteOrderLastTimeChng": es.CurrentOrders[f][b].TimestampLastOrderStatusChange,
				}).Debug("sysstate HandleRegularUpdate: Comparison of orders")
			}*/

		}
	}

	applyUpdatesToLocalSystem(es)
	acceptOrdersWeCanGuarantee()
	//@TODO backup here
	sendAckMessages()
}

func notifySystemOfBackup(es et.ElevState) {
	// called when that elevator has restarted
}

func verifyOrderChangesOk(es et.ElevState, oldEs et.ElevState) bool {
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			// Only care about where orders may have been lost, so if there was no order in the old ElevState, skip
			if !(oldEs.CurrentOrders[f][b].GetID() == "") {
				// Check if IDs are different
				if !(es.CurrentOrders[f][b].GetID() == oldEs.CurrentOrders[f][b].GetID()) {
					// Since IDs are different, check if the order in the old ElevState has since been finished
					if !isOrderAlreadyFinished(es, oldEs.CurrentOrders[f][b].GetID()) {
						// If Prev order was accepted, we have removed an order that MUST be completed, which is a problem.
						if oldEs.CurrentOrders[f][b].IsAccepted() {
							return false
						}
					}
				}
			}

		}
	}
	return true
}

func isOrderAlreadyFinished(es et.ElevState, orderID string) bool {
	if orderID == "" {
		return false
	}
	for _, order := range es.FinishedOrders {
		if order.Id == orderID {
			return true
		}
	}
	return false
}

func applyUpdatesToLocalSystem(es et.ElevState) {
	//localSys := GetLocalSystem()
	/*print("Finished Orders before merge: ")
	for _, order := range localSys.FinishedOrders {
		print(order.Id, " ")
	}
	println()*/
	mergeFinishedOrdersQueue(es)
	//localSys = GetLocalSystem()
	/*print("Finished Orders after merge: ")
	for _, order := range localSys.FinishedOrders {
		print(order.Id, " ")
	}
	println()*/
	mergeOrdersToLocalSystem(es)
	//localSys := systems[LocalID]
	//println("localSystem f", 0, " b", 0, "=", localSys.CurrentOrders[0][0].Id, "stat:", localSys.CurrentOrders[0][0].Status)
	addLocalAckToOrders()
	applyRemoteOrderAckLogicalOR(es)
	//println("localSystem f", 0, " b", 0, "=", localSys.CurrentOrders[0][0].Id, "stat:", localSys.CurrentOrders[0][0].Status)
}

func mergeOrdersToLocalSystem(es et.ElevState) {
	localSystem, _ := systems[LocalID]
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			o, err := updateSingleOrder(&es, localSystem.CurrentOrders[f][b], es.CurrentOrders[f][b])
			if err != nil {
				//handle
			} else {
				localSystem.CurrentOrders[f][b] = o
				//println("localSystem f", f, " b", b, "=", o.Id)
			}
		}
	}
	systems[LocalID] = localSystem

}

func mergeFinishedOrdersQueue(remoteSystem et.ElevState) {
	localSystem := systems[LocalID]
	// Check if the remote order is not in our finished orders list
	for _, oRemote := range remoteSystem.FinishedOrders {
		isInList := false
		for _, oLocal := range localSystem.FinishedOrders {
			if oLocal.Id == oRemote.Id {
				isInList = true
			}
		}
		if !isInList {
			// Place the remote order in our finished order list ONLY if:
			// it was finished more recently than the oldest order in the current finished order list of localSystem
			indexOfOldestOrder := 0
			for index, oLocal := range localSystem.FinishedOrders {
				if oLocal.TimestampLastOrderStatusChange < localSystem.FinishedOrders[indexOfOldestOrder].TimestampLastOrderStatusChange {
					indexOfOldestOrder = index
				}
			}
			if oRemote.TimestampLastOrderStatusChange > localSystem.FinishedOrders[indexOfOldestOrder].TimestampLastOrderStatusChange {
				localSystem.FinishedOrders[indexOfOldestOrder] = oRemote
			}
		}
	}

	systems[LocalID] = localSystem
}

func updateSingleOrder(remoteSystem *et.ElevState, localOrder et.ElevOrder, remoteOrder et.ElevOrder) (et.ElevOrder, error) {
	var err error
	var o et.ElevOrder
	localSystem := systems[LocalID]

	if localOrder.IsCabOrder() || remoteOrder.IsCabOrder() {
		return localOrder, nil
	}
	// PRINTS
	/*if localOrder.Order.Floor == 1 && localOrder.Order.Button == et.BT_HallUp && localOrder.Id != "" ||
		remoteOrder.Order.Floor == 1 && remoteOrder.Order.Button == et.BT_HallUp && remoteOrder.Id != "" {
		println("floor 1, hallup")
		println(" Remote order finished?", isOrderAlreadyFinished(localSystem, remoteOrder.Id))
		println(" Local order finished?", isOrderAlreadyFinished(localSystem, localOrder.Id))
		println("..end")
	} else if localOrder.Order.Floor == 0 && localOrder.Order.Button == et.BT_HallUp && localOrder.Id != "" ||
		remoteOrder.Order.Floor == 0 && remoteOrder.Order.Button == et.BT_HallUp && remoteOrder.Id != "" {
		println("floor 0, hallup")
		println(" Remote order finished?", isOrderAlreadyFinished(localSystem, remoteOrder.Id))
		println(" Local order finished?", isOrderAlreadyFinished(localSystem, localOrder.Id))
		println("..end")
	}*/

	// If the order {f, b} is finished either locally and/or remotely, we don't need any complex logic:
	remoteFinished := isOrderAlreadyFinished(localSystem, remoteOrder.Id)
	localFinished := isOrderAlreadyFinished(localSystem, localOrder.Id)
	if localFinished && remoteFinished {
		o = et.EmptyOrder()
		return o, nil
	} else if localFinished {
		o = remoteOrder
		return o, nil
	} else if remoteFinished {
		o = localOrder
		return o, nil
	}

	// Neither the local nor the remote order is finished. But, one may be empty, which would simplify:
	if localOrder.IsEmpty() {
		//println("Local empty. Returning remote order:", remoteOrder.Id)
		return remoteOrder, nil
	} else if remoteOrder.IsEmpty() {
		//println("Remote emtpy. Returning local order:", localOrder.Id)
		return localOrder, nil
	}

	// So, we have two orders in the same floor/button. Likely, it is the same order ID.
	// Still, though different IDs should in theory never occur, as we know, it probably will.
	if localOrder.GetID() != remoteOrder.GetID() {
		log.WithFields(log.Fields{"ID1": localOrder.GetID(), "ID2": remoteOrder.GetID()}).Error("sysstate updateSingleOrder: Non-matching IDs")
		err = errors.New("sysstate updateSingleOrder: Non-matching IDs: " + localOrder.GetID() + ", " + remoteOrder.GetID())
		if localOrder.IsAccepted() && !remoteOrder.IsAccepted() {
			o = localOrder
		} else if !localOrder.IsAccepted() && remoteOrder.IsAccepted() {
			o = remoteOrder
			// Both orders were accepted. However, since they are are the same ButtonEvent (see call in mergeOrdersToLocalSystem),
			// the order will be performed no matter which order is stored in the queues
		} else if localOrder.IsAccepted() && remoteOrder.IsAccepted() {

			if localOrder.TimestampLastOrderStatusChange > remoteOrder.TimestampLastOrderStatusChange {
				o = localOrder
			} else {
				o = remoteOrder
			}
			log.WithFields(log.Fields{"retainedOrderId": o.Id, "localOrderId": localOrder.Id, "remoteOrderId": remoteOrder.Id}).Warn("sysstate updateSingleOrder: Two accepted conflicting orders, retaining one")
		} else {
			if localOrder.TimestampLastOrderStatusChange > remoteOrder.TimestampLastOrderStatusChange {
				o = localOrder
			} else {
				o = remoteOrder
			}
		}
	} else {
		// Same order ID
		if localOrder.IsAccepted() && remoteOrder.IsAccepted() {
			if localOrder.TimestampLastOrderStatusChange > remoteOrder.TimestampLastOrderStatusChange {
				o = localOrder
			} else {
				o = remoteOrder
			}
		} else if localOrder.IsAccepted() {
			o = localOrder
		} else if remoteOrder.IsAccepted() {
			o = remoteOrder
		} else {
			if localOrder.TimestampLastOrderStatusChange > remoteOrder.TimestampLastOrderStatusChange {
				o = localOrder
			} else {
				o = remoteOrder
			}
		}
	}

	return o, err
}

func addLocalAckToOrders() {
	localSystem, _ := systems[LocalID]
	activeSystems := network.GetSystemsInNetwork()

	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if localSystem.CurrentOrders[f][b].Id != "" {
				alreadyRegistered := false
				for _, ack := range localSystem.CurrentOrders[f][b].Acks {
					if ack == LocalID {
						alreadyRegistered = true
					}
				}
				if !alreadyRegistered {
					if contains(activeSystems, localSystem.CurrentOrders[f][b].Assignee) {
						localSystem.CurrentOrders[f][b].Acks = append(localSystem.CurrentOrders[f][b].Acks, LocalID)
						acksForBroadcasting = append(acksForBroadcasting, et.AckNackMsg{MsgType: et.MsgACK, MsgData: localSystem.CurrentOrders[f][b].Id, MsgSender: LocalIP})
					}
				}
			}

		}
	}
	systems[LocalID] = localSystem
}

func applyRemoteOrderAckLogicalOR(es et.ElevState) {
	localSystem := systems[LocalID]
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if es.CurrentOrders[f][b].Id == localSystem.CurrentOrders[f][b].Id {
				newAcks := getAcksOnlyRegisteredRemotely(localSystem.CurrentOrders[f][b], es.CurrentOrders[f][b])
				for _, ack := range newAcks {
					localSystem.CurrentOrders[f][b].Acks = append(localSystem.CurrentOrders[f][b].Acks, ack)
				}
			}
		}
	}
	systems[LocalID] = localSystem

}

func contains(container []int32, element int32) bool {
	for _, elem := range container {
		if elem == element { // 0 if equal
			return true
		}
	}
	return false
}

func getAcksOnlyRegisteredRemotely(local et.ElevOrder, remote et.ElevOrder) []int32 {
	var acks []int32

	for _, remoteAck := range remote.Acks {
		ackAlreadyRegisteredLocally := false
		for _, localAck := range local.Acks {
			if remoteAck == localAck {
				ackAlreadyRegisteredLocally = true
			}
		}
		if !ackAlreadyRegisteredLocally {
			acks = append(acks, remoteAck)
		}
	}
	return acks
}

func acceptOrdersWeCanGuarantee() {
	localSystem, _ := systems[LocalID]
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if localSystem.CurrentOrders[f][b].Status == et.Received &&
				canGuaranteeOrderCompletion(localSystem.CurrentOrders[f][b]) {
				log.WithField("o", localSystem.CurrentOrders[f][b]).Debug("sysstate acceptOrders: Can guarantee order; accepting")
				accept(&localSystem, localSystem.CurrentOrders[f][b])
			}
		}
	}
	systems[LocalID] = localSystem
}
func sendAckMessages() {
	// seeeeeeeeeeeeend
}
func rejectOrder(orderID string) {
	s, _ := systems[LocalID]

	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if s.CurrentOrders[f][b].GetID() == orderID && // Found order to reject
				!s.CurrentOrders[f][b].IsAccepted() { // Only reject if we have not yet accepted!

				s.CurrentOrders[f][b] = et.EmptyOrder()
			}
		}

	}
	systems[LocalID] = s

}

/*func updateOrderAcksAfterMsg(o et.ElevOrder, msg et.AckNackMsg) {
	ackAlreadyRegistered := false
	for _, id := range o.Acks {
		if id == msg.MsgSender {
			ackAlreadyRegistered = true
		}
	}
	if !ackAlreadyRegistered {
		o.Acks = append(o.Acks, msg.MsgSender)
		o.TimestampLastOrderStatusChange = time.Now().Unix()
	}
}
*/
func canGuaranteeOrderCompletion(o et.ElevOrder) bool {
	c := countOrderOccurrencesInSystems(o)
	if c >= 2 {
		return true
	}
	return false

}
func countOrderOccurrencesInSystems(o et.ElevOrder) int {
	count := 0

	for _, sysID := range o.Acks {
		system, exists := systems[sysID]
		if exists && system.CurrentOrders[o.GetFloor()][int(o.GetButton())].GetID() == o.GetID() {
			count++
		}
	}
	return count
}

/*updateFinishedOrders checks which orders have been finished by the local elevator
 * these orders are marked as status Finished and moved to the FinishedOrderList of the local system.
 */
func updateFinishedOrders() {
	if !initialized {
		initSysState()
	}
	s, _ := systems[LocalID]

	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if s.CurrentOrders[f][b].Assignee == LocalID && // Check that this elevator is supposed to carry out the order
				s.CurrentOrders[f][b].SentToAssigneeElevator && // Check that the order has been sent to the elevator FSM
				s.CurrentOrders[f][b].IsAccepted() && // Check if the order has been accepted
				s.E.Orders[f][b].IsEmpty() { // Check that the elevator FSM has carried out the order
				markOrderFinished(&s, f, b)
				putOrderInFinishedOrdersList(&s, f, b)
			}
		}
	}
	systems[LocalID] = s

}

func accept(localSys *et.ElevState, o et.ElevOrder) {
	(*localSys).CurrentOrders[o.GetFloor()][int(o.GetButton())].Status = et.Accepted
	(*localSys).CurrentOrders[o.GetFloor()][int(o.GetButton())].TimestampLastOrderStatusChange = time.Now().Unix()
}

func findOrder(orderID string) (et.ElevOrder, error) {
	var o et.ElevOrder = et.EmptyOrder()
	var err error
	for _, system := range systems {
		for f := 0; f < et.NumFloors; f++ {
			for b := 0; b < et.NumButtons; b++ {
				if system.CurrentOrders[f][b].Id == orderID {
					if system.CurrentOrders[f][b].TimestampLastOrderStatusChange > o.TimestampLastOrderStatusChange {
						o = system.CurrentOrders[f][b]
					}
				}

			}
		}
		for _, order := range system.FinishedOrders {
			if order.Id == orderID {
				// Here we return instantly since the order has already been finished
				return order, nil
			}
		}
	}
	if o.IsEmpty() {
		err = errors.New("sysstatelogic findOrder: cannot find order " + orderID)
	}
	return o, err
}

func markOrderFinished(es *et.ElevState, floor int, button int) {
	es.CurrentOrders[floor][button].Status = et.Finished
	es.CurrentOrders[floor][button].TimestampLastOrderStatusChange = time.Now().Unix()
}
func putOrderInFinishedOrdersList(s *et.ElevState, floor int, button int) {
	indexOfOldestOrder := 0
	for index, order := range (*s).FinishedOrders {
		if order.TimestampLastOrderStatusChange < (*s).FinishedOrders[indexOfOldestOrder].TimestampLastOrderStatusChange {
			indexOfOldestOrder = index
		}
	}
	s.FinishedOrders[indexOfOldestOrder] = s.CurrentOrders[floor][button]
	s.CurrentOrders[floor][button] = et.EmptyOrder()
}

func isLocalOrder(o et.ElevOrder) bool {
	return o.Assignee == LocalID
}

// Confusing name.... checks if an order already exists
func isOrderAlreadyActive(btn et.ButtonEvent) bool {
	system := systems[LocalID]
	if system.CurrentOrders[btn.Floor][int(btn.Button)].IsActive() {
		log.WithField("Order", system.CurrentOrders[btn.Floor][int(btn.Button)]).Debug("sysstate isAlreadyActive: This is the order which registers as the same")
		return true
	}
	return false
}
