package elevtype

import (
	"time"
)

//Debugging: if true, log trace statements
const VerboseDebugLog = false

// Elevator
const NumFloors = 4
const NumButtons = 3
const BOTTOMFLOOR = 0
const TOPFLOOR = 3

const AckHandlerPort = 26969

const BackupRecoverInterval = -time.Second * 20

type MessageType int

const (
	//MsgHeartbeat      MessageType = iota
	MsgACK  MessageType = iota
	MsgNACK MessageType = iota
	//MsgRegularUpdates MessageType = iota
	//MsgUnknown        MessageType = iota
)
