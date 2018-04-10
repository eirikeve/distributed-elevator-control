package netack

import (
	"fmt"
	"net"
	"time"

	et "../../elevtype"
	"../bcast"
	conn "../conn"
	"../localip"
	np "../netprotocol"
	log "github.com/sirupsen/logrus"
)

type AckNackMsg struct {
	MsgType et.MessageType
	MsgData string
}

var stopAckHandlerChan chan bool
var initialized = false

func StartAckHandler(send <-chan AckNackMsg, recv chan<- AckNackMsg) {
	log.Debug("netack StartAckHandler: Starting")
	stopAckHandlerChan = make(chan bool, 2)
	initialized = true
	port := et.AckHandlerPort
	go ackHandler(send, recv, port)
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

func ackHandler(send <-chan AckNackMsg, recv chan<- AckNackMsg, port int) {
	output := make(chan string, 2)
	input := make(chan string, 2)

	go bcast.Transmitter(port, output)
	go bcast.Receiver(port, input)

	localIP, _ := localip.LocalIP()
	localAddr := localIP + fmt.Sprintf(":%d", port)

	var buf [1024]byte
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	var receivedAckNackMsgs []AckNackMsg

	for {
		// Check if we have received some ACK/NACK msgs
		readConnAndUpdateRecvMsgs(&conn, &receivedAckNackMsgs, &buf, &localAddr)
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
		case <-stopAckHandlerChan:
			return
		}
	}

}

func ackHandler2(send <-chan AckNackMsg, recv chan<- AckNackMsg, port int) {
	output := make(chan string, 2)
	input := make(chan string, 2)

	go bcast.Transmitter(port, output)
	go bcast.Receiver(port, input)

}

func readConnAndUpdateRecvMsgs(conn *net.PacketConn, receivedAckNackMsgs *[]AckNackMsg, buf *[1024]byte, localAddr *string) {
	const interval = 15 * time.Millisecond
	(*conn).SetReadDeadline(time.Now().Add(interval))
	n, addr, err := (*conn).ReadFrom((*buf)[0:])
	log.WithField("MyAddr", (*conn).LocalAddr()).Info("This address")
	log.WithField("MsgAddr", addr).Info("Receive msg src address")

	recvmsg := string((*buf)[:n])
	recvmsgData, recvmsgType, _ := np.DeFormatMessage(recvmsg)
	if err == nil && (recvmsgType == np.MsgACK || recvmsgType == np.MsgNACK) {
		msgWasSentFromThisAddr := (addr.String() == (*localAddr))
		if !msgWasSentFromThisAddr {
			(*receivedAckNackMsgs) = append((*receivedAckNackMsgs), AckNackMsg{recvmsgType, recvmsgData})
		}
		log.WithField("MsgSentFromThisAddr", msgWasSentFromThisAddr).Info("my msg?")

	}
}

func deleteElementZero(slice []AckNackMsg) {
	if len(slice) > 1 {
		slice = slice[1 : len(slice)-1]
	} else {
		slice = make([]AckNackMsg, 0)
	}
}
