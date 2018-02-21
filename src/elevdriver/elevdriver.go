package elevdriver

/*
elevdriver.go contains abstractions for methods from elevio.go
*/

import "../elevtype"
import "sync"
import log "github.com/sirupsen/logrus"

const elevatorAddress = "127.0.0.1"

/*
 * Driver initializes a connection to the elevator, and then pushes inputs and outputs to/from the elevator
 *
 * @arg directionInput: Input channel for setting the elev movement dir
 * @arg buttonLampInput: Input channel for setting button lamps on/off
 * @arg floorIndicatorInput: Input channel for setting the floor indicator
 * @arg doorOpenLampInput: Input channel for setting the door open light
 * @arg stopLampInput: Input channel for setting the stop lamp light
 * @arg buttonPressSensorOut: Output channel for button press readings
 * @arg floorSensorOut: Output channel for floor sensor
 * @arg stopButtonSensorOut: Output channel for the stop button press readings
 * @arg obstructionSensorOut: Output channel for the
 * @arg stopDriver: Input channel for signalling the driver to stop itself and its spawned GoRoutines
 * @arg numFloors: m floors. Max floor output for signals is set to numFloors.
 * @arg handlerWg: Handler's WaitGroup, for ensuring that Driver exits properly before Handler finishes shutdown
 */
func Driver(
	directionInput <-chan elevtype.MotorDirection,
	buttonLampInput <-chan elevtype.ButtonLamp,
	floorIndicatorInput <-chan int,
	doorOpenLampInput <-chan bool,
	stopLampInput <-chan bool,
	buttonPressSensorOut chan<- elevtype.ButtonEvent,
	floorSensorOut chan<- int,
	stopButtonSensorOut chan<- bool,
	obstructionSensorOut chan<- bool,
	stopDriver <-chan bool,
	numFloors int,
	handlerWg *sync.WaitGroup,
) {
	// ElevHandler WG for synchronization upon exit
	handlerWg.Add(1)
	defer handlerWg.Done()

	log.Debug("elevdriver Driver: Init connection")
	Init(elevatorAddress, numFloors)
	defer _conn.Close()

	wg := &sync.WaitGroup{}
	shutdown := make(chan bool, 10)
	runDriver := true

	// Outputs from driver to handler
	go PollButtons(buttonPressSensorOut, shutdown, wg)
	go PollFloorSensor(floorSensorOut, shutdown, wg)
	go PollStopButton(stopButtonSensorOut, shutdown, wg)
	go PollObstructionSwitch(obstructionSensorOut, shutdown, wg)
	log.Debug("elevdriver Driver: Started GoRoutines, running driver")

	for runDriver == true {
		select {
		// Inputs to driver from Handler
		case _ = <-stopDriver:
			log.Debug("elevdriver Driver: Recv stopDriver")
			runDriver = false
		case dir := <-directionInput:
			log.WithField("Dir", dir).Debug("elevdriver Driver: Setting motor dir")
			SetMotorDirection(dir)
		case btnLampInput := <-buttonLampInput:
			log.WithField("BtnLampInput", btnLampInput).Debug("elevdriver Driver: Setting Btn Lamp")
			SetButtonLamp(btnLampInput)
		case floor := <-floorIndicatorInput:
			log.WithField("FloorIndicator", floor).Debug("elevdriver Driver: Setting floor ind. light")
			SetFloorIndicator(floor)
		case doorOpenLampVal := <-doorOpenLampInput:
			log.WithField("DoorOpenLamp", doorOpenLampVal).Debug("elevdriver Driver: Setting door open lamp val")
			SetDoorOpenLamp(doorOpenLampVal)
		case stopLampVal := <-stopLampInput:
			log.WithField("StopLampVal", stopLampVal).Debug("elevdriver Driver: Setting stop lamp val")
			SetStopLamp(stopLampVal)

		}
	}
	// Stop GoRoutines
	log.Debug("elevdriver Driver: Shutdown GoRoutines")
	fill(shutdown, true)
	// WaitGroup synchronization
	wg.Wait()
	log.Debug("elevdriver Driver: Driver shutting down")
	return
}
