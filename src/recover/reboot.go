package recover

import (
	"os/exec"
	"syscall"
	"net"
	"time"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	
)

const BROADCAST_PERIOD = 500 
const MSG_MISSED_THRESHOLD = 10
const BACKUP_PORT = "23200"


func RunBackUpProcess(){
	isBackUp := true

	var signalStopListen = make(chan int)
	var markTimeout = make(chan int)
	var isAlive bool
	var lastBroadcast time
	if isBackUp{
		go backUpListenProcess(signalStopListen,markTimeout)
	}
	<-markTimeout
	
	isBackUp = false
	isAlive = true

	time.Sleep(time.Millisecond * 500)

	if !(isBackUp){
		spawnBackup()
		for{
			if (time.Now().Sub(lastBroadcast) < 0){
			lastBroadcast = time.Now()
			PrimaryBroadcastProcess(isAlive)
			}
		}
	}

}

func backUpListenProcess(stop chan int, marktimeout chan int){
	var isRunning bool
	localUDPAddress, err := net.ResolveUDPAddr("udp",BACKUP_PORT)
	if err != nil {
		log.WithField("Failed to Resolve UPD ADDR",err).Error("Reboot listenProcess")
	}
	conn, err := net.ListenUDP("udp",localUDPAddress)
	if err != nil{
		log.WithField("Failed to ListenUDP",err).Error("Reboot listenProcess")
	}
	defer conn.Close()
	
	var buf [1024]byte
	for {
		conn.SetDeadline(time.Now().Add(time.Millisecond*BROADCAST_PERIOD *3))

		n,_,err := conn.ReadFromUDP(buf[:])
		json.Unmarshal(buf[0:n],isRunning)

		if err == nil{
			println("Recv message")
		}else {
			marktimeout <-1
			return
		}



	}
}

func PrimaryBroadcastProcess(isAlive bool) {

	remoteUDPAddress, err := net.ResolveUDPAddr("udp4", "127.0.0.1:"+BACKUP_PORT)
	if err != nil {
		println(err.Error())
	}
	conn, err := net.DialUDP("udp4", nil, remoteUDPAddress)
	if err != nil {
		println(err.Error())
	}
	defer conn.Close()

	jsonBuf, _ := json.Marshal(isAlive)
	conn.Write(jsonBuf)
}

func spawnBackup() {
	// For Ubuntu:
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go")).Run()
	//For OSX:
	//(exec.Command("osascript", "-e", "tell app \"Terminal\" to do script \"go run ..."")).Run()

	println("Back up is created and is now running!")
}

func Reboot(){
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "go test")).Run()
	syscall.Exit(1)

}