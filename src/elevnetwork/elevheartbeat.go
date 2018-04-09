package elevnetwork

import (
	"fmt"
	"sync"
	"time"

	et "../elevtype"
	l "./localip"
	p "./peers"

	log "github.com/sirupsen/logrus"
)

var systemsInNetwork et.PeerUpdate
var mutex = &sync.Mutex{}

// Constants
const HEARTBEATINTERVAL = 15 * time.Millisecond
const HEARTBEATTIMEOUT = 30 * HEARTBEATINTERVAL

// Variable
var signalHeartBeatToStop chan bool

func StartHeartBeat() {
	signalHeartBeatToStop = make(chan bool)
	port := 20102
	ID, err := l.LocalIP()
	if err != nil {
		log.Debug("elevheartbeat startHeartBeat: Couldnt get local ip")
	}
	runHeartBeat(port, ID, signalHeartBeatToStop)

}

func StopHeartBeat() {
	signalHeartBeatToStop <- true
}

/*
 * Runs the heartbeat protocol wich monitors the participant of the netowrk.
 * Version 2: Uses the given peers functions
 * @arg port: broadcast on given port
 * @arg PeersCh: List of current peers in network on channel
 */
func runHeartBeat(port int, heartbeatMsg string, signalHeartBeat <-chan bool) {
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
			updateSystemsInNetwork(msg)

		case <-signalHeartBeatToStop:
			return

		default:
			sendPeerCh <- false
		}

	}

}

/*
 * Returns the active elevators by their ID in the network
 */
func GetSystemsInNetwork() []string {
	mutex.Lock()
	readActiveSystems := systemsInNetwork.Peers
	mutex.Unlock()
	return readActiveSystems
}

/*
 * Updates the active elevators in the network for each Heartbeat
 */
func updateSystemsInNetwork(updatedPeers et.PeerUpdate) {
	mutex.Lock()
	systemsInNetwork = updatedPeers
	mutex.Unlock()
}
