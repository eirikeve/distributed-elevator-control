package elevfsm

import (
	et "../elevtype"
)

/*
 * Orderlogic contains functionality to determine basic elevator operations and inquiries
 */

////////////////////////////////
// Interface
////////////////////////////////

/*OrderLogicOrdersAbove (.) returns true if there exists a accepted order
 * above the elevators current floor, else false
 * @arg e: Elevator containing current active orders
 */
func OrderLogicOrdersAbove(e et.Elevator) bool {
	for f := e.Floor + 1; f < et.NumFloors; f++ {
		for btn := 0; btn < et.NumButtons; btn++ {
			if e.Orders[f][btn].IsActive() {
				return true
			}
		}
	}
	return false
}

/*OrderLogicOrdersBelow (.) returns true if there exists a accepted order
 * below the elevators current floor, else false
 * @arg e: Elevator containing current active orders
 */
func OrderLogicOrdersBelow(e et.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < et.NumButtons; btn++ {
			if e.Orders[f][btn].IsActive() {
				return true
			}
		}
	}
	return false
}

/*OrderLogciGetMoveMentDirection (.) returns a new motor direction based
 * on current elevator stats: Movement direction from last floor and orders.
 * @arg e: Elevator with information about current orders and previous movement direction from last floor
 */
func OrderLogicGetMovementDirection(e et.Elevator) et.MotorDirection {
	switch e.MovDirFromLastFloor {
	case et.MD_Up:
		if OrderLogicOrdersAbove(e) {
			return et.MD_Up
		} else if OrderLogicOrdersBelow(e) {
			return et.MD_Down
		}
		return et.MD_Stop
	case et.MD_Down:
		fallthrough
	case et.MD_Stop:
		if OrderLogicOrdersBelow(e) {
			return et.MD_Down
		} else if OrderLogicOrdersAbove(e) {
			return et.MD_Up
		}
		return et.MD_Stop
	default:
		return et.MD_Stop
	}
}

/*OrderLogicCheckShouldStopAtFloor (.) decides if the elevator should at
 * detected floor, based on current elevator stats: Movement direction and orders
 * @arg e: Elevator containing current movement direction and active orders.
 */
func OrderLogicCheckShouldStopAtFloor(e et.Elevator) bool {
	switch e.MovementDirection {
	case et.MD_Down:
		if e.Orders[e.Floor][et.BT_HallDown].IsActive() ||
			e.Orders[e.Floor][et.BT_Cab].IsActive() ||
			!OrderLogicOrdersBelow(e) {
			return true
		}
		return false
	case et.MD_Up:
		if e.Orders[e.Floor][et.BT_HallUp].IsActive() ||
			e.Orders[e.Floor][et.BT_Cab].IsActive() ||
			!OrderLogicOrdersAbove(e) {
			return true
		}
		return false
	case et.MD_Stop:
		fallthrough
	default:
		return true
	}

}

/*OrderLogicGetRequestsWeCanClearOnCurrentFloor (.) evaluates and returns a slice containing order IDs
 * of all orders that have been served on current floor.
 * @arg e: Elevator containg current active orders
 * @arg travelDirFromLastFloor: Movement direction from last floor, to decide which orders to serve
 * @returns: A list of orders that are now finished from current floor.
 */
func OrderLogicGetRequestsWeCanClearOnCurrentFloor(e et.Elevator, travelDirFromFloor et.MotorDirection) []et.SimpleOrder {
	var ordersWeFinish []et.SimpleOrder

	if e.Orders[e.Floor][et.BT_Cab].IsAccepted() {
		ordersWeFinish = append(ordersWeFinish, e.Orders[e.Floor][et.BT_Cab])
	}
	switch travelDirFromFloor {
	case et.MD_Up:
		if e.Orders[e.Floor][et.BT_HallUp].IsAccepted() {
			ordersWeFinish = append(ordersWeFinish, e.Orders[e.Floor][et.BT_HallUp])
		}

		if !OrderLogicOrdersAbove(e) {
			if e.Orders[e.Floor][et.BT_HallDown].IsAccepted() {
				ordersWeFinish = append(ordersWeFinish, e.Orders[e.Floor][et.BT_HallDown])
			}
		}
	case et.MD_Down:
		if e.Orders[e.Floor][et.BT_HallDown].IsAccepted() {
			ordersWeFinish = append(ordersWeFinish, e.Orders[e.Floor][et.BT_HallDown])
		}
		if !OrderLogicOrdersBelow(e) {
			if e.Orders[e.Floor][et.BT_HallUp].IsAccepted() {
				ordersWeFinish = append(ordersWeFinish, e.Orders[e.Floor][et.BT_HallUp])
			}
		}
	case et.MD_Stop:
		fallthrough
	default:
		if e.Orders[e.Floor][et.BT_HallUp].IsAccepted() {
			ordersWeFinish = append(ordersWeFinish, e.Orders[e.Floor][et.BT_HallUp])
		}
		if e.Orders[e.Floor][et.BT_HallDown].IsAccepted() {
			ordersWeFinish = append(ordersWeFinish, e.Orders[e.Floor][et.BT_HallDown])
		}
	}
	return ordersWeFinish
}

/*OrderLogicCheckIFRequestsAtCurrentFloor (.) returns true if there are any active
 * orders on current floor.
 */
func OrderLogicCheckIfRequestsAtCurrentFloor(elevator et.Elevator) bool {
	return elevator.Orders[elevator.Floor][et.BT_HallDown].IsActive() ||
		elevator.Orders[elevator.Floor][et.BT_Cab].IsActive() ||
		elevator.Orders[elevator.Floor][et.BT_HallUp].IsActive()
}

/*OrderLogicClearRequestsOnCurrentFloor (.) is used by elevordelegation to examines if the elevator should clear
 * a request on the current floor. Hence this function is ONLY used for simulation of the elevator.
 * @arg e: Contains the current orders in the Elevator
 * @arg travelDirFromFloor: Movement direction from last floor, to decide which orders to clear
 * @return: The elevator with updated queue
 */
func OrderLogicClearRequestsOnCurrentFloor(e et.Elevator, travelDirFromFloor et.MotorDirection) et.Elevator {
	// Only take passengers going in the directin the elevator will be moving
	e.Orders[e.Floor][et.BT_Cab] = et.SimpleOrder{}
	switch travelDirFromFloor {
	case et.MD_Up:
		e.Orders[e.Floor][et.BT_HallUp] = et.SimpleOrder{}

		if !OrderLogicOrdersAbove(e) {
			e.Orders[e.Floor][et.BT_HallDown] = et.SimpleOrder{}
		}

	case et.MD_Down:
		e.Orders[e.Floor][et.BT_HallDown] = et.SimpleOrder{}
		if !OrderLogicOrdersBelow(e) {
			e.Orders[e.Floor][et.BT_HallUp] = et.SimpleOrder{}
		}
	case et.MD_Stop:
		fallthrough
	default:
		e.Orders[e.Floor][et.BT_HallUp] = et.SimpleOrder{}
		e.Orders[e.Floor][et.BT_HallDown] = et.SimpleOrder{}
	}
	return e
}
