package elevnetwork

import (
	"strconv"
	"sync"
	"time"

	et "../elevtype"
	l "./localip"
	p "./peers"

	log "github.com/sirupsen/logrus"
)

/*
 * Module used to monitore currently active systems in network by
 * trasmitting and receiving unique ID's. If the module stops receiving ID from a system
 * it will be removed within a defined time.
 */

////////////////////////////////
// Module variables
////////////////////////////////
const HEARTBEATINTERVAL = 5 * time.Millisecond

var systemsInNetwork et.PeerUpdate
var mutex = &sync.Mutex{}
var signalHeartBeatToStop chan bool

////////////////////////////////
// Interface
////////////////////////////////

/*StartHeartBeat (.) starts running the *hearbeat protocol
 */
func StartHeartBeat() {
	signalHeartBeatToStop = make(chan bool)
	port := 20102
	ID, err := l.LocalID()
	if err != nil {
		log.Debug("elevheartbeat startHeartBeat: Couldnt get local ip")
	}
	IDstring := strconv.FormatInt(int64(ID), 10)
	runHeartBeat(port, IDstring, signalHeartBeatToStop)

}

/*StopHeartBeat (.) stops the heartbeat protocol
 * by inserting true into singalHeartBeatToStop channel
 */
func StopHeartBeat() {
	signalHeartBeatToStop <- true
}

/* runHeartBeat (.) runs the heartbeat protocol which monitors the active participant of the netowrk.
 * Uses given peers functions from peers.go

 * @arg port: broadcast on given port
 * @arg PeersCh: List of current peers in network on channel
 * @arg heartbeatMsg: LocalID which is broadcasted
 * @arg signalHeartBeat: Cannel used to stop the protocol
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
			log.WithField("ID's:", msg.Peers).Info("Change in active systems")
			updateSystemsInNetwork(msg)

		case <-signalHeartBeatToStop:
			return

		default:
			sendPeerCh <- false
		}

	}

}

/*GetSystemsInNetwork (.) returns a list of active elevators
 * by their ID in the network
 */
func GetSystemsInNetwork() []int32 {
	mutex.Lock()
	activeSystemsString := systemsInNetwork.Peers
	mutex.Unlock()

	var activeSystems []int32
	for _, sys := range activeSystemsString {
		val, _ := strconv.Atoi(sys)
		activeSystems = append(activeSystems, int32(val))
	}

	return activeSystems
}

////////////////////////////////
// Auxiliary
////////////////////////////////

/*updateSystemsInNetwork (.) used to update the active elevators
 * in the network from each heartbeat
 */
func updateSystemsInNetwork(updatedPeers et.PeerUpdate) {
	mutex.Lock()
	systemsInNetwork = updatedPeers
	mutex.Unlock()
}
