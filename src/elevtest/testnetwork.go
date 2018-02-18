package elevtest

import "testing"
import "../elevnetwork"
import "sync"

// TestBroadcastAndListener tests functionality of Listener and Pinger in ../elevnetwork/elevcomm.go
func TestBroadcastAndListener(t *testing.T) {
	t.Log("Initializing variables\n")
	var s sync.WaitGroup

	keepalivelistener := make(chan int)
	keepalivebroadcaster := make(chan int)
	sendmessages := make(chan byte, 2048)
	recvmessages := make (chan byte, 2048)

	Listener l := elevnetwork.GeneralListener(s, recvmessages, keepalivelistener)
	Pinger b := elevnetwork.GeneralBroadcaster(s, sendmessages, keepalivebroadcaster)

	t.Log("Starting test\n")


}
