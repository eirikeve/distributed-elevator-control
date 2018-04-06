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
	exsistsInSystems := false
	for index, element := range systems {
		if element.ID == LocalIP {
			exsistsInSystems = true
			systems[index].E = *e
			updateOrderStatesUponElevatorUpdate()
			break
		}

	}
	if !exsistsInSystems {
		log.Error("sysstate Update: Failed to update local elevator - has sysstate been initialized?")
	}
}

/*PushButtonEvent creates an order to handle the received btn event, but only if it is not yet an active order.
 *
 */
func PushButtonEvent(sysIndex int, btn et.ButtonEvent) {
	log.Debug("sysstate Push: Pushing btn")
	if !isOrderAlreadyAccepted(btn) {
		log.Debug("sysstate Push: Is new order!")

		t := time.Now().Unix()
		o := et.ElevOrder{
			Id:                strconv.FormatInt(time.Now().Unix(), 16),
			Order:             btn,
			TimestampReceived: t,
			Status:            et.Accepted, //@TODO change to received instead.
			TimestampLastOrderStatusChange: t,
			Assignee:                       systems[sysIndex].ID,
		}
		// Note that sysIndex says which system should perform the order.
		// But, since THIS system is the one delegating,
		// The order is stored in THIS system|s ElevState!
		// Other system's elevstate is only used for updating which orders
		// we will perform locally, we don't modify them
		exsistsInSystems := false
		for index, element := range systems {
			if element.ID == LocalIP {
				exsistsInSystems = true
				systems[index].CurrentOrders[btn.Floor][int(btn.Button)] = o
				log.Debug("sysstate Update: Set order")
			}
		}
		if !exsistsInSystems {
			log.Error("sysstate Push: Failed to push order - has sysstate been initialized?")
		}
	}
	// @TODO check if only one sys is active, if so, do NOT accept order!!!
}

/*Will contain logic to do stuff with orders that have been finished etc.
 */
func updateOrderStatesUponElevatorUpdate() {
	for index, element := range systems {
		if element.ID == LocalIP {
			// Check which accepted orders were finished
			updateFinishedOrders(systems[index])
		}
	}
}

func updateFinishedOrders(es et.ElevState) {
	e := es.E
	for f := 0; f < et.NumFloors; f++ {
		for b := 0; b < et.NumButtons; b++ {
			if isAcceptedOrder(es.CurrentOrders[f][b]) &&
				e.Orders[f][b].IsEmpty() {
				es.CurrentOrders[f][b].Status = et.Finished
			}
		}
	}

}

func isLocalOrder(o et.ElevOrder) bool {
	return o.Assignee == LocalIP
}

func isActiveOrder(o et.ElevOrder) bool {
	return o.Status == et.Accepted || o.Status == et.Received
}

func isAcceptedOrder(o et.ElevOrder) bool {
	return o.Status == et.Accepted
}

func isOrderAlreadyAccepted(btn et.ButtonEvent) bool {
	for _, element := range systems {
		if isActiveOrder(element.CurrentOrders[btn.Floor][int(btn.Button)]) {
			log.WithField("Order", element.CurrentOrders[btn.Floor][int(btn.Button)]).Debug("sysstate isAlreadyAcc: This is the order which registers as the same")
			return true
		}
	}
	return false
}
