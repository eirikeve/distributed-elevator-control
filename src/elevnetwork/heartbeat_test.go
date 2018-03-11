package elevnetwork

import (
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
	go StartHeartBeat()

	for {

		if time.Now().Sub(startTime) > time.Second*5 {
			StopHeartBeat()
			break

		}

	}

}
