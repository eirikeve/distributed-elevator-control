package elevorderevaluation

import (
	"errors"
	"strconv"
	"time"

	fsm "../elevfsm"
	loc "../elevnetwork/localip"
	et "../elevtype"

	log "github.com/sirupsen/logrus"
)

/*
 * Description
 */

////////////////////////////////
// Module varibles
////////////////////////////////
const TRAVEL_TIME = 3
const DOOR_OPEN_TIME = 5

////////////////////////////////
// Interface
////////////////////////////////

/*FindOptimalSystem (.) takes a list of systems and a new received order, simulates all
 * elevators with the new order and returns the ID of the elevator best fit, according to timeToIdle,
 * to execute the new order. Except for cab orders, they will always be delegated to the local elevator
 *
 * @arg systems: Currently active elevators
 * @arg newOrder: The new order which is to be delegated
 * @arg return: The ID corresponding to the optimal system
 */
func FindOptimalSystem(systems []et.ElevState, newOrder et.ButtonEvent) (int32, error) {
	var elevators []et.Elevator

	if et.IsCabButton(newOrder) {
		return loc.LocalID()
	}
	for _, sys := range systems {
		elevators = append(elevators, sys.E)
	}
	elevIndex, err := delegateOrder(elevators, newOrder)
	optSysID := systems[elevIndex].ID
	return optSysID, err
}

////////////////////////////////
// Auxiliary
////////////////////////////////

/*timeToIdle (.) estimates how much time the elevator will use to execute
* all its order.

* @arg elev: Takes an Elevator as arguemtent, making it possible to simalute its actions
* @return: Total duration for executing all orders
 */
func timeToIdle(elev et.Elevator) int {
	duration := 0
	isSimulating := true

	switch elev.State {
	case et.Idle:
		elev.MovementDirection = fsm.OrderLogicGetMovementDirection(elev)
		if elev.MovementDirection == et.MD_Stop {
			return duration
		}
	case et.Moving:
		duration += TRAVEL_TIME / 2
		elev.Floor += int(elev.MovementDirection)
	case et.Unloading:
		duration += DOOR_OPEN_TIME / 2
	default:
	}
	for isSimulating == true {
		if fsm.OrderLogicCheckShouldStopAtFloor(elev) == true {
			elev = fsm.OrderLogicClearRequestsOnCurrentFloor(elev, elev.MovementDirection)
			duration += DOOR_OPEN_TIME
			elev.MovementDirection = fsm.OrderLogicGetMovementDirection(elev)
			if elev.MovementDirection == et.MD_Stop {
				return duration
			}
		}
		elev.Floor += int(elev.MovementDirection)
		duration += TRAVEL_TIME
	}
	return duration
}

/*delegateOrder (.) Takes a list of elevators, and simulates them all finding
 * which is best fit to execute a hall order, according to timeToIdle.

 * @arg elevList: List of elevators
 * @arg newOrder: The new hall order which is to be assigned
 * @return: The index corresponding to the optimal elevator from elevList
 */
func delegateOrder(elevList []et.Elevator, newOrder et.ButtonEvent) (int, error) {
	var durations []int
	var err error = nil

	for _, elev := range elevList {
		if orderExsists(elev, newOrder) {
			err = errors.New("elevorderevalution: Order exists")
		}
		elev = insertElevatorOrder(elev, newOrder)
		tempDuration := timeToIdle(elev)
		durations = append(durations, tempDuration)
	}

	optElevIndex := findMinIndex(durations)

	return optElevIndex, err
}

/*findMinIndex (.) takes a list of integers and returns the index
 * containing the smallest value.
 * @arg list: list containing integers
 * @return: smallest value of the list
 */
func findMinIndex(list []int) int {

	var maxIndex int
	var maxValue int
	if list == nil {
		log.Error("FindMinIndex containing empty list")
	}
	for index, element := range list {
		if index == 0 {
			maxValue = element
			maxIndex = index
		}
		if element < maxValue {
			maxValue = element
			maxIndex = index
		}
	}
	return maxIndex
}

/*insertElevatorOrder (.) inserts a order in the elevator queue
 * at the given location.

* @arg elev: Elevator
* @arg ButtonEvent: Which button was triggered and at which floor is was pressed
* @return: The given elevator updated with the new order
*/
func insertElevatorOrder(elev et.Elevator, bEvent et.ButtonEvent) et.Elevator {
	//'TODO make actual order id :)
	order := et.SimpleOrder{Id: strconv.FormatInt(time.Now().Unix(), 16), Order: bEvent}
	elev.Orders[order.Order.Floor][order.Order.Button] = order
	return elev
}

/*orderExsists (.) checks if new order already exsists,
 * then it should not be necessary to redelegate order

 * @arg elev: An elevator contaning the current orders
 * @arg bEvent: The button which was pressed
 * @return: true if given order exsist in the elevator, else false
 */
func orderExsists(elev et.Elevator, bEvent et.ButtonEvent) bool {
	if elev.Orders[bEvent.Floor][bEvent.Button].Id != "" {
		return true
	} else {
		return false
	}
}
