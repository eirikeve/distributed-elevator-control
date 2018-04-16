package main

import (
	"time"

	eh "./elevhandler"
	et "./elevtype"
	nh "./nethandler"
	setup "./setup"
)

func main() {
	// Parse command line arguments, initialize logging/systems/etc., recover backup from file (if found)
	setup.Setup()
	// If a backup of the local system was recovered, we'll run the system from our previous position/state
	elevator := setup.CheckBackupForExistingElevator()
	run(elevator)
}

func run(elevatorStartValues *et.Elevator) {

	// Channels for communication between elevhandler and nethandler
	orderQueueFromNethandler := make(chan [et.NumFloors][et.NumButtons]et.SimpleOrder, 12)
	buttonLightsFromNethandler := make(chan et.ButtonLamp, 12)
	buttonPressesToNethandler := make(chan et.ButtonEvent, 12)
	elevStateToNethandler := make(chan et.Elevator, 12)

	// Spawns the elevhandler goroutine, which controls the local elevator
	eh.StartElevHandler(orderQueueFromNethandler,
		buttonLightsFromNethandler,
		buttonPressesToNethandler,
		elevStateToNethandler,
		elevatorStartValues)

	// Spawns the nethandler goroutine, which controls communication and order delegation
	nh.StartNetHandler(orderQueueFromNethandler,
		buttonLightsFromNethandler,
		buttonPressesToNethandler,
		elevStateToNethandler)

	// Block the program from shutting down the goroutines
	running := true
	for running == true {
		time.Sleep(time.Second)
	}

}
