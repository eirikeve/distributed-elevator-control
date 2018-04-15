package elevhandler

import (
	"testing"
	"time"

	"../elevlog"
	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

func TestElevHandler(t *testing.T) {
	elevlog.InitLog(log.DebugLevel, false)
	networkToElev := make(chan et.GeneralOrder, 12)
	elevToNetwork := make(chan et.ButtonEvent, 12)

	log.Info("elevhandler TestElevHandler: Starting test")
	StartElevatorHandler(networkToElev, elevToNetwork)
	time.Sleep(time.Second * 20)
	StopElevatorHandler()
	time.Sleep(time.Second)

	log.Info("elevhandler TestElevHandler: Done")
}

func TestElevHandlerPanic(t *testing.T) {
	// @BUG this doesn't work correctly...
	// Add time.Sleep(time.Second*10) in elevhandler instance loop, then run this test
	elevlog.InitLog(log.DebugLevel, false)
	defer func() {
		if r := recover(); r != nil {
			log.Info("elevhandler TestElevHandlerPanic: Recovered")
			StopElevatorHandler()
			time.Sleep(time.Second)
			// re-init here
		}
	}()
	networkToElev := make(chan et.GeneralOrder, 12)
	elevToNetwork := make(chan et.ButtonEvent, 12)

	log.Info("elevhandler TestElevHandlerPanic: Starting test")
	StartElevatorHandler(networkToElev, elevToNetwork)
	time.Sleep(time.Second * 10)
	defer func() {
		StopElevatorHandler()
		time.Sleep(time.Second)
	}()

	log.Info("elevhandler TestElevHandler: Done")
}

func TestElevHandlerWithNetworkSim(t *testing.T) {
	elevlog.InitLog(log.DebugLevel, false)
	networkToElev := make(chan et.GeneralOrder, 12)
	elevToNetwork := make(chan et.ButtonEvent, 12)

	log.Info("elevhandler TestElevHandlerWithNetworkTest: Starting test")
	StartElevatorHandler(networkToElev, elevToNetwork)
	go simNetworkHandler(networkToElev, elevToNetwork)
	time.Sleep(time.Second * 60)
	StopElevatorHandler()

	time.Sleep(time.Second * 3)

	log.Info("elevhandler TestElevHandler: Done")
}

/*simNetworkHandler accepts all orders and returns them to the handler
 *
 */
func simNetworkHandler(networkToElev chan<- et.GeneralOrder, elevToNetwork <-chan et.ButtonEvent) {
	var orderBuffer [1000]et.SimpleOrder
	cnt := 0
	i := 0
	startTime := time.Now()
	for time.Now().Sub(startTime) < time.Second*60 {
		log.WithField("buffer sz", i).Warn("elevhandler simNetworkHandler")
		select {
		case b := <-elevToNetwork:
			log.WithField("b", b).Debug("Button")
			order := et.SimpleOrder{Id: string(cnt), Order: et.ButtonEvent{Floor: b.Floor, Button: b.Button}}
			cnt++
			log.WithField("o", order).Debug("Button")
			if i >= 1000 {
				log.Warning("elevhandler simNetworkHandler: Buffer at 1000, ignoring order")
			} else {
				orderBuffer[i] = order
				i += 1
			}
		}
		if i > 0 {
			select {
			case networkToElev <- orderBuffer[i-1]:
				i -= 1

			}
		}
		time.Sleep(time.Millisecond * 5)
	}

}
