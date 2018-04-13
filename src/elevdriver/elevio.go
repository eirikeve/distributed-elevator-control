package elevdriver

/*
Elevio contains io drivers from https://github.com/TTK4145/driver-go/blob/master/elevio/elevator_io.go
*/

import (
	"net"
	"sync"
	"time"

	def "../elevdef"

	et "../elevtype"

	log "github.com/sirupsen/logrus"
)

const _pollRate = 20 * time.Millisecond

var PortNum string // @SIM
var ioInitialized = false
var ioNumFloorsElevator int
var ioLock sync.Mutex
var ioConn net.Conn
var stopBtnWg sync.WaitGroup
var stopBtnShutdownChan = make(chan bool, 2)

func initConnectionAndSetNumFloors(addr string, NumFloorsElevator int) {
	if ioInitialized {
		log.Warning("elevdriver initConnectionAndSetNumFloors: Driver already initialized")
		return
	}

	ioNumFloorsElevator = NumFloorsElevator
	ioLock = sync.Mutex{}
	var err error
	ioConn, err = net.Dial("tcp", addr+":"+def.SystemIpPort)
	if err != nil {
		log.WithField("Err", err.Error()).Fatal("elevdriver initConnectionAndSetNumFloors: Cannot establish conn")
	}
	ioInitialized = true
}

func shutdownConnection() {
	if !ioInitialized {
		log.Warning("elevdriver initConnectionAndSetNumFloors: Not running")
		return
	}
	ioConn.Close()
	ioInitialized = false
}

func setMotorDirection(dir et.MotorDirection) {
	ioLock.Lock()
	defer ioLock.Unlock()
	ioConn.Write([]byte{1, byte(dir), 0, 0})
}

func setButtonLamp(b et.ButtonLamp) {
	// changed the input args, may be a bug here
	floor := b.Floor
	button := b.Button
	value := b.Value
	if 0 <= floor && floor < ioNumFloorsElevator {
		ioLock.Lock()
		defer ioLock.Unlock()
		ioConn.Write([]byte{2, byte(button), byte(floor), toByte(value)})
	} else {
		log.WithFields(log.Fields{"floor": floor}).Error("elevdriver SetButtonLamp: Invalid floor")
	}
}

func setFloorIndicator(floor int) {
	if 0 <= floor && floor < ioNumFloorsElevator {
		ioLock.Lock()
		defer ioLock.Unlock()
		ioConn.Write([]byte{3, byte(floor), 0, 0})
	} else {
		log.WithFields(log.Fields{"floor": floor}).Error("elevdriver SetFloorIndicator: Invalid floor")
	}

}

func setDoorOpenLamp(value bool) {
	ioLock.Lock()
	defer ioLock.Unlock()
	ioConn.Write([]byte{4, toByte(value), 0, 0})
}

func setStopLamp(value bool) {
	ioLock.Lock()
	defer ioLock.Unlock()
	ioConn.Write([]byte{5, toByte(value), 0, 0})
}

func pollButtons(receiver chan<- et.ButtonEvent, shutdown <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	prev := make([][3]bool, ioNumFloorsElevator)
	for {
		select {
		case _ = <-shutdown:
			return
		default:
			time.Sleep(_pollRate)
			for f := 0; f < ioNumFloorsElevator; f++ {
				for b := et.ButtonType(0); b < 3; b++ {
					v := getButton(b, f)
					if v != prev[f][b] && v != false {
						// This might get stuck here. Use Select? @todo

						receiver <- et.ButtonEvent{f, et.ButtonType(b)}
					}
					prev[f][b] = v
				}

			}

		}
	}
}

func pollFloorSensor(receiver chan<- int, shutdown <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	prev := -1
	for {
		select {
		case _ = <-shutdown:
			return
		default:
			time.Sleep(_pollRate)
			v := getFloor()
			if v != prev && v != -1 && 0 <= v && v < ioNumFloorsElevator {
				println("\n\n pullFloor \n\n")
				receiver <- v
			}
			prev = v
		}

	}
}

func pollStopButton(receiver chan<- bool, shutdown <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	prev := false
	for {
		select {
		case _ = <-shutdown:
			return
		default:
			time.Sleep(_pollRate)
			v := getStop()
			if v != prev {
				receiver <- v
			}
			prev = v
		}

	}
}

func pollObstructionSwitch(receiver chan<- bool, shutdown <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	prev := false
	for {
		select {
		case _ = <-shutdown:
			return
		default:
			time.Sleep(_pollRate)
			v := getObstruction()
			if v != prev {
				receiver <- v
			}
			prev = v
		}
	}
}

func getButton(button et.ButtonType, floor int) bool {
	if 0 <= floor && floor < ioNumFloorsElevator {
		ioLock.Lock()
		defer ioLock.Unlock()
		ioConn.Write([]byte{6, byte(button), byte(floor), 0})
		var buf [4]byte
		ioConn.Read(buf[:])
		return toBool(buf[1])
	}
	log.WithFields(log.Fields{"floor": floor}).Error("elevdriver getButton: Invalid floor, returning false")
	return false

}

func getFloor() int {
	ioLock.Lock()
	defer ioLock.Unlock()
	ioConn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	ioConn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	}
	return -1
}

func getStop() bool {
	ioLock.Lock()
	defer ioLock.Unlock()
	ioConn.Write([]byte{8, 0, 0, 0})
	var buf [4]byte
	ioConn.Read(buf[:])
	return toBool(buf[1])
}

func getObstruction() bool {
	ioLock.Lock()
	defer ioLock.Unlock()
	ioConn.Write([]byte{9, 0, 0, 0})
	var buf [4]byte
	ioConn.Read(buf[:])
	return toBool(buf[1])
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

func fill(a chan<- bool, val bool) {
	for {
		select {
		case a <- val:
			// do nothing
			continue
		default:
			// a is filled
			return
		}
	}
}
