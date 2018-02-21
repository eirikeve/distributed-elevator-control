package elevnetwork

import(
	"./conn"
	"./bcast"
	"./localip"
	"./peers"
	"time"
	"encoding/json"
	"net"
)

const HEARTBEATINTERVAL = 100*time.Millisecond
const HEARTBEATTIMEOUT = 10 * HEARTBEATINTERVAL

type struct HeartbeatMessage{
	Status string json:"status"
	// IP/Address ? 	
}

func udpSendHeartBeat(port string,peerListLocalCh chan []string){
	localIP, err := LocalIP()
	//get staus 
}
