package elevfsm


func ordersAbove(e elevtype.Elevator) bool {
	// @todo handle if floor is -1
	for (f := e.Floor; f < elevtype.NumFloors; f++ ) {
		for (btn := 0; btn < elevtype.NumButtons; ++btn) {
			if(e.Orders[f][btn] != nil) {
				return true
			}
		}
	}
	return false
}
func ordersBelow(e elevtype.Elevator) bool {
	// @todo handle if floor is -1
	for (f := 0; f < e.Floor; f++ ) {
		for (btn := 0; btn < elevtype.NumButtons; ++btn) {
			if(e.Orders[f][btn] != nil) {
				return true
			}
		}
	}
	return false
}
func getMovementDirection(e elevtype.Elevator) elevtype.MotorDirection {
	switch (e.Direction) {
	case elevtype.MD_Up:
		if ordersAbove(e) {
			return MD_Up
		} else if ordersBelow(e){
			return MD_Down
		}
		return MD_Stop
	case elevtype.MD_Down:
	case elevtype.MD_Stop:
		if ordersBelow(e) {
			return MD_Down
		} else if ordersAbove(e){
			return MD_Up
		}
		return MD_Stop
	default:
		return MD_Stop
		// @todo
		// log error
		// if possible, try to move
	}
}
func checkShouldStopAtFloor(e elevtype.Elevator) bool {
	switch (e.MovementDirection) {
	case MD_Down:
		if (e.Orders[e.Floor][elevtype.BT_HallDown] ||
			e.Orders[e.Floor][elevtype.BT_Cab]		||
			!ordersBelow(e) ) {
				return true
		}
		return false
	case MD_Up:
		if (e.Orders[e.Floor][elevtype.BT_HallUp] ||
			e.Orders[e.Floor][elevtype.BT_Cab]	||
			!ordersAbove(e) ) {
				return true
		}
		return false
	case MD_Stop:
	default:
		// @todo log - this should probably not happen.
		return true
	}
}
func clearRequestsOnCurrentFloor(e elevtype.Elevator) {
	// https://github.com/TTK4145/Project-resources/blob/master/elev_algo/requests.c

}
