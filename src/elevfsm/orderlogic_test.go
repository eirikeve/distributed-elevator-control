package elevfsm

import (
	"fmt"
	"testing"
	"time"

	et "../elevtype"
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
			elev.Orders[floor][button] = nil
		}
	}

	return elev

}

func printElevatorQueue(elev et.Elevator) {
	println("\t\t BT_HallUp \t BT_HallDown \t BT_Cab")
	for floor := 0; floor < et.NumFloors; floor++ {
		fmt.Printf("Floor %v: \t  ", floor)
		for button := 0; button < et.NumButtons; button++ {
			if elev.Orders[floor][button] == nil {
				print("FALSE \t  ")
			} else {
				print("TRUE \t\t ")
			}
		}
		print("\n")
	}

}

func setElevatorOrder(elev et.Elevator, floor int, button et.ButtonType, ID string, stat et.OrderStatus) et.Elevator {
	bEvent := et.ButtonEvent{floor, button}
	eOrder := et.ElevOrder{ID, bEvent, 2, stat, 2, "Elev"}
	elev.Orders[floor][button] = &eOrder
	return elev
}

func TestOrderLogicsAboveandBelow(t *testing.T) {

	elev := initializeElevator(3)
	elev = setElevatorOrder(elev, 1, et.BT_Cab, "1", et.Accepted)
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
