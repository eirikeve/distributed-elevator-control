package sysstate

import (
	"strconv"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

/*UpdateLocalElevator contains logic for updating the state & orders of the local system
 */
func UpdateLocalElevator(e *et.Elevator) {
	//log.Debug("sysstate Update: Update local elevator")

	if !initialized {
		initSysState()
	}

	system, _ := systems[LocalIP]
	system.E = *e
	systems[LocalIP] = system

	updateFinishedOrders()
}

/*PushButtonEvent creates an order to handle the received btn event, but only if it is not yet an active order.
 *
 */
func PushButtonEvent(sysID string, btn et.ButtonEvent) {

	if !initialized {
		initSysState()
	}

	log.Debug("sysstate Push: Pushing btn")
	if !isOrderAlreadyAccepted(btn) {
		log.Debug("sysstate Push: Is new order!")

		t := time.Now().Unix()
		o := et.ElevOrder{
			Id:                LocalIP + strconv.FormatInt(int64(btn.Floor), 10) + "-" + strconv.FormatInt(int64(btn.Button), 10) + "-" + strconv.FormatInt(time.Now().Unix(), 16),
			Order:             btn,
			TimestampReceived: t,
			Status:            et.Received, //@TODO change to received instead.
			TimestampLastOrderStatusChange: t,
			Assignee:                       sysID,
		}
		// Note that sysIndex says which system should perform the order.
		// But, since THIS system is the one delegating,
		// The order is stored in THIS system|s ElevState!
		// Other system's elevstate is only used for updating which orders
		// we will perform locally, we don't modify them

		// Verify that the system the order is assigned to exists
		_, ok := systems[sysID]
		// The order is saved in the currentOrders of the local elevator
		system, _ := systems[LocalIP]
		if ok {
			system.CurrentOrders[btn.Floor][int(btn.Button)] = o
			systems[LocalIP] = system
			log.WithField("o", o).Debug("sysstate Update: Set received order")
		} else {
			log.WithField("sysID", sysID).Error("sysstate Push: Order assigned to unknown system")
		}
	}
	// @TODO check if only one sys is active, if so, do NOT accept order!!!
}

/*updateFinishedOrders checks which orders have been finished by the local elevator
 * these orders are marked as status Finished and moved to the FinishedOrderList of the local system.
 */
func updateFinishedOrders() {
	if !initialized {
		initSysState()
	}
	s, _ := systems[LocalIP]

	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if s.CurrentOrders[f][b].Assignee == LocalIP && // Check that this elevator is supposed to carry out the order
				s.CurrentOrders[f][b].SentToAssigneeElevator && // Check that the order has been sent to the elevator FSM
				s.CurrentOrders[f][b].IsAccepted() && // Check if the order has been accepted
				s.E.Orders[f][b].IsEmpty() { // Check that the elevator FSM has carried out the order
				markOrderFinished(s, f, b)
				putOrderInFinishedOrdersList(s, f, b)
			}
		}
	}
	systems[LocalIP] = s

}

func markOrderFinished(es et.ElevState, floor int, button int) {
	es.CurrentOrders[floor][button].Status = et.Finished
	es.CurrentOrders[floor][button].TimestampLastOrderStatusChange = time.Now().Unix()
}
func putOrderInFinishedOrdersList(s et.ElevState, floor int, button int) {
	s.FinishedOrders = append(s.FinishedOrders, s.CurrentOrders[floor][button])
	s.CurrentOrders[floor][button] = et.EmptyOrder()
}

func isLocalOrder(o et.ElevOrder) bool {
	return o.Assignee == LocalIP
}

// Confusing name.... checks if an order already exists
func isOrderAlreadyAccepted(btn et.ButtonEvent) bool {
	for _, system := range systems {
		if system.CurrentOrders[btn.Floor][int(btn.Button)].IsActive() {
			log.WithField("Order", system.CurrentOrders[btn.Floor][int(btn.Button)]).Debug("sysstate isAlreadyAcc: This is the order which registers as the same")
			return true
		}
	}
	return false
}
