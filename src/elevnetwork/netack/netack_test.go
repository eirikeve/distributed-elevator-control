package netack

import (
	"testing"
	"time"

	np "../netprotocol"
	log "github.com/sirupsen/logrus"
)

func TestAckHandler(t *testing.T) {
	send1 := make(chan AckNackMsg, 2)
	send2 := make(chan AckNackMsg, 2)
	recv1 := make(chan AckNackMsg, 2)
	recv2 := make(chan AckNackMsg, 2)
	port1 := 20008
	port2 := 20009
	go startTestInstances(send1, recv1, send2, recv2, port1, port2)
	msg1 := AckNackMsg{np.MsgACK, "This Is ID of ACK, msg1"}
	msg2 := AckNackMsg{np.MsgNACK, "This is ID of NACK, msg2"}
	msg3 := AckNackMsg{np.MsgACK, "ssgfsefec, msg3"}
	msg4 := AckNackMsg{np.MsgNACK, "adfsgdfda msg4"}

	log.WithField("msg1", msg1).Info("netack Test: Start")
	log.WithField("msg2", msg2).Info("netack Test: Start")

	log.Info("netack Test: Sending messages")

	send1 <- msg1
	time.Sleep(time.Second)
	send2 <- msg2
	log.Info("netack Test: Passed messages to Instances")

	timestamp := time.Now()
	for time.Now().Sub(timestamp) < time.Second {
		select {
		case msg1 = <-recv1:
			log.Info("Received message from instance 1")
		case msg2 = <-recv2:
			log.Info("Received message from instance 2")
		default:
		}
	}

	log.WithField("msg3", msg3).Info("netack Test: Result")
	log.WithField("msg4", msg4).Info("netack Test: Result")

	log.Info("Exited?")
	stopTestInstances()

	log.Info("Exited")

}

func startTestInstances(send1 <-chan AckNackMsg, recv1 chan<- AckNackMsg,
	send2 <-chan AckNackMsg, recv2 chan<- AckNackMsg,
	port1 int, port2 int) {
	stopAckHandlerChan = make(chan bool, 2)
	go ackHandler(send1, recv1, port1)
	go ackHandler(send2, recv2, port2)
}

func stopTestInstances() {
	stopAckHandlerChan <- true
	stopAckHandlerChan <- true
	time.Sleep(time.Millisecond * 100)
}
