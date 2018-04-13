package elevorderevaluation

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	def "../elevdef"
	fsm "../elevfsm"
	loc "../elevnetwork/localip"
	et "../elevtype"
)

const TRAVEL_TIME = 3
const DOOR_OPEN_TIME = 5

/*
* Calculates how much time the Elevator will use to execute
* all its order, thus going into the Idle state
* @arg elev: Takes an Elevator as arguemtent, making it possible to simalute its actions
 */
func timeToIdle(elev et.Elevator) int {
	duration := 0
	isSimulating := true

	switch elev.State {
	case et.Idle:
		elev.MovementDirection = fsm.OrderLogicGetMovementDirection(elev)
		if elev.MovementDirection == et.MD_Stop {
			println("\n\n et.MD_Stop \n\n")
			return duration
		}
	case et.Moving:
		duration += TRAVEL_TIME / 2
		elev.Floor += int(elev.MovementDirection) //[BUG] Not sure if converts motor type to int to floor
	case et.Unloading:
		duration += DOOR_OPEN_TIME / 2
	default:
		//Should not be possible to enter default
		println("Entered defualt in timeToIdle. This should not happen")
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
		duration += TRAVEL_TIME //[POTENTIAL BUG] Not sure if converts motor type to int to floor
	}

	return duration
}

/*
* Takes a list of Elevators, and simulates them all finding which elevator
* is best fit to take and execute an order.
* @arg elev[]: List of Elevators
 */
func delegateOrder(elevList []et.Elevator, newOrder et.ButtonEvent) (int, error) {
	var durations []int
	var err error = nil

	for _, elev := range elevList {
		if orderExsists(elev, newOrder) {
			err = errors.New("elevorderevalution: Order exists")
		}
		elev = insertElevatorOrder(elev, newOrder)
		printElevatorQueue(elev)
		tempDuration := timeToIdle(elev)
		durations = append(durations, tempDuration)
	}
	for index, element := range durations {
		fmt.Printf("delegateOrder, Elevator  %v: Duration: %v \n", index, element)
	}

	optElevIndex := findMinIndex(durations)

	return optElevIndex, err
}

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

/*
* Takes a list of integers and returns the index
* containing the largest value
* @arg list: list containing integers
 */
func findMinIndex(list []int) int {

	var maxIndex int
	var maxValue int
	if list == nil {
		println("Empty list, Crashing")
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

/*
 * Prints the Elevator Queue
 * @arg: An Elevator containg the Elevator Queue
 */
func printElevatorQueue(elev et.Elevator) {
	println("\t\t BT_HallUp \t BT_HallDown \t BT_Cab")
	for floor := 0; floor < def.NumFloors; floor++ {
		fmt.Printf("Floor %v: \t  ", floor)
		for button := 0; button < def.NumButtons; button++ {
			if elev.Orders[floor][button].IsActive() {
				print("TRUE \t\t  ")
			} else {
				print("FALSE \t  ")
			}
		}
		print("\n")
	}

}

/*
* Sets a order in the Elevator queue at the given location
* @arg elev: Elevator
* @arg ElevOrder: The new Elevator Order
* @arg ButtonEvent: Which button was triggered and at which floor is was pressed
 */
func insertElevatorOrder(elev et.Elevator, bEvent et.ButtonEvent) et.Elevator {
	//'TODO make actual order id :)
	isLocal := true
	removeTag := false
	order := et.SimpleOrder{Id: strconv.FormatInt(time.Now().Unix(), 16), Order: bEvent, IsLocalOrder: isLocal, TagRemoveOrder: removeTag}
	elev.Orders[order.Order.Floor][order.Order.Button] = order
	return elev
}

/*
 * Check if new order already exsists, then it should not be necessary
 * to redelegate order
 * @arg elev: An elevator contaning the current orders
 * @arg bEvent: The button which was pressed
 */
func orderExsists(elev et.Elevator, bEvent et.ButtonEvent) bool {
	if elev.Orders[bEvent.Floor][bEvent.Button].Id != "" {
		return true
	} else {
		return false
	}
}
