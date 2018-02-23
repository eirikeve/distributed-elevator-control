package elevdriver

/*
elevdriver.go contains abstractions for methods from elevio.go
*/

import (
	"sync"
	"time"

	"../elevtype"

	log "github.com/sirupsen/logrus"
)

const elevatorAddress = "127.0.0.1:15657" // port taken from c driver
const stdNumFloorsElevator = 4
const timeWaitForDriverToStartMs = 20

var wg = &sync.WaitGroup{}
var stopDriverChan = make(chan bool)
var initialized = false

/*StartDriver creates a (singleton) driver instance running on a separate thread.
 * The driver uses channels to set Elevator parameters (Input chans),
 * and output elevator sensor information (Out chans).
 * Can be called directly - not necessary to call as a GoRoutine.
 *
 * Typedefs for custom types are found in the ../elevtype package.
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
 * @arg numFloorsElevators: m floors. Max floor output for signals is set to numFloorsElevator.
 */
func StartDriver(
	numFloorsElevator int,
	directionInput <-chan elevtype.MotorDirection,
	buttonLampInput <-chan elevtype.ButtonLamp,
	floorIndicatorInput <-chan int,
	doorOpenLampInput <-chan bool,
	//stopLampInput <-chan bool,
	buttonPressSensorOut chan<- elevtype.ButtonEvent,
	floorSensorOut chan<- int,
	//stopButtonSensorOut chan<- bool,
	//obstructionSensorOut chan<- bool,

) {
	log.Debug("elevdriver StartDriver: Driver starting")
	if initialized {
		log.Warning("elevdriver StartDriver: Driver already running. Returning")
		return
	}
	// Reinitialize variables to make sure wg is cleared and chan is empty
	// wg is for ensuring driver's goroutines are stopped before StopDriver exits
	wg := sync.WaitGroup{}
	stopDriverChan = make(chan bool)

	// Assert valid number of floors
	if numFloorsElevator < 1 {
		log.WithField("numFloorsElevator", numFloorsElevator).Error("elevdriver StartDriver: < 1 floors input! Defaulting to numFloorsElevator=4")
		numFloorsElevator = stdNumFloorsElevator
	}

	// Create driver instance
	go driver(
		numFloorsElevator,
		directionInput,
		buttonLampInput,
		floorIndicatorInput,
		doorOpenLampInput,
		//stopLampInput,
		buttonPressSensorOut,
		floorSensorOut,
		//stopButtonSensorOut,
		//obstructionSensorOut,
		stopDriverChan,
		&wg,
	)
	initialized = true
	time.Sleep(time.Millisecond * timeWaitForDriverToStartMs)

	log.Debug("elevdriver StartDriver: Driver started")
	return
}

/*StopDriver stops the running driver instance
 * Can be called directly - not necessary to call as a GoRoutine.
 */
func StopDriver() {
	log.Debug("elevdriver StopDriver: Driver stopping")
	if !initialized {
		log.Warning("elevdriver StopDriver: Driver already stopped. Returning")
		return
	}
	stopDriverChan <- true
	wg.Wait()
	initialized = false
	log.Debug("elevdriver StopDriver: Driver stopped")
}

/*driver (.) initializes a connection to the elevator, and then pushes inputs and outputs to/from the elevator
 * Start/Stop driver using StartDriver(.) and StopDriver().
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
 * @arg wg: WaitGroup used in StopDriver, for ensuring that Driver GoRoutines exit properly
 * 			before StopDriver returns
 */
func driver(
	numFloorsElevator int,
	directionInput <-chan elevtype.MotorDirection,
	buttonLampInput <-chan elevtype.ButtonLamp,
	floorIndicatorInput <-chan int,
	doorOpenLampInput <-chan bool,
	//stopLampInput <-chan bool,
	buttonPressSensorOut chan<- elevtype.ButtonEvent,
	floorSensorOut chan<- int,
	//stopButtonSensorOut chan<- bool,
	//obstructionSensorOut chan<- bool,
	stopDriver <-chan bool,
	wg *sync.WaitGroup,
) {
	// wg used in StopDriver() for synchronization upon exit
	wg.Add(1)
	defer wg.Done()

	log.Debug("elevdriver Driver: Init connection")
	initConnectionAndSetNumFloors(elevatorAddress, numFloorsElevator)
	defer _conn.Close() // _conn is a public variable in elevio.go

	shutdown := make(chan bool, 10)
	runDriver := true
	// Turn off lights here @todo

	// Outputs from driver to handler
	go PollButtons(buttonPressSensorOut, shutdown, wg)
	go PollFloorSensor(floorSensorOut, shutdown, wg)
	//go PollStopButton(stopButtonSensorOut, shutdown, wg)
	//go PollObstructionSwitch(obstructionSensorOut, shutdown, wg)
	log.Debug("elevdriver Driver: Started GoRoutines, running driver")

	driverDebugLogMsgTimer := time.Now()
	const driverDebugLogMsgFreq = time.Second

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
		//case stopLampVal := <-stopLampInput:
		//log.WithField("StopLampVal", stopLampVal).Debug("elevdriver Driver: Setting stop lamp val")
		//SetStopLamp(stopLampVal)
		default:
			if time.Now().Sub(driverDebugLogMsgTimer) > driverDebugLogMsgFreq {
				driverDebugLogMsgTimer = time.Now()
				log.Debug("elevdriver driver: Running")
			}

		}
	}
	// Stop GoRoutines before exiting
	log.Debug("elevdriver Driver: Shutdown GoRoutines, exiting")
	fill(shutdown, true)
	return
}
