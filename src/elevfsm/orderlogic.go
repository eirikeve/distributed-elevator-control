package elevfsm

import (
	et "../elevtype"
)

func OrderLogicOrdersAbove(e et.Elevator) bool {
	// @todo handle if floor is -1
	for f := e.Floor + 1; f < et.NumFloors; f++ {
		for btn := 0; btn < et.NumButtons; btn++ {
			if e.Orders[f][btn].IsActive() {
				return true
			}
		}
	}
	return false
}
func OrderLogicOrdersBelow(e et.Elevator) bool {
	// @todo handle if floor is -1
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < et.NumButtons; btn++ {
			if e.Orders[f][btn].IsActive() {
				return true
			}
		}
	}
	return false
}
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
		// @todo
		// log error
		// if possible, try to move
	}
}
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
		// [@TODO] log - this should probably not happen.
		return true
	}
		
}
func OrderLogicClearRequestsOnCurrentFloor(e et.Elevator, travelDirFromFloor et.MotorDirection) et.Elevator {
	//@TODO add support for storing finished orders in some list
	// https://github.com/TTK4145/Project-resources/blob/master/elev_algo/requests.c
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

func OrderLogicCheckIfRequestsAtCurrentFloor(elevator et.Elevator) bool {
	return elevator.Orders[elevator.Floor][et.BT_HallDown].IsActive() ||
		elevator.Orders[elevator.Floor][et.BT_Cab].IsActive() ||
		elevator.Orders[elevator.Floor][et.BT_HallUp].IsActive()
}
