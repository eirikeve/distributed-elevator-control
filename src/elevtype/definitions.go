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

var SystemIpPort string = "15657"
var BackupPort string = "23003"

const SystemIpPortDefault = "15657"
const BackupPortDefault = "23003"
const AckHandlerPort = 26969
const DisplayListenerPort = 28000

const BackupRecoverInterval = -time.Second * 300

const OrderTimeoutSeconds int64 = 20                  //seconds
const RedelegationTimeLimitFromOriginalRecv int64 = 0 // seconds
const RedelegationTimeLimitFromOtherSys int64 = 5

type MessageType int

const (
	MsgHeartbeat      MessageType = iota
	MsgACK            MessageType = iota
	MsgNACK           MessageType = iota
	MsgRegularUpdates MessageType = iota
	MsgUnknown        MessageType = iota
)
