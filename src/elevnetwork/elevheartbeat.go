package elevnetwork

import (
	"fmt"
	"time"

	p "./peers"
	et "../elevtype"
	l "./localip"

	log "github.com/sirupsen/logrus"
)

// Constants
const HEARTBEATINTERVAL = 30 * time.Millisecond
const HEARTBEATTIMEOUT = 10 * HEARTBEATINTERVAL

// Variable
var signalHeartBeatToStop chan bool


func startHeartBeat(){
	signalHeartBeatToStop = make(chan bool)
	port := 20102
	ID,err  := l.LocalIP()
	if err != nil{
		log.Debug("elevheartbeat startHeartBeat: Couldnt get local ip")
	}
	runHeartBeat(port,ID, signalHeartBeatToStop)


}

func stopHeartBeat(){
	signalHeartBeatToStop <- true
}
/*
 * Runs the heartbeat protocol wich monitors the participant of the netowrk.
 * Version 2: Uses the given peers functions
 * @arg port: broadcast on given port
 * @arg PeersCh: List of current peers in network on channel
 */
func runHeartBeat(port int, heartbeatMsg string, signalHeartBeat <- chan bool) {
	lastTranmissionTime := time.Now()

	recvPeerCh := make(chan et.PeerUpdate)
	sendPeerCh := make(chan bool)


	go p.Transmitter(port, heartbeatMsg, sendPeerCh)
	go p.Receiver(port, recvPeerCh)

	for {
		

		if time.Now().Sub(lastTranmissionTime) > HEARTBEATINTERVAL {
			sendPeerCh <- true
			lastTranmissionTime = time.Now()
		}

		select {

		case msg := <-recvPeerCh:
			fmt.Printf("ID: %v \n", msg.Peers)
		
		case <-signalHeartBeatToStop:
			return

		default:
			sendPeerCh <- false
		}

	}
	

}
