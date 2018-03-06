package elevtimer

import (
	"testing"
	"time"

	"../elevlog"
	log "github.com/sirupsen/logrus"
)

func TestStart(*testing.T) {
	elevlog.InitLog(log.DebugLevel, false)
	println("Starting")

	var signal1 = make(chan bool)
	var signal2 = make(chan bool)
	var signal3 = make(chan bool)
	var signalStop = make(chan bool, 3)

	go channelListener(signal1, signalStop, "signal1")
	go channelListener(signal2, signalStop, "signal2")

	log.Debug("elevtimer TestTimer: Starting Timers of 2 and 5 seconds")
	Start(signal1, time.Second*2, "signal1 Timer")
	Start(signal2, time.Second*5, "signal2 Timer")

	time.Sleep(time.Second * 6)

	log.Debug("elevtimer TestTimer: Starting a timer")
	Start(signal3, time.Second, "signal3 Timer")
	select {
	case signal3 <- true:
	default:
		log.Debug("elevtimer TestTimer: Filling timer output channel, so it will produce error")
	}
	time.Sleep(3 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case signalStop <- true:
		default:
			log.Debug("elevtimer TestTimer: Sending stop signal to listeners")
		}
	}
	time.Sleep(time.Second)
	log.Debug("elevtimer TestTimer: Exiting")
	return
}

func TestUpdateAndStop(*testing.T) {
	elevlog.InitLog(log.DebugLevel, false)
	println("Starting")

	var signal1 = make(chan bool)
	var signal2 = make(chan bool)
	var signal3 = make(chan bool)
	var signalStop = make(chan bool, 3)

	go channelListener(signal1, signalStop, "signal1")
	go channelListener(signal2, signalStop, "signal2")
	go channelListener(signal3, signalStop, "signal3")

	log.Info("Testing Stop and Update with non-created channels")
	Stop("Some timer")
	Update("Another timer", time.Second)
	time.Sleep(time.Second * 2)

	log.Info("Testing Update and Stop")
	Start(signal1, time.Second, "Timer1")
	Start(signal2, time.Second, "Timer2")
	Start(signal3, time.Second, "Timer3")
	Update("Timer1", time.Second*3)
	Stop("Timer2")

	time.Sleep(3 * time.Second)

}

func channelListener(channel chan bool, signalStop chan bool, channelName string) {
	for {
		select {
		case <-channel:
			log.WithField("channelName", channelName).Info("elevtimer TestTimer: Recv signal")
		case <-signalStop:
			log.WithField("channelName", channelName).Info("elevtimer TestTimer: Recv stop signal, exiting.")
			return
		}
	}
}
