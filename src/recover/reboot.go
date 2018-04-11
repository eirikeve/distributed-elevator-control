package recover

import (
	"encoding/json"
	"net"
	"os/exec"
	"syscall"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

const BROADCAST_PERIOD = 300
const MSG_MISSED_THRESHOLD = 10

func StartSurveillanceOfPrimary() {
	runSurveillanceProcess()

}

func runSurveillanceProcess() {
	isBackUp := true
	isAlive := false

	var signalStopListen = make(chan int)
	var markTimeout = make(chan int)
	var lastBroadcast time.Time
	if isBackUp {
		go surveillanceListenProcess(signalStopListen, markTimeout)
	}
	<-markTimeout

	isBackUp = false
	isAlive = true

	time.Sleep(time.Millisecond * 500)

	if !(isBackUp) {
		spawnBackup()
		go func() {
			for {
				if time.Now().Sub(lastBroadcast) > time.Millisecond*BROADCAST_PERIOD {
					lastBroadcast = time.Now()
					primaryBroadcastProcess(isAlive)
				}
			}
		}()
	}
}

func surveillanceListenProcess(stop chan int, marktimeout chan int) {
	var isRunning bool
	missedMSG := 0
	localUDPAddress, err := net.ResolveUDPAddr("udp", ":"+et.BackupPort)
	if err != nil {
		log.WithField("Failed to Resolve UPD ADDR", err).Error("Reboot listenProcess")
	}
	conn, err := net.ListenUDP("udp", localUDPAddress)
	if err != nil {
		log.WithField("Failed to ListenUDP", err).Error("Reboot listenProcess")
	}
	defer conn.Close()

	var buf [1024]byte
	for {

		conn.SetDeadline(time.Now().Add(time.Millisecond * BROADCAST_PERIOD))

		n, _, err := conn.ReadFromUDP(buf[:])
		json.Unmarshal(buf[0:n], isRunning)

		if err == nil {
			log.Info("Secondary Recv message from Primary")
			missedMSG = 0
		} else {
			log.Warning("Secondary Missed msg from Primary")
			missedMSG += 1
			if missedMSG >= MSG_MISSED_THRESHOLD {
				marktimeout <- 1
				return
			}
		}

	}
}

func primaryBroadcastProcess(isAlive bool) {

	remoteUDPAddress, err := net.ResolveUDPAddr("udp4", "127.0.0.1:"+et.BackupPort)
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
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go setup.go -port="+et.SystemIpPort+" -backupPort="+et.BackupPort)).Run() // "go test")).Run()
	//For OSX:
	//(exec.Command("osascript", "-e", "tell app \"Terminal\" to do script \"go run ..."")).Run()

	log.Info("Secondary is created and is now surveillance!")
}

func Reboot() {
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "go test -run TestReboot")).Run()
	syscall.Exit(1)

}
