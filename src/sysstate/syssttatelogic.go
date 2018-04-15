package sysstate

import (
	"strconv"
	"time"

	network "../elevnetwork"
	et "../elevtype"
	sb "../sysbackup"
	log "github.com/sirupsen/logrus"
)

/*

sysstatelogic.go contains methods for updating finished orders, acknowledging orders,
accepting/rejecting orders, and incorporating remote changes into this system's queue.

*/

////////////////////////////////
// Interface
////////////////////////////////

/*PushLocalElevatorUpdate updates the current state of the local FSM in systems
 * This new update is also used to determine which orders have been completed.
 * @arg e: Most recent state (including queue & finished orders) of the local FSM
 */
func PushLocalElevatorUpdate(e *et.Elevator) {

	system, _ := systems[LocalID]
	system.E = *e
	systems[LocalID] = system

	updateFinishedOrders()
}

/*PushButtonEvent creates an order to handle received buttonEvents
 * But - duplicate orders are ignored, and orders may be rejected.
 * @arg assigneeSysID: ID of the system which this order will be delegated to if accepted
 * @arg btn: Floor/button corresponding to the order
 */
func PushButtonEvent(assigneeSysID int32, btn et.ButtonEvent) {

	t := time.Now().Unix()
	o := et.ElevOrder{
		Id:                strconv.FormatInt(int64(LocalID), 10) + strconv.FormatInt(int64(btn.Floor), 10) + strconv.FormatInt(int64(btn.Button), 10) + strconv.FormatInt(time.Now().Unix(), 30),
		Order:             btn,
		TimestampReceived: t,
		Status:            et.Received,
		TimestampLastOrderStatusChange: t,
		Assignee:                       assigneeSysID,
		CountTimesTimeout:              0,
	}

	// Ignore orders for floor/btn combinations which are already Received/Accepted.
	if !isOrderAlreadyActive(btn) {
		// Due to sysbackup we don't lose orders even in case of shutdowns. So we always accept cab orders.
		if o.IsCabOrder() {
			pushCabOrder(o)
		} else {
			pushHallOrder(o)
		}

	}
}

/*HandleRegularUpdate incorporates changes from another system into our system.
 *This is the way we distribute order information - including ACKs, order delegation, etc.
 * @arg
 */
func HandleRegularUpdate(remote et.ElevState) {

	// Ignore updates which are empty or sent locally
	if remote.ID == LocalID || remote.ID == 0 {
		return
	}

	// If we already have that ID in systems, check that no unexpected changes have occured
	oldRemote, existsInSystems := systems[remote.ID]
	if existsInSystems {
		if oldRemote.StartupTime == remote.StartupTime {
			ok := verifyNoLostAcceptedOrders(remote, oldRemote)
			if !ok {
				// Has not occurred once yet. Logging in case of large, system-breaking bugs.
				log.Error("sysstate HandleRegularUpdate: Can not verify order changes. ")
			}
		}
	}

	systems[remote.ID] = remote
	applyUpdatesToLocalSystem(remote)
	acceptOrdersWeCanGuarantee()
}

/*CheckForAndHandleOrderTimeout finds all orders in our queue which have timed out, and redelegates them
 * New Assignee is chosen from the list of elevators which have ACK'd the order.
 * Redelegation is cyclic & deterministic.
 */
func CheckForAndHandleOrderTimeouts() {
	sys := systems[LocalID]

	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			o := sys.CurrentOrders[f][b]
			if o.TimeSinceTimeout() > 0 {
				handleSingleOrderTimeout(&sys, o)
				// Reassignment only for logging
				o := sys.CurrentOrders[f][b]
				log.WithFields(log.Fields{"orderID": o.Id, "newAssignee": o.Assignee}).Info("Order timeout")
			}
		}
	}
	systems[LocalID] = sys
}

////////////////////////////////
// Auxiliary functions
////////////////////////////////

/*pushCabOrder pushes an cab order to the queue (as Accepted) as long as we believe we can finish it.
 * If we have several timeouts in the current cab orders, then the elevator might be malfunctioning.
 * In that case, we will not accept the order.
 * @arg o: A cab order
 */
func pushCabOrder(o et.ElevOrder) {
	c := countTimeoutsOfCabOrders()
	if c <= et.MaxAcceptableCountCabTimeouts {
		o.Status = et.Accepted
		o.Acks = append(o.Acks, LocalID)
		o.Assignee = LocalID
		o.TimestampLastOrderStatusChange = time.Now().Unix()

		system, _ := systems[LocalID]
		system.CurrentOrders[o.Order.Floor][int(o.Order.Button)] = o
		systems[LocalID] = system

		sb.Backup(GetSystemsStates())
		log.WithField("o", o).Debug("sysstate Update: Accepted cab order")
	} else {
		log.WithField("countTimeoutsOfCabOrders", c).Warn("sysstate Push: Rejected cab order due to many timeouts in current cab orders. Retry after they are Finished.")
	}

}

/*pushHallOrder pushes a hall order to the queue (only as Received) as long as there are other systems in the network.
 * Note that the order will not be Accepted (and thus carried out) before it is acknowledged by at least 1 other system
 * @arg o: A hall order
 */
func pushHallOrder(o et.ElevOrder) {
	// For hall orders, we need redundancy in order to accept orders.
	activeSystems := network.GetSystemsInNetwork()

	if len(activeSystems) < 2 {
		log.Warn("sysstate Push: Order rejected. Cannot guarantee completion due to few active systems")
		return
	}

	o.Acks = append(o.Acks, LocalID)

	// Verify that the system the order is assigned to exists
	_, ok := systems[o.Assignee]
	// The order is saved in the currentOrders of the local elevator (but with assigneeSysID as the assignee)
	system, _ := systems[LocalID]
	if ok {
		system.CurrentOrders[o.Order.Floor][int(o.Order.Button)] = o
		systems[LocalID] = system
		log.WithField("o", o).Debug("sysstate Update: Set received order")
	} else {
		log.WithField("sysID", o.Assignee).Error("sysstate Push: Order ignored due to unknown assignee")
	}
}

/*handleSingleOrderTimeout is used to redelegate orders when they time out.
 * Non-accepted orders are simply rejected (deleted).
 * For accepted orders, new assignee is chosen from the list of elevators which have ACK'd it.
 * @arg localSys: The local ElevState
 * @arg o: Order which has timed out
 */
func handleSingleOrderTimeout(localSys *et.ElevState, o et.ElevOrder) {
	// We haven't accepted the order, so we remove it from the queue.
	if o.Status == et.Received {
		empty(localSys, o)
		// We have accepted the order, so we redelegate it to the next system which has ACK'd it
	} else if o.Status == et.Accepted {

		o.CountTimesTimeout += 1

		if len(o.Acks) > 0 {
			indexOfNewAssignee := 0
			for index, sysID := range o.Acks {
				if sysID == o.Assignee {
					indexOfNewAssignee = index + 1
					break
				}
			}
			if indexOfNewAssignee >= len(o.Acks) {
				indexOfNewAssignee = 0
			}
			if o.Acks[indexOfNewAssignee] == o.Assignee {
				// Could not find any other system to take this order (most likely a cab order)
				redelegate(localSys, o, LocalID)
			} else {
				redelegate(localSys, o, o.Acks[indexOfNewAssignee])
			}
		} else {
			redelegate(localSys, o, LocalID)
		}
	} else {
		// Order was neither received nor accepted.
		empty(localSys, o)
	}
}

/*verifyOrderChangesOk checks that no accepted orders were removed without being finished
 * @arg remoteSys: current state of remote system
 * @arg prevRemoteSys: Previously registered state of remote system
 * @return true if all OK, false if some accepted order was removed without being finished
 */
func verifyNoLostAcceptedOrders(remoteSys et.ElevState, prevRemoteSys et.ElevState) bool {
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			// Only care about where orders may have been lost, so if there was no order in the old ElevState, skip
			if !(prevRemoteSys.CurrentOrders[f][b].GetID() == "") {
				// Check if IDs are different
				if !(remoteSys.CurrentOrders[f][b].GetID() == prevRemoteSys.CurrentOrders[f][b].GetID()) {
					// Since IDs are different, check if the order in the old ElevState has since been finished
					if !isOrderAlreadyFinished(remoteSys, prevRemoteSys.CurrentOrders[f][b].GetID()) {
						// If Prev order was accepted, we have removed an order that MUST be completed, which is a problem.
						if prevRemoteSys.CurrentOrders[f][b].IsAccepted() {
							return false
						}
					}
				}
			}

		}
	}
	return true
}

/*isOrderAlreadyFinished checks if sys has registered a finished order with Id orderID
 * @arg sys: State of some system
 * @arg orderID: Id of an order
 * @return: True if an order with the corresponding Id was found to be finished
 */
func isOrderAlreadyFinished(sys et.ElevState, orderID string) bool {
	if orderID == "" {
		return false
	}
	for _, order := range sys.FinishedOrders {
		if order.Id == orderID {
			return true
		}
	}
	return false
}

/*applyUpdatesToLocalSystem calls functions which updates the local system with updates from remote.
 * @arg remote: Another system's state
 */
func applyUpdatesToLocalSystem(remote et.ElevState) {

	mergeFinishedOrdersQueue(remote)
	mergeOrdersToLocalSystem(remote)
	addLocalAckToOrders()
	applyOrderAcksFromRemote(remote)

}

/*mergeOrdersToLocalSystem updates the local queue based on remote's queue
 * @arg remote: Another system's state
 */
func mergeOrdersToLocalSystem(remote et.ElevState) {
	localSystem, _ := systems[LocalID]
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			o := updateSingleOrder(localSystem.CurrentOrders[f][b], remote.CurrentOrders[f][b])
			localSystem.CurrentOrders[f][b] = o
		}
	}
	systems[LocalID] = localSystem

}

/*updateSingleOrder determines which of the remote and locally stored order should be kept
 * This is determined by checking the order type, if they are empty, and which order was last modified
 * @arg localOrder: order that we have in this system's CurrentOrder queue
 * @arg remoteOrder: order that we have in another system's CurrentOrder queue
 * @return: Order which will be kept
 */
func updateSingleOrder(localOrder et.ElevOrder, remoteOrder et.ElevOrder) et.ElevOrder {
	var o et.ElevOrder
	localSystem := systems[LocalID]

	// Cab orders are handled locally, so always keep the local order.
	if localOrder.IsCabOrder() || remoteOrder.IsCabOrder() {
		return localOrder
	}

	// Orders finished elsewhere have been registered in localSystem as well.
	remoteFinished := isOrderAlreadyFinished(localSystem, remoteOrder.Id)
	localFinished := isOrderAlreadyFinished(localSystem, localOrder.Id)

	if localFinished && remoteFinished {
		o = et.EmptyOrder()
		return o
	} else if localFinished {
		o = remoteOrder
		return o
	} else if remoteFinished {
		o = localOrder
		return o
	}

	// If an order is empty, it's of no interest
	if localOrder.IsEmpty() {
		return remoteOrder
	} else if remoteOrder.IsEmpty() {
		return localOrder
	}

	// We have two orders in the same floor/button. Likely, it is the same order ID.
	if localOrder.GetID() != remoteOrder.GetID() {
		//@TODO remove and use the other one.
		o = updateSingleOrderNonMatchingIDs(localOrder, remoteOrder)
	} else {
		o = updateSingleOrderMatchingIDs(localOrder, remoteOrder)
	}

	return o
}

/*updateSingleOrderNonMatchingIDs contains logic for determining which order should be kept when both exist, but have nonmatching IDs
 * It chooses by looking at whether they are accepted, and when they were last modified.
 * @arg localOrder: order that we have in this system's CurrentOrder queue
 * @arg remoteOrder: order that we have in another system's CurrentOrder queue
 * @return: Order which will be kept
 */
func updateSingleOrderNonMatchingIDs(localOrder et.ElevOrder, remoteOrder et.ElevOrder) et.ElevOrder {
	var o et.ElevOrder
	log.WithFields(log.Fields{"ID1": localOrder.GetID(), "ID2": remoteOrder.GetID()}).Warn("sysstate updateSingleOrder: Non-matching IDs")
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
	return o

}

/*updateSingleOrderMatchingIDs contains logic for determining which order should be kept when both exist, and have matching IDs
 * It chooses by looking at whether they are accepted, and when they were last modified.
 * @arg localOrder: order that we have in this system's CurrentOrder queue
 * @arg remoteOrder: order that we have in another system's CurrentOrder queue
 * @return: Order which will be kept
 */
func updateSingleOrderMatchingIDs(localOrder et.ElevOrder, remoteOrder et.ElevOrder) et.ElevOrder {
	var o et.ElevOrder
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
	return o
}

/*mergeFinishedOrdersQueue updates our local system's list of finished orders with the ones from remoteSystem.
 * @arg remoteSystem: Another system's state
 */
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

/*addLocalAckToOrders registeres our ACK to orders in our queue where it has not yet been registered.
 * We only register ACK if the order's assignee is active (in our heartbeat system)
 */
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

					}
				}
			}

		}
	}
	systems[LocalID] = localSystem
}

/*applyOrderAcksFromRemote finds ACKs registered remotely on orders and registers them on our local orders (orders with same ID)
 * @arg remote: Another system's state
 */
func applyOrderAcksFromRemote(remote et.ElevState) {
	localSystem := systems[LocalID]
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if remote.CurrentOrders[f][b].Id == localSystem.CurrentOrders[f][b].Id {
				newAcks := getAcksOnlyRegisteredRemotely(localSystem.CurrentOrders[f][b], remote.CurrentOrders[f][b])
				for _, ack := range newAcks {
					localSystem.CurrentOrders[f][b].Acks = append(localSystem.CurrentOrders[f][b].Acks, ack)
				}
			}
		}
	}
	systems[LocalID] = localSystem

}

/*contains checks if container has an item equal to element
 * @return true if container contains an identical item to element
 */
func contains(container []int32, element int32) bool {
	for _, elem := range container {
		if elem == element {
			return true
		}
	}
	return false
}

/*getAcksOnlyRegisteredRemotely finds the ACKs registered only for remoteOrder, and returns them as a slice
 * @arg localOrder: locally registered order
 * @arg remoteOrder: order from update from another sys
 * @return: Slice of ACKs (system IDs)
 */
func getAcksOnlyRegisteredRemotely(localOrder et.ElevOrder, remoteOrder et.ElevOrder) []int32 {
	var acks []int32

	for _, remoteAck := range remoteOrder.Acks {
		ackAlreadyRegisteredLocally := false
		for _, localAck := range localOrder.Acks {
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

/*acceptOrdersWeCanGuarantee checks this system's queue for orders we can accept (redundancy is a prerequisite), and accepts them
 * After accepting each order, we backup, to ensure no loss of accepted orders occurs.
 */
func acceptOrdersWeCanGuarantee() {
	localSystem, _ := systems[LocalID]
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if localSystem.CurrentOrders[f][b].Status == et.Received &&
				canGuaranteeOrderCompletion(localSystem.CurrentOrders[f][b]) {
				log.WithField("o", localSystem.CurrentOrders[f][b]).Debug("sysstate acceptOrders: Can guarantee order; accepting")
				accept(&localSystem, localSystem.CurrentOrders[f][b])
				sb.Backup(GetSystemsStates())

			}
		}
	}
	systems[LocalID] = localSystem
}

/*rejectOrder removes an order from the local queue
 * It's used for rejecting received orders which have timed out etc.
 * @arg orderID: Order to reject
 */
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

/*canGuaranteeOrderCompletion returns true only if at least two systems know about and have acknowledged an order.
 * @arg o: Order which we want to check if can be guaranteed.
 */
func canGuaranteeOrderCompletion(o et.ElevOrder) bool {
	c := countOrderOccurrencesInSystems(o)
	if c >= 2 {
		return true
	}
	return false
}

/*countOccurrencesInSystems returns the number of systems which have registered and acknowledged an order.
 * @arg o: Order which we want to count occurrences of
 */
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
	s, _ := systems[LocalID]

	// Match orders in CurrentOrders to orderIDs finished by the local elevator.
	for _, o := range s.E.FinishedOrders {

		if !o.IsEmpty() &&
			(s.CurrentOrders[o.Order.Floor][int(o.Order.Button)].Id == o.Id ||
				s.CurrentOrders[o.Order.Floor][int(o.Order.Button)].Assignee == LocalID && s.CurrentOrders[o.Order.Floor][int(o.Order.Button)].IsAccepted()) {
			finish(&s, o.Order.Floor, int(o.Order.Button))
			putOrderInFinishedOrdersList(&s, o.Order.Floor, int(o.Order.Button))
		}
	}

	systems[LocalID] = s

}

/*puOrderInFinishedOrdersList moves an order from CurrentOrders to FinishedOrders.
 * It takes the place of the oldest order in FinishedOrders.
 * @arg s: Local system's state
 * @arg floor: floor of order
 * @arg button: button of order
 */
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

/*isOrderAlreadyActive checks if there already exists an active order at a specific floor/button
 * @arg btn: ButtonEvent corresponding to the floor/button of an order
 * @return: True if an active order exists, false otherwise
 */
func isOrderAlreadyActive(btn et.ButtonEvent) bool {
	system := systems[LocalID]
	if system.CurrentOrders[btn.Floor][int(btn.Button)].IsActive() {
		log.WithField("Order", system.CurrentOrders[btn.Floor][int(btn.Button)]).Debug("sysstate isAlreadyActive: This is the order which registers as the same")
		return true
	}
	return false
}

/*countTimeoutsOfCabOrders returns the sum of the # of timouts of current local cab orders
 */
func countTimeoutsOfCabOrders() int {
	localSys := systems[LocalID]
	count := 0
	for f := 0; f < et.NumFloors; f++ {
		count += localSys.CurrentOrders[f][int(et.BT_Cab)].CountTimesTimeout
	}
	return count
}

/*accept registers an order as accepted.
 */
func accept(localSys *et.ElevState, o et.ElevOrder) {
	(*localSys).CurrentOrders[o.GetFloor()][int(o.GetButton())].Status = et.Accepted
	(*localSys).CurrentOrders[o.GetFloor()][int(o.GetButton())].TimestampLastOrderStatusChange = time.Now().Unix()
}

/*empty empties an order.
 */
func empty(localSys *et.ElevState, o et.ElevOrder) {
	(*localSys).CurrentOrders[o.GetFloor()][int(o.GetButton())] = et.EmptyOrder()
}

/*redelegate redelegates an order to the specified elevator.
 */
func redelegate(localSys *et.ElevState, o et.ElevOrder, newAssignee int32) {
	(*localSys).CurrentOrders[o.GetFloor()][int(o.GetButton())].Assignee = newAssignee
	(*localSys).CurrentOrders[o.GetFloor()][int(o.GetButton())].TimestampLastOrderStatusChange = time.Now().Unix()
	(*localSys).CurrentOrders[o.GetFloor()][int(o.GetButton())].TimestampReceived = time.Now().Unix()
}

/*finish marks an order as finished.
 */
func finish(es *et.ElevState, floor int, button int) {
	es.CurrentOrders[floor][button].Status = et.Finished
	es.CurrentOrders[floor][button].TimestampLastOrderStatusChange = time.Now().Unix()
}
