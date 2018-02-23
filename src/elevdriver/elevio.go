package elevdriver

/*
Elevio contains io drivers from https://github.com/TTK4145/driver-go/blob/master/elevio/elevator_io.go
*/

import "time"
import "sync"
import "net"

import "../elevtype"

import log "github.com/sirupsen/logrus"

const _pollRate = 20 * time.Millisecond

var _initialized = false
var _numFloorsElevator = stdNumFloorsElevator
var _mtx sync.Mutex
var _conn net.Conn

func initConnectionAndSetNumFloors(addr string, NumFloorsElevator int) {
	if _initialized {
		log.Warning("elevdriver Init: Driver already initialized")
		return
	}
	_numFloorsElevator = NumFloorsElevator
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		log.WithField("Err", err.Error()).Fatal("elevdriver Init: Cannot establish conn")
	}
	_initialized = true
}

func SetMotorDirection(dir elevtype.MotorDirection) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(b elevtype.ButtonLamp) {
	// changed the input args, may be a bug here
	floor := b.Floor
	button := b.Button
	value := b.Value
	if 0 <= floor && floor < _numFloorsElevator {
		_mtx.Lock()
		defer _mtx.Unlock()
		_conn.Write([]byte{2, byte(button), byte(floor), toByte(value)})
	} else {
		log.WithFields(log.Fields{"floor": floor}).Error("elevdriver SetButtonLamp: Invalid floor")
	}
}

func SetFloorIndicator(floor int) {
	if 0 <= floor && floor < _numFloorsElevator {
		_mtx.Lock()
		defer _mtx.Unlock()
		_conn.Write([]byte{3, byte(floor), 0, 0})
	} else {
		log.WithFields(log.Fields{"floor": floor}).Error("elevdriver SetFloorIndicator: Invalid floor")
	}

}

func SetDoorOpenLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- elevtype.ButtonEvent, shutdown <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	prev := make([][3]bool, _numFloorsElevator)
	for {
		select {
		case _ = <-shutdown:
			return
		default:
			time.Sleep(_pollRate)
			for f := 0; f < _numFloorsElevator; f++ {
				for b := elevtype.ButtonType(0); b < 3; b++ {
					v := getButton(b, f)
					if v != prev[f][b] && v != false {
						// This might get stuck here. Use Select? @todo

						receiver <- elevtype.ButtonEvent{f, elevtype.ButtonType(b)}
					}
					prev[f][b] = v
				}

			}

		}
	}
}

func PollFloorSensor(receiver chan<- int, shutdown <-chan bool, wg *sync.WaitGroup) {
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
			if v != prev && v != -1 && 0 <= v && v < _numFloorsElevator {
				receiver <- v
			}
			prev = v
		}

	}
}

func PollStopButton(receiver chan<- bool, shutdown <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	prev := false
	for {
		select {
		case _ = <-shutdown:
			return
		}

		time.Sleep(_pollRate)
		v := getStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool, shutdown <-chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	prev := false
	for {
		select {
		case _ = <-shutdown:
			return
		}

		time.Sleep(_pollRate)
		v := getObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v

	}
}

func getButton(button elevtype.ButtonType, floor int) bool {
	if 0 <= floor && floor < _numFloorsElevator {
		_mtx.Lock()
		defer _mtx.Unlock()
		_conn.Write([]byte{6, byte(button), byte(floor), 0})
		var buf [4]byte
		_conn.Read(buf[:])
		return toBool(buf[1])
	} else {
		log.WithFields(log.Fields{"floor": floor}).Error("elevdriver getButton: Invalid floor, returning false")
		return false
	}

}

func getFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}

func getStop() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{8, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func getObstruction() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{9, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
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
