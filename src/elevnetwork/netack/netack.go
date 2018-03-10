package netack

import (
	"fmt"
	"net"
	"time"

	et "../../elevtype"
	conn "../conn"
	np "../netprotocol"
	log "github.com/sirupsen/logrus"
)

type ackNackMsg struct {
	MsgType np.MessageType
	MsgData string
}

var stopAckHandlerChan chan bool
var initialized = false

func StartAckHandler(send <-chan ackNackMsg, recv chan<- ackNackMsg) {
	log.Debug("netack StartAckHandler: Starting")
	stopAckHandlerChan = make(chan bool, 2)
	initialized = true
	go ackHandler(send, recv)
}

func StopAckHandler() {
	if !initialized {
		//.... @TODO
		return
	}
	stopAckHandlerChan <- true
	log.Debug("netack StopAckHandler: Stopped")
	initialized = false
}

func ackHandler(send <-chan ackNackMsg, recv chan<- ackNackMsg) {

	var buf [1024]byte
	conn := conn.DialBroadcastUDP(et.AckHandlerPort)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", et.AckHandlerPort))
	var receivedAckNackMsgs []ackNackMsg

	for {
		// Check if we have received some ACK/NACK msgs
		readConnAndUpdateRecvMsgs(&conn, &receivedAckNackMsgs, &buf)
		// Pass ACK/NACK to nethandler
		if len(receivedAckNackMsgs) > 0 {
			select {
			case recv <- receivedAckNackMsgs[0]:
				deleteElementZero(receivedAckNackMsgs)
			}
		}
		// Send ACK/NACK if nethandler has passed one to send
		select {
		case sendmsg := <-send:
			formattedMsg := np.FormatForTransmission(sendmsg.MsgData, sendmsg.MsgType)
			conn.SetDeadline(time.Now().Add(time.Second))
			conn.WriteTo([]byte(formattedMsg), addr)
		}
	}

}

func readConnAndUpdateRecvMsgs(conn *net.PacketConn, receivedAckNackMsgs *[]ackNackMsg, buf *[1024]byte) {
	const interval = 15 * time.Millisecond
	(*conn).SetReadDeadline(time.Now().Add(interval))
	n, _, err := (*conn).ReadFrom((*buf)[0:])

	recvmsg := string((*buf)[:n])
	recvmsgData, recvmsgType, _ := np.DeFormatMessage(recvmsg)
	if err == nil && recvmsgType == np.MsgACK || recvmsgType == np.MsgNACK {
		(*receivedAckNackMsgs) = append((*receivedAckNackMsgs), ackNackMsg{recvmsgType, recvmsgData})
	}
}

func deleteElementZero(slice []ackNackMsg) {
	if len(slice) > 1 {
		slice = slice[1 : len(slice)-1]
	} else {
		slice = make([]ackNackMsg, 0)
	}
}
