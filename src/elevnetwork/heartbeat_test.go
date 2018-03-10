package elevnetwork

import(
	"testing"
	"./localip"
	"./bcast"
	"time"
	"encoding/json"
	
	
	
)

// See if LocalIP function manages to return IP-address
func TestLocalIP(t *testing.T){
	localip,_ := localip.LocalIP()
	if localip == ""{
		t.Errorf("Could not read LocalIP ",localip)
	}else{
		println("Local IP found: ",localip)
	}

}



// Testing UDP Transmission and Reception

func TestUdpbcast(t *testing.T){

	port := 20008
	localip,_:= localip.LocalIP()
	sendMsgCh := make(chan HeartbeatMessage)
	recvMsgCh := make(chan HeartbeatMessage)

	go bcast.Transmitter(port,sendMsgCh)
	go bcast.Receiver(port,recvMsgCh)

	i:= 0
	start := time.Now()
	for {

		sendMsgCh <- HeartbeatMessage{localip}
		time.Sleep(time.Second*1)
		msg := <- recvMsgCh	
		jsonMsg,_ := json.Marshal(msg)
		println(string(jsonMsg))
		i++
		if time.Second*1 < time.Now().Sub(start){
			return		
		}
	}
}
/*
* Test for how the heartbeat protocol should approximatley be run in either
* the Network handler or Network State module.
* Makes a local list of Peers, which is continously updated from the Peers channel.
*/
func TestUdpHeartbeat(t *testing.T){
	port := 20008
	PeersCh := make(chan []Peer ) 
	go runHeartBeat(port,PeersCh)
	for{
		select{
		case  peerNetworkList:= <-PeersCh:
			for _,element :=range peerNetworkList{
				peerJsonMsg,_ := json.Marshal(element)
				println(string(peerJsonMsg))
			}
		
		default:
		}
	}
	
}
/*
 * Another test for the hearbeat protocol, but uses runHeartBeatVol2 instead of
 * runHeartHeat, which are implented somewhat differently.
 */
func TestHeartBeatVol2(t *testing.T){
	port := 20102
	ID,_ := localip.LocalIP()
	runHearBeatVol2(port,ID)
	
}