package elevnetwork

import (
	"fmt"
	"testing"
	"time"

	"./localip"
)

// See if LocalIP function manages to return IP-address
func TestLocalIP(t *testing.T) {
	localip, _ := localip.LocalIP()
	if localip == "" {
		t.Errorf("Could not read LocalIP ", localip)
	} else {
		println("Local IP found: ", localip)
	}

}

/*
 * A Heartbeat protocol test
 * runs two runHeartBeat functions simulatainiously, with different IP
 * Both ID's should appear in list
 */
func TestHeartBeat(t *testing.T) {

	startTime := time.Now()
	go startHeartBeat()

	for {

		if time.Now().Sub(startTime) > time.Second*5 {
			stopHeartBeat()
			break

		}

	}

}

func TestMutexHeartBeat(t *testing.T) {
	startTime := time.Now()
	go startHeartBeat()

	for {
		time.Sleep(time.Millisecond * 50)
		fmt.Printf("List of active systems: %v \n", getSystemsInNetwork().Peers)

		if time.Now().Sub(startTime) > time.Second*5 {
			stopHeartBeat()
			break

		}

	}
}
