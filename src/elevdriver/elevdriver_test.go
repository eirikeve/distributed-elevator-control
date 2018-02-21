package elevdriver

import (
	"sync"
	"time"

	"../elevlog"

	"testing"

	"../elevtype"
	log "github.com/sirupsen/logrus"
)

/*
TestLog tests basic log functionality, and shows how to use the log system
*/
func TestDriver(*testing.T) {
	elevlog.InitLog(log.DebugLevel, true)

	log.Info("elevdriver TestDriver: Starting test of elevdriver and elevio")

	// WaitGroup for syncing upon shutdown
	wg := &sync.WaitGroup{}

	// Elevator control
	setMotorDirChan := make(chan elevtype.MotorDirection)
	setButtonLampChan := make(chan elevtype.ButtonLamp, 12)
	setFloorIndicatorChan := make(chan int)
	setDoorOpenLampChan := make(chan bool)
	setStopLampChan := make(chan bool)

	// Elevator sensors
	getButtonSensorChan := make(chan elevtype.ButtonEvent)
	getFloorSensorChan := make(chan int)
	getStopButtonSensorChan := make(chan bool)
	getObstructionSensorChan := make(chan bool)

	// Stop signal to Driver
	signalStopDriver := make(chan bool)

	// Number of floors we're using (m=4)
	numFloors := 4

	log.Debug("elevdriver TestDriver: Initialized vars")

	Driver(setMotorDirChan,
		setButtonLampChan,
		setFloorIndicatorChan,
		setDoorOpenLampChan,
		setStopLampChan,
		getButtonSensorChan,
		getFloorSensorChan,
		getStopButtonSensorChan,
		getObstructionSensorChan,
		signalStopDriver,
		numFloors,
		wg)

	log.Info("elevdriver TestDriver: Initialized Driver, polling sensors for 10 seconds")
	timestamp := time.Now()
	for time.Now().Sub(timestamp) < time.Second*10 {
		select {
		case btnEv := <-getButtonSensorChan:
			log.WithField("ButtonEvent", btnEv).Info("Button Press!")
		case floor := <-getFloorSensorChan:
			log.WithField("FloorSensor", floor).Info("Floor Signal!")
		case stop := <-getStopButtonSensorChan:
			log.WithField("StopButton", stop).Info("Stop Signal!")
		case obstr := <-getObstructionSensorChan:
			log.WithField("Obstruction", obstr).Info("Obstruction!")
		}
	}
	log.Info("elevdriver TestDriver: Setting Elev params for 10 seconds")

	// Move up
	log.Info("elevdriver TestDriver: Motor Dir Up")
	setMotorDirChan <- elevtype.MD_Up
	time.Sleep(time.Second)

	// Turn on inside lights
	var light = elevtype.ButtonLamp{
		Floor:  1,
		Button: elevtype.BT_Cab,
		Value:  true}
	ctxLogger := log.WithFields(log.Fields{"Light": &light})
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 2
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	time.Sleep(time.Second)
	light.Floor = 3
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	time.Sleep(time.Second)
	light.Floor = 4
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	time.Sleep(time.Second)

	// Move down
	log.Info("elevdriver TestDriver: Motor Dir Down")
	setMotorDirChan <- elevtype.MD_Down
	time.Sleep(time.Second)

	// Turn off inside lights
	light.Value = false
	light.Floor = 1
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 2
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 3
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 4
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light

	// Turn on outside lights
	light.Button = elevtype.BT_HallUp
	light.Value = true
	light.Floor = 1
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 2
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 3
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 4
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light

	time.Sleep(time.Second)

	// Stop moving
	log.Info("elevdriver TestDriver: Motor Dir Stop")
	setMotorDirChan <- elevtype.MD_Stop
	time.Sleep(time.Second)

	// Turn off outside lights
	light.Button = elevtype.BT_HallUp
	light.Value = false
	light.Floor = 1
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 2
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 3
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light
	light.Floor = 4
	ctxLogger.Debug("elevdriver TestDriver: Setting Light")
	setButtonLampChan <- light

	log.Info("elevdriver TestDriver: Finished setting parameters.")

}
