package elevtype

type AckNackMsg struct {
	MsgType   MessageType `json: "MsgType"`
	MsgData   string      `json: "MsgData"`
	MsgSender string      `json: "MsgSender"`
}
