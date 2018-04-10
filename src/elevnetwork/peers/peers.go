package peers

import (
	"fmt"
	"net"
	"sort"
	"time"

	et "../../elevtype"
	"../conn"
	np "../netprotocol"
	log "github.com/sirupsen/logrus"
)

const interval = 15 * time.Millisecond
const timeout = 50 * time.Millisecond

func Transmitter(port int, id string, transmitEnable <-chan bool) {
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			id_formated := np.FormatForTransmission(id, et.MsgHeartbeat)
			conn.WriteTo([]byte(id_formated), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- et.PeerUpdate) {

	var buf [1024]byte
	var p et.PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id_formated := string(buf[:n])
		id, msgType, err := np.DeFormatMessage(id_formated)

		// Adding new connection
		p.New = ""
		// Checks if message content is valid
		if id != "" && msgType == et.MsgHeartbeat && err == nil {

			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}
			lastSeen[id] = time.Now()
		} else if id != "" && msgType != et.MsgHeartbeat {
			log.WithField("msgType", msgType).Warning("peers Receiver: Received a msgType not equal to MsgHeartbeat")
		} else if id != "" && err != nil {
			log.WithField("Error", err).Warning("peers Receiver: Received a id with err != nil")
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}
