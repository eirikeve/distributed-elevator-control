package netprotocol

type MessageType int

const (
	MsgHeartbeat      MessageType = iota
	MsgACK            MessageType = iota
	MsgRegularUpdates MessageType = iota
)

/*MsgAsProtocol converts a string message to a message on the form
 * "msgtype=<type>, msg=<message>"
 * and returns it
 * @arg msg: A message to represented with our simple protocol
 * @arg msgType: The kind of message to be represented. An identifier.
 * @return: msg represented with our protocol
 */
/*func MsgAsProtocol(msg string, msgType MessageType) string {
	msgTypeString := getMsgTypeString(msgType)

	return "msgtype=" + msgTypeString + ", msg=" + msg
}*/

// Switches a MessageType to its string representation
func getMsgTypeString(msgType MessageType) string {
	switch msgType {
	case MsgHeartbeat:
		return "OP_EIRIK_MSG_HEARTBEAT"
	case MsgACK:
		return "OP_EIRIK_MSG_ACK"
	case MsgRegularUpdates:
		return "OP_EIRIK_MSG_REGULAR"
	default:
		return "OP_EIRIK_MSG_UNKNOWN"
	}
}
