package netprotocol

import (
	"errors"
	"regexp"
	"strings"

	et "../../elevtype"
	log "github.com/sirupsen/logrus"
)

var initialized = false
var msgTypeRegexp *regexp.Regexp
var msgDataRegexp *regexp.Regexp
var msgTypeStringMap map[et.MessageType]string

const msgTypeTag = "msgtype="
const msgDataTag = " msgdata="

/*MsgAsProtocol converts a string message to a message on the form
 * "msgtype=<type>, msg=<message>"
 * and returns it
 * @arg msg: A message to represented with our simple protocol
 * @arg msgType: The kind of message to be represented. An identifier.
 * @return: msg represented with our protocol
 */
func FormatForTransmission(msgData string, msgType et.MessageType) string {
	if !(initialized) {
		initMsgTypeStringMap()
		initRegexp()
		initialized = true
	}
	msgTypeString := msgTypeStringMap[msgType]
	return msgTypeTag + msgTypeString + msgDataTag + msgData
}

func DeFormatMessage(msgRecv string) (string, et.MessageType, error) {
	if !(initialized) {
		initMsgTypeStringMap()
		initRegexp()
		initialized = true
	}
	msgDataSubString := msgDataRegexp.FindAllString(msgRecv, 1)
	msgTypeSubString := msgTypeRegexp.FindAllString(msgRecv, 1)
	if msgDataSubString == nil || msgTypeSubString == nil {
		return "", et.MsgUnknown, errors.New("netprotocol DeFormatMessage: Unknown format")
	}
	msgTypeAsString := strings.TrimSuffix(strings.TrimPrefix(msgTypeSubString[0], msgTypeTag), msgDataTag)
	msgData := strings.TrimPrefix(msgDataSubString[0], msgDataTag)

	var msgType = et.MsgUnknown
	for mType, mTypeString := range msgTypeStringMap {
		if mTypeString == msgTypeAsString {
			msgType = mType
			break
		}
	}
	if msgType == et.MsgUnknown {
		return "", et.MsgUnknown, errors.New("netprotocol DeFormatMessage: Unknown msgType")
	}

	return msgData, msgType, nil

}

func initRegexp() {
	var err error
	var err1 error
	msgTypeRegexp, err = regexp.Compile("^" + msgTypeTag + ".+" + msgDataTag)
	msgDataRegexp, err1 = regexp.Compile(msgDataTag + ".+$")
	if err != nil {
		log.WithField("initErr", err.Error()).Error("netprotocol initRegexp: Could not compile")
	} else if err1 != nil {
		log.WithField("initErr", err1.Error()).Error("netprotocol initRegexp: Could not compile")
	}
}

func initMsgTypeStringMap() {
	msgTypeStringMap = make(map[et.MessageType]string)
	msgTypeStringMap[et.MsgHeartbeat] = "OP_EIRIK_MSG_HEARTBEAT"
	msgTypeStringMap[et.MsgACK] = "OP_EIRIK_MSG_ACK"
	msgTypeStringMap[et.MsgNACK] = "OP_EIRIK_MSG_NACK"
	msgTypeStringMap[et.MsgRegularUpdates] = "OP_EIRIK_MSG_REGULAR"
	msgTypeStringMap[et.MsgUnknown] = "OP_EIRIK_MSG_UNKNOWN"
}
