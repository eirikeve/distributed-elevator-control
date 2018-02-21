package elevnetwork

import (
	"time"
	"./bcast"
	"./localip"
	"encoding/json"
	)
const HEARTBEATINTERVAL = 300*time.Millisecond
const HEARTBEATTIMEOUT = 10 * HEARTBEATINTERVAL


type HeartbeatMessage struct{
	//Status string `json:"status"`
	// If status is to be used, it must be implemented
	LocalIP string `json:"LocalIP"` 	
}

var listPeers[] HeartbeatMessage
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
	starttime := time.Now()

	sendMsgCh := make(chan HeartbeatMessage,2)
	recvMsgCh := make(chan HeartbeatMessage,2)

	for time.Now().Sub(starttime) < time.Second*3{
	

	go udpSendHeartBeat(port, sendMsgCh)
	go udpRecvHeartBeat(port,recvMsgCh)

	time.Sleep(HEARTBEATINTERVAL)
	msg := <- recvMsgCh	
	jsonMsg,_ := json.Marshal(msg)
	println("Heartbeat Message:" + string(jsonMsg))
	if listPeers != nil{
	
	for _,element :=range listPeers{
		if (element != msg){
			listPeers = append(listPeers,msg)
		}
	}
	}else {
		listPeers = append(listPeers,msg)
	}
}
	println("Loop completed!")
	for _,element :=range listPeers{
		peerJsonMsg,_ := json.Marshal(element)
		println("Current Peers ID: " + string(peerJsonMsg))
	}
	
	
}