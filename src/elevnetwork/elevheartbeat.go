package elevnetwork

import (
	"time"
	"./bcast"
	"./localip"
	"encoding/json"
	)
// Constants
const HEARTBEATINTERVAL = 300*time.Millisecond
const HEARTBEATTIMEOUT = 10 * HEARTBEATINTERVAL


type HeartbeatMessage struct{
	//Status string `json:"status"`
	// If status is to be used, it must be implemented
	LocalIP string `json:"LocalIP"` 	
}

type Peer struct{
	ID HeartbeatMessage `json:"ID"`
	ReceiveTime time.Time `json:"ReceiveTime"`
}

// ***GLOBAL VARIABLES***

// List of Peer members of the Network, by IP
var listPeers[] Peer
// List containg information about last Heartbeat Message received by Peer members
// ***FUNCTIONS***

// Transmits a heartbeat containing a HeartbeatMessage for each HEARTBEATINTERVAL
//@arg port: Transmit on given port
func udpSendHeartBeat(port int, transMsgCh chan HeartbeatMessage){
	locIP,_ := localip.LocalIP()
	if locIP == ""{
		println("Could not find Local IP address.")
		println("No error handling impleneted")
	}else{
		//Wait time for next transmission
		time.Sleep(HEARTBEATINTERVAL)
		//Wrtie to transmission channel	
		transMsgCh <- HeartbeatMessage{locIP}

		//Transmit content on msg to port
		bcast.Transmitter(port,transMsgCh)
	}	
}

// Receive HeartbeatMessage from Broadcasters
// @arg port: Listen on given port
// @arg recvMsgCh: Receive on given channel
func udpRecvHeartBeat(port int, recvMsgCh chan HeartbeatMessage){
	//Looking to Receive message
	bcast.Receiver(port,recvMsgCh)
	recvMsg := <- recvMsgCh
	if recvMsg.LocalIP == ""{
		println("Gotcha")
	}else{
		println("Missing")
	}
}

func runHeartBeat(port int){

	startTime := time.Now()
	lastTranmissionTime := time.Now()

	sendMsgCh := make(chan HeartbeatMessage,2)
	recvMsgCh := make(chan HeartbeatMessage,2)
	
	//for time.Now().Sub(starttime) < time.Second*4{
	go udpRecvHeartBeat(port,recvMsgCh)
	for time.Now().Sub(startTime) < time.Second*10{
		
		if (time.Now().Sub(lastTranmissionTime)>HEARTBEATINTERVAL){
			go udpSendHeartBeat(port, sendMsgCh)
			lastTranmissionTime =time.Now()

		}
		
		select{
			case msg := <-recvMsgCh:
				println("Message received")
				jsonMsg,_ := json.Marshal(msg)
				println("Heartbeat Message:" + string(jsonMsg))	
				updateToPeerList(msg,&listPeers,listPeers)
				

			default:
		}

	}
	/*
	time.Sleep(HEARTBEATINTERVAL)
	msg := <- recvMsgCh	
	jsonMsg,_ := json.Marshal(msg)
	println("Heartbeat Message:" + string(jsonMsg))
	
	updateToPeerList(msg,&listPeers,listPeers)
	*/
	//}

	//Print to see if listPeers is correct
	println("Loop completed!")
	for _,element :=range listPeers{
		peerJsonMsg,_ := json.Marshal(element)
		println(string(peerJsonMsg))
	}

	
	
}	
//}

// Updates the P2P network,
// Add new received IP-address
// Update timestap from received messages
// @arg msg: Contains the Heartbeat Message that was received
// @arg, @arg pointerList, List: Contains the list of already exsisting members of the network
// Ops: The same list is taken as two different parameters to acheive call-by-reference for indexing and appending
func updateToPeerList(msg HeartbeatMessage, pointerList* []Peer, list []Peer){

	//Add new elements and update timestap from received messages
	addElement := true
	currentPeer:= Peer{msg,time.Now()}
	if(*pointerList == nil){
		
		*pointerList = append(*pointerList,currentPeer)
	}else{
		for index,element :=range *pointerList{
			if (element.ID == msg){
				addElement = false
				list[index].ReceiveTime = time.Now()
				break
				
			}
		}
		if (addElement){
			*pointerList = append(*pointerList,currentPeer)
		}
	}

	//Remove P2P members, that we have not got any response from within HARTBEATTIMEOUT
	// OPS what happen with index and element if one P2P connection is removed?!?!?
	for index,element := range list{
		if time.Now().Sub(element.ReceiveTime) >= HEARTBEATTIMEOUT{
			*pointerList = append(list[:index],list[index+1:]...)
		}
	}
}




