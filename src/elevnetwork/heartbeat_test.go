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

func TestUdpHeartbeat(t *testing.T){
	port := 20008
	runHeartBeat(port)
	
}