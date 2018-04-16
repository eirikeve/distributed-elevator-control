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
	var handlerDebugLogMsgFreq = 2 * time.Second	// Frequency for log messages in elevhandler	
	var sendFSMUpdatesFreq = 100 * time.Millisecond // Frequency for updating the FSM from elevhandler

	//Nethandler
	var netHandlerDebugLogMsgFreq = 2 * time.Second 				// Frequency for how often debug messages are shown 
	var netHandlerAutoBackupFreq = 500 * time.Millisecond   		// Frequency for how often system backups are saved
	var netHandlerSendRegularUpdateFreq = 100 * time.Millisecond	// Frequency for how often local system is sent to the network 
	var netHandlerSendElevatorQueueFreq = 50 * time.Millisecond     // Frequency for how often the order queue is sent to the local system 
	var netHandlerSendElevatorLightsFreq = 200 * time.Millisecond   // Frequency for how often the order lights are updated to the elevhandler


// Timeout
const OrderTimeoutSeconds int64 = 20 //[seconds]

type MessageType int


