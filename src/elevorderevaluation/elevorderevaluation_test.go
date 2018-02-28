package elevorderevaluation

import(
	"testing"
	et "../elevtype"
	"fmt"
)




/*
* Test if finding maximum index, i.e. element with largest integer value, is working properly
*/
func TestMinIndex(t *testing.T){
	durations := []int{27,22,27}
	index := findMinIndex(durations)
	println("MinIndex:",index)
	println("Containg the value", durations[index])

}

func TestTimeToIdle(t *testing.T){
	// Initialize elevator at Floor 1
	elevOne:= initializeElevator(1)
	
	// Set Queue for elevator
	elevOne = setElevatorOrder(elevOne, 1, et.BT_HallUp, "1", et.Accepted)
	elevOne = setElevatorOrder(elevOne,3,et.BT_Cab,"2",et.Accepted)
	elevOne = setElevatorOrder(elevOne,1,et.BT_HallDown,"3", et.Accepted)

	printElevatorQueue(elevOne)

	// Get duration
	fmt.Printf("\n")
	duration:= timeToIdle(elevOne)
	fmt.Printf("Duration: %v \n\n\n",duration)
}

func TestOrderDelegation(t *testing.T){
	// Initialize elevator at Floor 1
	elevOne:= initializeElevator(1)
	elevTwo:= initializeElevator(1)
	elevThree:= initializeElevator(1)
	
	// Set Queue for elevator 1
	elevOne = setElevatorOrder(elevOne, 1, et.BT_HallUp, "1", et.Accepted)
	elevOne = setElevatorOrder(elevOne,3,et.BT_Cab,"2",et.Accepted)
	elevOne = setElevatorOrder(elevOne,1,et.BT_HallDown,"3", et.Accepted)

	// Set Queue for elevator 1
	elevTwo = setElevatorOrder(elevTwo, 1, et.BT_HallUp, "1", et.Accepted)
	//elevTwo = setElevatorOrder(elevTwo,3,et.BT_Cab,"2",et.Accepted)
	elevTwo = setElevatorOrder(elevTwo,1,et.BT_HallDown,"3", et.Accepted)

	// Set Queue for elevator 1
	elevThree = setElevatorOrder(elevThree, 1, et.BT_HallUp, "1", et.Accepted)
	elevThree = setElevatorOrder(elevThree,3,et.BT_Cab,"2",et.Accepted)
	elevThree = setElevatorOrder(elevThree,1,et.BT_HallDown,"3", et.Accepted)

	listElev := []et.Elevator{elevOne, elevTwo, elevThree}
	bestElev:= delegateOrder(listElev)

	fmt.Printf("Best elevator: %v \n", bestElev)


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
			elev.Orders[floor][button] = et.EmptyOrder()
		}
	}

	return elev

}

func printElevatorQueue(elev et.Elevator) {
	println("\t\t BT_HallUp \t BT_HallDown \t BT_Cab")
	for floor := 0; floor < et.NumFloors; floor++ {
		fmt.Printf("Floor %v: \t  ", floor)
		for button := 0; button < et.NumButtons; button++ {
			if elev.Orders[floor][button].Status == et.Accepted {
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
* @arg floor: Set which floor the order is from
* @arg button: Set which Buttontype that was pressed
* @arg ID: Give unique ID for the Order
* @arg stat: Set status for Order
*/
func setElevatorOrder(elev et.Elevator, floor int, button et.ButtonType, ID string, stat et.OrderStatus) et.Elevator {
	bEvent := et.ButtonEvent{floor, button}
	elev.Orders[floor][button] = et.ElevOrder{ID, bEvent, 2, stat, 2, "Elev"}
	return elev
}

