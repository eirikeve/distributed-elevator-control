package elevfsm

import (
	et "../elevtype"
	"fmt"
	"testing"
	"time"
)

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

func setElevatorOrder(elev et.Elevator, floor int, button et.ButtonType, ID string, stat et.OrderStatus) et.Elevator {
	bEvent := et.ButtonEvent{floor, button}
	elev.Orders[floor][button] = et.ElevOrder{ID, bEvent, 2, stat, 2, "Elev"}
	return elev
}

func TestOrderLogicsAboveandBelow(t *testing.T) {

	elev := initializeElevator(1)
	elev = setElevatorOrder(elev, 3, et.BT_HallUp, "1", et.Accepted)
	elev = setElevatorOrder(elev, 0, et.BT_HallDown, "1", et.Accepted)
	fmt.Printf("Starting order Function OrdersAbove \n\n\n")
	time.Sleep(time.Second * 1)
	orderAbove := OrderLogicOrdersAbove(elev)
	orderBelow := OrderLogicOrdersBelow(elev)
	fmt.Printf("Current Floor: %v \n", elev.Floor)
	fmt.Printf("OrderAbove: %v \t", orderAbove)
	fmt.Printf("OrderBelow: %v \n", orderBelow)
	print("\n\n")
	printElevatorQueue(elev)

}

func TestOrderLogicGetMovementDirection(t* testing.T){
	elev := initializeElevator(1)
	//elev = setElevatorOrder(elev, 1, et.BT_HallUp, "1", et.Accepted)
	//elev = setElevatorOrder(elev,3,et.BT_Cab,"1",et.Accepted)
	
	printElevatorQueue(elev)
	movDirection := OrderLogicGetMovementDirection(elev)
	fmt.Printf("Current floor %v \n", elev.Floor)
	fmt.Printf("MovementDirection: %v \n",movDirection)
}

func TestOrderLogicCheckShouldStopAtFloor(t* testing.T){

	//Initialize Elevator Scenario
	elev := initializeElevator(1)
	elev = setElevatorOrder(elev, 1, et.BT_HallUp, "1", et.Accepted)
	//elev = setElevatorOrder(elev,3,et.BT_Cab,"1",et.Accepted)

	//Find Movement Direction
	elev.MovementDirection = OrderLogicGetMovementDirection(elev)

	printElevatorQueue(elev)

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
	printElevatorQueue(elev)

	shouldStop = OrderLogicCheckShouldStopAtFloor(elev)
	movDirection := OrderLogicGetMovementDirection(elev)
	//Pritn info
	fmt.Printf("\n\n\n")
	fmt.Printf("Current floor: %v \n", elev.Floor)
	fmt.Printf("Movement Direction %v \n", movDirection)
	fmt.Printf("Should stop at current floor: %v \n", shouldStop)
	
	
	
}