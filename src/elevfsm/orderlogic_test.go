package elevfsm

import (
	"fmt"
	"testing"
	"time"

	et "../elevtype"
)

func InitializeElevator(startingFloor int) et.Elevator {
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
	elev = InitilizeElevatorQueue(elev)

	// Initilize Elevator ErrorState
	elev.ErrorState = et.FullFunctionality

	return elev
}

// Initilizes the Elevetors queue to empty
func InitilizeElevatorQueue(elev et.Elevator) et.Elevator {
	for floor := 0; floor < et.NumFloors; floor++ {
		for button := 0; button < et.NumButtons; button++ {
			elev.Orders[floor][button] = et.EmptyOrder()
		}
	}

	return elev

}

func PrintElevatorQueue(elev et.Elevator) {
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
func SetElevatorOrder(elev et.Elevator, floor int, button et.ButtonType, ID string, stat et.OrderStatus) et.Elevator {
	bEvent := et.ButtonEvent{floor, button}
	elev.Orders[floor][button] = et.ElevOrder{ID, bEvent, 2, stat, 2, "Elev"}
	return elev
}

/*
* Test the two functions: OrderLogicOrdersAbove & OrderLogicOrdersBelow
*/
func TestOrderLogicsAboveandBelow(t *testing.T) {

	elev := InitializeElevator(3)
	elev = SetElevatorOrder(elev, 3, et.BT_HallUp, "1", et.Accepted)
	elev = SetElevatorOrder(elev, 0, et.BT_HallDown, "2", et.Accepted)
	elev = SetElevatorOrder(elev, 0, et.BT_Cab, "3", et.Accepted)
	fmt.Printf("Starting order Function OrdersAbove \n\n\n")
	time.Sleep(time.Second * 1)
	orderAbove := OrderLogicOrdersAbove(elev)
	orderBelow := OrderLogicOrdersBelow(elev)
	fmt.Printf("Current Floor: %v \n", elev.Floor)
	fmt.Printf("OrderAbove: %v \t", orderAbove)
	fmt.Printf("OrderBelow: %v \n", orderBelow)
	print("\n\n")
	PrintElevatorQueue(elev)

}

/*
* Test the orderLogicGetMovementDirection
*/
func TestOrderLogicGetMovementDirection(t* testing.T){
	elev := InitializeElevator(3)
	elev = SetElevatorOrder(elev, 3, et.BT_HallUp, "1", et.Accepted)
	elev = SetElevatorOrder(elev, 0, et.BT_HallDown, "2", et.Accepted)
	elev = SetElevatorOrder(elev, 0, et.BT_Cab, "3", et.Accepted)
	
	PrintElevatorQueue(elev)
	movDirection := OrderLogicGetMovementDirection(elev)
	fmt.Printf("Current floor %v \n", elev.Floor)
	fmt.Printf("MovementDirection: %v \n",movDirection)
}
/*
 * Test all the functions in orderLogic
 * Description:
 * Init: Elevator is initilized in Floor 1, not moving with current Orders: Floor 1  BT_HallUP & Floor 1 BT_HallDown & Floor 3 BT_Cab
 * Expected bahaviour:
 *		- Elevator stops at currentFloor and removes request Floor 1 BT_HallUp, request Floor 1 BT_HallDown remains
 *		- Elevator starts moving upwards towards order Floor 3 BT_Cab 
 */
func TestOrderLogicCheckShouldStopAtFloor(t* testing.T){

	//Initialize Elevator Scenario
	elev := InitializeElevator(1)
	elev = SetElevatorOrder(elev, 1, et.BT_HallUp, "1", et.Accepted)
	elev = SetElevatorOrder(elev,3,et.BT_Cab,"2",et.Accepted)
	elev = SetElevatorOrder(elev,1,et.BT_HallDown,"3", et.Accepted)

	//Find Movement Direction
	elev.MovementDirection = OrderLogicGetMovementDirection(elev)

	PrintElevatorQueue(elev)

	//Decide if elevator should stop at current floor
	shouldStop := OrderLogicCheckShouldStopAtFloor(elev)
	//Pritn info
	fmt.Printf("\n\n\n")
	fmt.Printf("Current floor: %v \n", elev.Floor)
	fmt.Printf("Movement Direction %v \n", elev.MovementDirection)
	fmt.Printf("Should stop at current floor: %v \n", shouldStop)
	
	if shouldStop{
		elev = OrderLogicClearRequestsOnCurrentFloor(elev, elev.MovementDirection)
	}



	fmt.Printf("\n\n\n")
	PrintElevatorQueue(elev)

	shouldStop = OrderLogicCheckShouldStopAtFloor(elev)
	movDirection := OrderLogicGetMovementDirection(elev)
	//Pritn info
	fmt.Printf("\n\n\n")
	fmt.Printf("Current floor: %v \n", elev.Floor)
	fmt.Printf("Movement Direction %v \n", movDirection)
	fmt.Printf("Should stop at current floor: %v \n", shouldStop)
	
	
	
}