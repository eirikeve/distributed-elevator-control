package phoenix

import (
	"encoding/json"
	"net"
	"os/exec"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

/*
 * Description a process for monitoring the Primary system. If the primary systems stops
 * responding it will be rebooted by a secondary system, which becomes primary and spawns
 * a new secondary system
 */

////////////////////////////////
// Module variables
////////////////////////////////

const BROADCAST_PERIOD = 300
const MSG_MISSED_THRESHOLD = 10

////////////////////////////////
// Interface
////////////////////////////////

/*StartPheonicService starts as a secondary system, monitoring  the primary.
 * If primary does not exsist, crashes, etc. i.e. fails to send 10 continous messages, the secondary becomes
 * primary and creates a new backup
 */
func StartPhoenixService() {
	isBackUp := true

	var markTimeout = make(chan int)
	var lastBroadcast time.Time
	if isBackUp {
		go phoenixListenProcess(markTimeout)
	}
	<-markTimeout
	log.Warn("phoenix: No primary, taking over as primary and spawing secondary")
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

////////////////////////////////
// Auxiliary
////////////////////////////////

/* phoenixListenProcess monitors the primary process, by listening for messages
 * If 10 messages are continously dropped, the timeout is set
 * and the function returns
 * @arg marktimeout: channel written to if primary is not broadcasting
 */
func phoenixListenProcess(marktimeout chan int) {
	missedMSG := 0
	localUDPAddress, err := net.ResolveUDPAddr("udp", ":"+et.BackupPort)
	if err != nil {
		log.WithField("phoenix: Failed to Resolve UPD ADDR", err).Error("Reboot listenProcess")
	}
	conn, err := net.ListenUDP("udp", localUDPAddress)
	if err != nil {
		log.WithField("phoenix: Failed to ListenUDP", err).Error("Reboot listenProcess")
	}
	defer conn.Close()

	var buf [1024]byte
	for {

		conn.SetDeadline(time.Now().Add(time.Millisecond * BROADCAST_PERIOD))

		_, _, err := conn.ReadFromUDP(buf[:])

		if err == nil {
			log.Info("phoenix Secondary: Recv message from Primary")
			missedMSG = 0
		} else {
			log.Info("phoenix Secondary: Missed msg from Primary")
			missedMSG += 1
			if missedMSG >= MSG_MISSED_THRESHOLD {
				marktimeout <- 1
				return
			}
		}

	}
}

/* primaryBroadcastProcess broadcast a boolean value, enabling the monotring process
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

// spwanBackup spawns a new terminal running the main program
func spawnBackup() {

	// For Ubuntu:
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer")).Run()
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "./elevator -port="+et.SystemIpPort+" -backupPort="+et.BackupPort)).Run()
	log.Info("phoenix: Secondary is created and is now surveilling!")
}
