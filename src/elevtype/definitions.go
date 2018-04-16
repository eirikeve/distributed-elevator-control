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

// Ports
var SystemIpPort string = "15657"
var BackupPort string = "23003"

const SystemIpPortDefault = "15657"
const BackupPortDefault = "23003"

const AckHandlerPort = 26969
const DisplayListenerPort = 28000

// Frequencies

//Main
const BackupRecoverInterval = -time.Second * 300 // Time for how far back in time the system shall look for backups when initilzing

//Elevhandler
var HandlerDebugLogMsgFreq = 2 * time.Second           // Frequency for log messages in elevhandler
var HandlerSendFSMUpdatesFreq = 100 * time.Millisecond // Frequency for updating the FSM from elevhandler

//Nethandler
var NetHandlerDebugLogMsgFreq = 2 * time.Second               // Frequency for how often debug messages are shown
var NetHandlerAutoBackupFreq = 500 * time.Millisecond         // Frequency for how often system backups are saved
var NetHandlerSendRegularUpdateFreq = 100 * time.Millisecond  // Frequency for how often local system is sent to the network
var NetHandlerSendElevatorQueueFreq = 50 * time.Millisecond   // Frequency for how often the order queue is sent to the local system
var NetHandlerSendElevatorLightsFreq = 200 * time.Millisecond // Frequency for how often the order lights are updated to the elevhandler

// Timeout
const OrderTimeoutSeconds int64 = 20 //[seconds]

type MessageType int
