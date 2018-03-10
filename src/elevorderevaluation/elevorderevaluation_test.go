package elevorderevaluation

import (
	"fmt"
	"testing"

	et "../elevtype"
)

/*
* Test if finding maximum index, i.e. element with largest integer value, is working properly
 */
func TestMinIndex(t *testing.T) {
	durations := []int{27, 8, 5}
	index := findMinIndex(durations)
	println("MinIndex:", index)
	println("Containg the value", durations[index])
}

func TestTimeToIdle(t *testing.T) {
	// Initialize elevator at Floor 1
	elevOne := initializeElevator(1)
	elevOne.MovementDirection = et.MD_Up
	elevOne.State = et.Unloading
	// Set Queue for elevator
	elevOne = setElevatorOrder(elevOne, 1, et.BT_HallUp, "1")
	elevOne = setElevatorOrder(elevOne, 0, et.BT_Cab, "2")
	//elevOne = setElevatorOrder(elevOne, 1, et.BT_HallDown, "3")

	printElevatorQueue(elevOne)

	// Get duration
	fmt.Printf("\n")
	duration := timeToIdle(elevOne)
	fmt.Printf("TimeToIdle, Duration: %v \n\n\n", duration)
}

func TestOrderDelegation(t *testing.T) {
	// Initialize elevator at Floor 1
	elevOne := initializeElevator(2)
	elevTwo := initializeElevator(0)
	elevThree := initializeElevator(1)

	// Set-up for elevator 1
	elevOne.MovementDirection = et.MD_Up
	elevOne.State = et.Moving
	elevOne = setElevatorOrder(elevOne, 1, et.BT_HallUp, "1")
	elevOne = setElevatorOrder(elevOne, 0, et.BT_Cab, "2")
	elevOne = setElevatorOrder(elevOne, 2, et.BT_HallDown, "3")

	// Set-up for elevator 2
	elevTwo = setElevatorOrder(elevTwo, 1, et.BT_HallUp, "1")
	elevTwo = setElevatorOrder(elevTwo, 3, et.BT_Cab, "2")
	elevTwo = setElevatorOrder(elevTwo, 1, et.BT_HallDown, "3")

	// Set-up for elevator 3
	elevThree.MovementDirection = et.MD_Up
	elevThree.State = et.Unloading
	elevThree = setElevatorOrder(elevThree, 1, et.BT_HallUp, "1")
	elevThree = setElevatorOrder(elevThree, 0, et.BT_Cab, "3")

	//New Order
	bEvent := et.ButtonEvent{1, et.BT_HallUp}

	listElev := []et.Elevator{elevOne, elevTwo, elevThree}
	bestElev := DelegateOrder(listElev, bEvent)

	fmt.Printf("TestOrderDelegation, Best elevator: %v \n", bestElev)

}

//[@todo]: Most of these functions are used in "elevfsm/orderlogic_test", maybe they should be imported?

func initializeElevator(startingFloor int) et.Elevator {
	// Initialize a Elevator
	var elev et.Elevator
	// Initilize starting Floor
	elev.Floor = startingFloor //Is on second floor

	// Initilize Movement Directions
	elev.MovementDirection = et.MD_Stop // The elevator is not moving
	elev.MovementDirection = et.MD_Stop // Should not affect the elevators behaviour

	// Initilze current Elevator State
	elev.State = et.Idle

	// Initilize the Elevators Queue to empty
	elev = initilizeElevatorQueue(elev)

	// Initilize Elevator ErrorState
	elev.ErrorState = et.FullFunctionality

	return elev
}

// Initilizes the Elevetors queue to empty
func initilizeElevatorQueue(elev et.Elevator) et.Elevator {
	for floor := 0; floor < et.NumFloors; floor++ {
		for button := 0; button < et.NumButtons; button++ {
			elev.Orders[floor][button] = et.SimpleOrder{}
		}
	}

	return elev

}

/*
* Sets a order in the Elevator queue at the given location
* @arg elev: Elevator
* @arg floor: Set which floor the order is from
* @arg button: Set which Buttontype that was pressed
* @arg ID: Give unique ID for the Order
* @arg stat: Set status for Order
 */
func setElevatorOrder(elev et.Elevator, floor int, button et.ButtonType, ID string) et.Elevator {
	bEvent := et.ButtonEvent{floor, button}
	elev.Orders[floor][button] = et.SimpleOrder{ID, bEvent}
	return elev
}
