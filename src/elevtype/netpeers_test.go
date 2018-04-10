package elevtype

import (
	"testing"

	"../elevlog"
	log "github.com/sirupsen/logrus"
)

func TestNetPeers(t *testing.T) {
	var p PeerUpdate
	log.WithField("Peers", p.Peers).Info("Does this work")
	p.Peers = make([]string, 10)
	p.Peers[0] = "Text"
	p.Peers[9] = "End"
	n := len(p.Peers)
	print(p.Peers[0])
	log.WithField("Peers", p.Peers).Info("Does this work")

	for i := 0; i < n; i++ {
		print(p.Peers[i])
	}
}

func TestPeerUpdate(t *testing.T) {
	elevlog.InitLog(log.DebugLevel, false)
	var p1 PeerUpdate
	var p2 PeerUpdate

	p1.Peers = make([]string, 1)
	p1.Peers[0] = "1"

	p2.Peers = make([]string, 2)
	p2.Peers[0] = "1"
	p2.Peers[1] = "2"
	println("First values:")
	printPeers(p1.Peers)
	printPeers(p2.Peers)
	p1.UpdatePeers(p2)
	println("After first update:")
	printPeers(p1.Peers)
	p2.Peers[1] = "3"
	p2.Peers[0] = "2"
	p1.UpdatePeers(p2)
	println("After 2nd update:")
	println("Peers: ")
	printPeers(p1.Peers)
	println("Lost: ")
	printPeers(p1.Lost)
	p2.Peers = make([]string, 0)
	p1.UpdatePeers(p2)
	println("After 3rd update:")
	println("Peers: ")
	printPeers(p1.Peers)
	println("Lost: ")
	printPeers(p1.Lost)
	p2.Peers = make([]string, 2)
	p2.Peers[1] = "3"
	p2.Peers[0] = "4"
	p1.UpdatePeers(p2)
	println("After 4th update:")
	println("Peers: ")
	printPeers(p1.Peers)
	println("Lost: ")
	printPeers(p1.Lost)

}

func printPeers(p []string) {
	for i := 0; i < len(p); i++ {
		println(p[i])
	}
}
