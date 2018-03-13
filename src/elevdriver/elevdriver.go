package elevdriver

/*
elevdriver.go contains abstractions for methods from elevio.go
*/

import (
	"sync"
	"time"

	et "../elevtype"

	log "github.com/sirupsen/logrus"
)

const elevatorAddress = "127.0.0.1:15657" // port taken from c driver at https://github.com/TTK4145/driver-c
const stdNumFloorsElevator = 4
const timeWaitForDriverToStartMs = 20

var wg = &sync.WaitGroup{}
var stopDriverChan = make(chan bool)
var driverRunning = false
var driverLock sync.Mutex

/*StartDriver creates a (singleton) driver instance running on a separate thread.
 * The driver uses channels to set Elevator parameters (Input chans),
 * and output elevator sensor information (Out chans).
 * Can be called directly - not necessary to call as a GoRoutine.
 *
 * Typedefs for custom types are found in the ../elevtype package.
 *
 * @arg numFloorsElevators: m floors. Max floor output for signals is set to numFloorsElevator.
 * @arg motorDirectionInput: Input channel for setting the elev movement dir
 * @arg buttonLampInput: Input channel for setting button lamps on/off
 * @arg floorIndicatorInput: Input channel for setting the floor indicator
 * @arg doorOpenLampInput: Input channel for setting the door open light
 * @arg buttonPressSensorOut: Output channel for button press readings
 * @arg floorSensorOut: Output channel for floor sensor
 */
func StartDriver(
	numFloorsElevator int,
	motorDirectionInput <-chan et.MotorDirection,
	buttonLampInput <-chan et.ButtonLamp,
	floorIndicatorInput <-chan int,
	doorOpenLampInput <-chan bool,
	buttonPressSensorOut chan<- et.ButtonEvent,
	floorSensorOut chan<- int,
) {
	driverLock.Lock()
	defer driverLock.Unlock()

	log.Info("elevdriver StartDriver: Driver starting")
	if driverRunning {
		log.Warning("elevdriver StartDriver: Driver already running. Returning")
		return
	}
	driverRunning = true
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
		motorDirectionInput,
		buttonLampInput,
		floorIndicatorInput,
		doorOpenLampInput,
		buttonPressSensorOut,
		floorSensorOut,
		stopDriverChan,
		&wg,
	)

	time.Sleep(time.Millisecond * timeWaitForDriverToStartMs)

	log.Debug("elevdriver StartDriver: Driver started")
	return
}

/*StopDriver stops the running driver instance
 * Can be called directly - not necessary to call as a GoRoutine.
 */
func StopDriver() {
	driverLock.Lock()
	defer driverLock.Unlock()

	log.Debug("elevdriver StopDriver: Driver stopping")
	if !driverRunning {
		log.Warning("elevdriver StopDriver: Driver already stopped. Returning")
		return
	}
	stopDriverChan <- true
	wg.Wait()
	driverRunning = false
	log.Info("elevdriver StopDriver: Driver stopped")
}

/*driver (.) initializes a connection to the elevator, and then pushes inputs and outputs to/from the elevator
 * Start/Stop driver using StartDriver(.) and StopDriver().
 *
 * @arg numFloorsElevators: m floors. Max floor output for signals is set to numFloorsElevator.
 * @arg motorDirectionInput: Input channel for setting the elev movement dir
 * @arg buttonLampInput: Input channel for setting button lamps on/off
 * @arg floorIndicatorInput: Input channel for setting the floor indicator
 * @arg doorOpenLampInput: Input channel for setting the door open light
 * @arg buttonPressSensorOut: Output channel for button press readings
 * @arg floorSensorOut: Output channel for floor sensor
 * @arg stopDriver: Chan used for signalling driver goroutine to stop
 * @arg wg: WaitGroup used in StopDriver, for ensuring that Driver GoRoutines exit properly
 * 			before StopDriver returns
 */
func driver(
	numFloorsElevator int,
	motorDirectionInput <-chan et.MotorDirection,
	buttonLampInput <-chan et.ButtonLamp,
	floorIndicatorInput <-chan int,
	doorOpenLampInput <-chan bool,
	buttonPressSensorOut chan<- et.ButtonEvent,
	floorSensorOut chan<- int,
	stopDriver <-chan bool,
	wg *sync.WaitGroup,
) {
	// wg used in StopDriver() for synchronization upon exit
	wg.Add(1)
	defer wg.Done()

	log.Debug("elevdriver Driver: Init connection")
	initConnectionAndSetNumFloors(elevatorAddress, numFloorsElevator)
	defer shutdownConnection()

	shutdown := make(chan bool, 10)
	// Signal to stop GoRoutines before exiting
	defer fill(shutdown, true)
	//runDriver := true
	// Turn off lights here @todo

	// Outputs from driver to handler
	go pollButtons(buttonPressSensorOut, shutdown, wg)
	go pollFloorSensor(floorSensorOut, shutdown, wg)
	log.Debug("elevdriver Driver: Started GoRoutines, running driver")

	driverDebugLogMsgTimer := time.Now()
	const driverDebugLogMsgFreq = 2 * time.Second

	for {
		select {
		// Inputs to driver from Handler
		case <-stopDriver:
			log.Info("elevdriver Driver: Stopping driver")
			return
		case dir := <-motorDirectionInput:
			//log.WithField("Dir", dir).Debug("elevdriver Driver: Setting motor dir")
			setMotorDirection(dir)
		case btnLampInput := <-buttonLampInput:
			//log.WithField("BtnLampInput", btnLampInput).Debug("elevdriver Driver: Setting Btn Lamp")
			setButtonLamp(btnLampInput)
		case floor := <-floorIndicatorInput:
			//log.WithField("FloorIndicator", floor).Debug("elevdriver Driver: Setting floor ind. light")
			setFloorIndicator(floor)
		case doorOpenLampVal := <-doorOpenLampInput:
			//log.WithField("DoorOpenLamp", doorOpenLampVal).Debug("elevdriver Driver: Setting door open lamp val")
			setDoorOpenLamp(doorOpenLampVal)
		}
		if time.Now().Sub(driverDebugLogMsgTimer) > driverDebugLogMsgFreq {
			driverDebugLogMsgTimer = time.Now()
			log.Debug("elevdriver driver: Running")
		}
	}
}

func StartStopButtonService(stopPressed chan<- bool) {
	go pollStopButton(stopPressed, stopBtnShutdownChan, &stopBtnWg)
}

func StopStopButtonService() {
	stopBtnShutdownChan <- true
	stopBtnWg.Wait()
}
