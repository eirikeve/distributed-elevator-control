package elevdriver

import (
	"sync"
	"time"

	"../elevlog"

	"testing"

	"../elevtype"
	log "github.com/sirupsen/logrus"
)

/*TestDriver tests basic log functionality, and shows how to use the log system
 * It runs a Light test (for lamps), and a Motor/Sensor test.
 * If elevator is connected and working,
 * lamps should flash and the cab should move up/down between floor 0 and 3
 * In addition, button presses are logged.
 */
func TestDriver(*testing.T) {
	elevlog.InitLog(log.DebugLevel, false)
	log.Info("elevdriver TestDriver: Starting test of elevdriver and elevio")

	// WaitGroup for syncing upon shutdown
	wg := sync.WaitGroup{}

	// Elevator control
	motorDirectionInput := make(chan elevtype.MotorDirection)
	buttonLampInput := make(chan elevtype.ButtonLamp, 12)
	floorIndicatorInput := make(chan int)
	doorOpenLampInput := make(chan bool)

	// Elevator sensors
	buttonPressSensorOut := make(chan elevtype.ButtonEvent)
	floorSensorOut := make(chan int)
	// Stop signal to Driver

	// Number of floors we're using (m=4)
	numFloorsElevator := 4

	log.Debug("elevdriver TestDriver: Initialized vars")

	StartDriver(
		numFloorsElevator,
		motorDirectionInput,
		buttonLampInput,
		floorIndicatorInput,
		doorOpenLampInput,
		buttonPressSensorOut,
		floorSensorOut,
	)
	log.Info("elevdriver TestDriver: Initialized Driver, running test")

	wg.Add(1)
	go RunLightTest(numFloorsElevator, buttonLampInput, floorIndicatorInput, doorOpenLampInput, &wg)

	wg.Add(1)
	go RunSensorAndMotorTest(numFloorsElevator, motorDirectionInput, buttonPressSensorOut, floorSensorOut, &wg)
	wg.Wait()
	log.Info("elevdriver TestDriver: Finished all tests")
	StopDriver()

}

/*RunLightTest sends light signals to the passed channel parameters.
 * It sets all lights on, and off, once.
 * Requires a running Driver for any effect.
 */
func RunLightTest(
	numFloors int,
	setButtonLampChan chan<- elevtype.ButtonLamp,
	setFloorIndicatorChan chan<- int,
	setDoorOpenLampChan chan<- bool,
	wg *sync.WaitGroup) {

	// Turn on inside lights
	var light = elevtype.ButtonLamp{
		Floor:  0,
		Button: elevtype.BT_Cab,
		Value:  true}

	// Logger for logging light variable
	ctxLogger := log.WithFields(log.Fields{"Light": &light})
	ctxLogger.Info("elevdriver RunLightTest: Starting Light Tests")

	// Turn on inside lights
	for i := 0; i < numFloors; i++ {
		setButtonLampChan <- light
		light.Floor = i
		ctxLogger.Debug("elevdriver RunLightTest: Setting Inside Light")
		setButtonLampChan <- light
	}

	time.Sleep(time.Second * 2)

	// Turn off inside lights
	light.Value = false
	for i := 0; i < numFloors; i++ {
		light.Floor = i
		ctxLogger.Debug("elevdriver RunLightTest: Setting Inside Light")
		setButtonLampChan <- light
	}

	time.Sleep(time.Second * 2)

	// Turn on outside lights
	light.Button = elevtype.BT_HallUp
	light.Value = true

	for i := 0; i < numFloors; i++ {
		light.Floor = i
		ctxLogger.Debug("elevdriver RunLightTest: Setting Hallup Light")
		setButtonLampChan <- light
	}

	time.Sleep(time.Second * 2)

	// Turn off outside lights
	light.Button = elevtype.BT_HallUp
	light.Value = false

	for i := 0; i < numFloors; i++ {
		ctxLogger.Debug("elevdriver RunLightTest: Setting Hallup Light")
		setButtonLampChan <- light
		light.Floor = i
	}

	time.Sleep(time.Second * 2)

	// Turn on HallDown lights
	light.Button = elevtype.BT_HallDown
	light.Value = true

	for i := 0; i < numFloors; i++ {
		light.Floor = i
		ctxLogger.Debug("elevdriver RunLightTest: Setting Halldown Light")
		setButtonLampChan <- light
	}

	time.Sleep(time.Second * 2)

	// Turn off HallDown lights
	light.Button = elevtype.BT_HallDown
	light.Value = false

	for i := 0; i < numFloors; i++ {
		ctxLogger.Debug("elevdriver RunLightTest: Setting Halldown Light")
		setButtonLampChan <- light
		light.Floor = i
	}

	time.Sleep(time.Second * 2)

	for i := 0; i < numFloors; i++ {
		log.WithField("Floor", i).Debug("elevdriver RunLightTest: Setting floor indicator")
		setFloorIndicatorChan <- i
		time.Sleep(time.Second)
	}

	log.Debug("elevdriver RunLightTest: Testing door open light")
	setDoorOpenLampChan <- true
	time.Sleep(time.Second * 2)
	setDoorOpenLampChan <- false

	log.Info("elevdriver RunLightTest: Done")

	wg.Done()

}

func RunSensorAndMotorTest(
	numFloors int,
	setMotorDirChan chan<- elevtype.MotorDirection,
	getButtonSensorChan <-chan elevtype.ButtonEvent,
	getFloorSensorChan <-chan int,
	wg *sync.WaitGroup) {

	log.Info("elevdriver RunSensorAndMotorTest: Running sensor and motor test")

	currentFloor := 1
	select {
	case currentFloor = <-getFloorSensorChan:
		if currentFloor == 0 {
			log.WithField("StartFloor", currentFloor).Info("elevdriver TestDriver: Startup motor dir up")
			setMotorDirChan <- elevtype.MD_Up
		} else {
			log.WithField("StartFloor", currentFloor).Info("elevdriver TestDriver: Startup motor dir down")
			setMotorDirChan <- elevtype.MD_Down
		}
	default:
		log.WithField("StartFloor", -1).Info("elevdriver TestDriver: Startup default to motor dir down")
		setMotorDirChan <- elevtype.MD_Down
	}

	timestamp := time.Now()
	smalltimecounter := time.Now()

	for time.Now().Sub(timestamp) < time.Second*30 {
		select {
		case btnEv := <-getButtonSensorChan:
			log.WithField("ButtonEvent", btnEv).Info("elevdriver TestDriver: Button Press!")
		case floor := <-getFloorSensorChan:
			log.WithField("FloorSensor", floor).Info("elevdriver TestDriver: Floor Signal!")
			if floor == 0 {
				setMotorDirChan <- elevtype.MD_Up
				log.Info("elevdriver TestDriver: Setting motor dir up")
			} else if floor == numFloors-1 {
				setMotorDirChan <- elevtype.MD_Down
				log.Info("elevdriver TestDriver: Setting motor dir down")
			}

		default:
			if time.Now().Sub(smalltimecounter) > time.Second*5 {
				smalltimecounter = time.Now()
				log.Debug("elevdriver RunFloorSensorAndMotorTest: In test loop")
			}

		}
	}
	log.Info("elevdriver TestDriver: Stopping motor")
	setMotorDirChan <- elevtype.MD_Stop

	log.Info("elevdriver RunSensorAndMotorTest: Done")

	wg.Done()

}
