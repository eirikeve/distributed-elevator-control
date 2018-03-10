package netprotocol

import "testing"
import log "github.com/sirupsen/logrus"
import "../../elevlog"

func TestFormatAndDeformat(t *testing.T) {
	elevlog.InitLog(log.DebugLevel, false)
	msg := "this is a regular update"
	msgType := MsgRegularUpdates
	msgFormatted := FormatForTransmission(msg, msgType)
	log.WithFields(log.Fields{"msgType": msgType, "msgdata": msg}).Debug("netprotocol Test: Input")
	expectedOutput := msgTypeTag + msgTypeStringMap[msgType] + msgDataTag + msg
	log.WithField("expected output", expectedOutput).Debug("netprotocol Test: Expected")
	log.WithField("expected output", msgFormatted).Debug("netprotocol Test: Actual")
	actualOutput, msgType2, err := DeFormatMessage(msgFormatted)
	if err != nil {
		print("oops")
	}
	log.WithFields(log.Fields{"msgType": msgType2, "msg": actualOutput}).Debug("netprotocol Test: Deformatted")

	someMessage := "msgtype=NOT_A_MSG_TYPE msg=Oh no"
	_, _, err2 := DeFormatMessage(someMessage)
	if err2 != nil {
		log.WithField("err", err2.Error()).Debug("netprotocol Test: Tried deformatting unknown msg")
	}
}
