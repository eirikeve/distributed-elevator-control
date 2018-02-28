package elevfsm

import (
	et "../elevtype"
	"fmt"
	"testing"
)

func initializeElevator() et.Elevator {
	// Initialize a Elevator
	var elev et.Elevator
	// Initilize starting Floor
	elev.Floor = 1 //Is on second floor

	// Initilize Movement Directions
	elev.MovementDirection = et.MD_Stop // The elevator is not moving
	elev.MovementDirection = et.MD_Stop // Should not affect the elevators behaviour

	// Initilze current Elevator State
	elev.State = et.Idle

	// Initilize the Elevators Queue
	elev = initilizeElevatorQueue(elev)

	// Initilize Elevator ErrorState
	elev.ErrorState = et.FullFunctionality

	return elev
}

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
	println("Worked before")
	elev.Orders[floor][button] = &eOrder
	println("Worked after")
	return elev
}

func TestOrderLogicsAbove(t *testing.T) {

	elev := initializeElevator()
	printElevatorQueue(elev)
	elev = setElevatorOrder(elev, 2, et.BT_Cab, "1", et.Accepted)
	elev = setElevatorOrder(elev, 3, et.BT_Cab, "1", et.Accepted)
	printElevatorQueue(elev)
}
