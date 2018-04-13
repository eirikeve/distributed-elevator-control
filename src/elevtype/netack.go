package elevtype

type AckNackMsg struct {
	MsgType   MessageType `json: "MsgType"`
	MsgData   string      `json: "MsgData"`
	MsgSender string      `json: "MsgSender"`
}

const (
	MsgHeartbeat      MessageType = iota
	MsgACK            MessageType = iota
	MsgNACK           MessageType = iota
	MsgRegularUpdates MessageType = iota
	MsgUnknown        MessageType = iota
)

type MessageType int
