package recover

import (
	"encoding/json"
	"net"
	"os/exec"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

const BROADCAST_PERIOD = 300
const MSG_MISSED_THRESHOLD = 10

//Starts the process for monitoring the Primary system
func StartSurveillanceOfPrimary() {
	runSurveillanceProcess()
}

/* Starts as secondary system, monitoring  the primary.
 * If primary crashes i.e. fails to send 10 continous messages, the secondary becomes
 * primary and creates a new backup
 */
func runSurveillanceProcess() {
	isBackUp := true

	var markTimeout = make(chan int)
	var lastBroadcast time.Time
	if isBackUp {
		go surveillanceListenProcess(markTimeout)
	}
	<-markTimeout

	isBackUp = false

	time.Sleep(time.Millisecond * 500)

	if !(isBackUp) {
		spawnBackup()
		go func() {
			for {
				if time.Now().Sub(lastBroadcast) > time.Millisecond*BROADCAST_PERIOD {
					lastBroadcast = time.Now()
					primaryBroadcastProcess()
				}
			}
		}()
	}
}

/* Monitors the primary process, by listening for messages
 * If 10 messages are continously dropped, the timeout is set
 * and the function returns
 * @arg marktimeout: channel written to if primary is not broadcasting
 */
func surveillanceListenProcess(marktimeout chan int) {
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

		_, _, err := conn.ReadFromUDP(buf[:])

		if err == nil {
			log.Info("Secondary: Recv message from Primary")
			missedMSG = 0
		} else {
			log.Warning("Secondary: Missed msg from Primary")
			missedMSG += 1
			if missedMSG >= MSG_MISSED_THRESHOLD {
				marktimeout <- 1
				return
			}
		}

	}
}

/* Broadcast a boolean value, enabling the monotring process
 * to know that the primary is running
 */
func primaryBroadcastProcess() {

	isAlive := true

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

// Spawns a new terminal running the main program
func spawnBackup() {

	// For Ubuntu:
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer")).Run()
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go -port="+et.SystemIpPort+" -backupPort="+et.BackupPort)).Run()
	log.Info("Secondary is created and is now surveilling!")
}
