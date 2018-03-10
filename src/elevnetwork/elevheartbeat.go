package elevnetwork

import (
	"encoding/json"
	"fmt"
	"time"

	"./bcast"
	"./localip"
	p "./peers"
	et "../elevtype"
)

// Constants
const HEARTBEATINTERVAL = 30 * time.Millisecond
const HEARTBEATTIMEOUT = 10 * HEARTBEATINTERVAL

type HeartbeatMessage struct {
	//Status string `json:"status"`
	// If status is to be used, it must be implemented
	LocalIP string `json:"LocalIP"`
}

type Peer struct {
	ID          HeartbeatMessage `json:"ID"`
	ReceiveTime time.Time        `json:"ReceiveTime"`
}

// ***GLOBAL VARIABLES***

// List of Peer members of the Network, by IP
var listPeers []Peer

// List containg information about last Heartbeat Message received by Peer members
// ***FUNCTIONS***

// Transmits a heartbeat containing a HeartbeatMessage for each HEARTBEATINTERVAL
//@arg port: Transmit on given port
func udpSendHeartBeat(port int, transMsgCh chan HeartbeatMessage) {
	locIP, _ := localip.LocalIP()
	if locIP == "" {
		println("Could not find Local IP address.")
		println("No error handling impleneted")
	} else {
		//Wait time for next transmission
		time.Sleep(HEARTBEATINTERVAL)
		//Wrtie to transmission channel
		transMsgCh <- HeartbeatMessage{locIP}

		//Transmit content on msg to port
		bcast.Transmitter(port, transMsgCh)
	}
}

/* Receive HeartbeatMessage from broadcasters
* @arg port: Listen on given port
* @arg recvMsgCh: Receive on given channel
 */
func udpRecvHeartBeat(port int, recvMsgCh chan HeartbeatMessage) {
	//Listening on network and receives message
	bcast.Receiver(port, recvMsgCh)

	/* Not needed? [Todo]
	recvMsg := <- recvMsgCh
	if recvMsg.LocalIP == ""{
		println("Gotcha")
	}else{
		println("Missing")
	}*/

}

/*
* Runs the heartbeat protocol, which monitors the participants of
* the network.
* @arg port: Broadcast on given port
* @arg exsistingPeersCh: List of Peers in newtork Channel			[Todo: Missing - but will most likely be implemented]
 */
func runHeartBeat(port int, existingPeersCh chan []Peer) {

	startTime := time.Now()
	lastTranmissionTime := time.Now()

	sendMsgCh := make(chan HeartbeatMessage, 2)
	recvMsgCh := make(chan HeartbeatMessage, 2)

	go udpRecvHeartBeat(port, recvMsgCh)
	for time.Now().Sub(startTime) < time.Second*10 {

		if time.Now().Sub(lastTranmissionTime) > HEARTBEATINTERVAL {
			go udpSendHeartBeat(port, sendMsgCh)
			lastTranmissionTime = time.Now()

		}

		select {
		case msg := <-recvMsgCh:
			println("Messagtime.Now().Sub(startTime) < time.Second*10e received")
			jsonMsg, _ := json.Marshal(msg)
			println("Heartbeat Message:" + string(jsonMsg))
			addToPeerList(msg, &listPeers, listPeers)
			existingPeersCh <- listPeers
		default:
			updateToPeersList(&listPeers, listPeers)
			existingPeersCh <- listPeers

		}

	}

	//Print to see if listPeers is correct
	println("Loop completed!")
	for _, element := range listPeers {
		peerJsonMsg, _ := json.Marshal(element)
		println(string(peerJsonMsg))
	}

}

//}
/*
 * THIS IS THE VERSION WE'RE USING

 * Runs the heartbeat protocol wich nitors the participant of the netowrk.
 * Version 2: Uses the given peers functions
 * @arg port: roadcast on given port
 * @arg PeersCh: List of current peers in network on channel
 *
 */
func runHeartBeatVol2(port int, heartbeatMsg string) {

	startTime := time.Now()
	lastTranmissionTime := time.Now()

	recvPeerCh := make(chan et.PeerUpdate)
	sendPeerCh := make(chan bool)


	go p.Transmitter(port, heartbeatMsg, sendPeerCh)
	go p.Receiver(port, recvPeerCh)

	for time.Now().Sub(startTime) < time.Second*10 {

		if time.Now().Sub(lastTranmissionTime) > HEARTBEATINTERVAL {
			sendPeerCh <- true
			lastTranmissionTime = time.Now()
		}
		select {

		case msg := <-recvPeerCh:
			fmt.Printf("ID: %v \n", msg.Peers)

		default:
			sendPeerCh <- false
		}

	}

}

/*
* Updates the P2P network,
* Add new received IP-address
* Update timestap from received messages
* @arg msg: Contains the Heartbeat Message that was received
* @arg, @arg pointerList, List: Contains the list of already exsisting members of the network
* [Todo]Ops: The same list is taken as two different parameters to acheive call-by-reference for indexing and appending
 */
func addToPeerList(msg HeartbeatMessage, pointerList *[]Peer, list []Peer) {

	//Add new elements and update timestap from received messages
	addElement := true
	currentPeer := Peer{msg, time.Now()}
	if *pointerList == nil {

		*pointerList = append(*pointerList, currentPeer)
	} else {
		for index, element := range *pointerList {
			if element.ID == msg {
				addElement = false
				list[index].ReceiveTime = time.Now()
				break

			}
		}
		if addElement {
			*pointerList = append(*pointerList, currentPeer)
		}
	}

}

//Remove P2P members, that we have not got any response from within HARTBEATTIMEOUT
// OPS what happen with index and element if one P2P connection is removed?!?!?
func updateToPeersList(pointerList *[]Peer, list []Peer) {
	for index, element := range list {
		if time.Now().Sub(element.ReceiveTime) >= HEARTBEATTIMEOUT {
			*pointerList = append(list[:index], list[index+1:]...)
		}
	}
}
