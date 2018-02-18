package elevnetwork

import (
	"log"
	"net"
	"sync"
	"time"
)


const (
	KEEPALIVE 	= iota
	STOP 		= iota
)
const KEEPALIVE_TIME := 2 // [seconds]



// Get preferred outbound ip of this machine
// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// http://ipengineer.net/2016/05/golang-net-package-udp-client-with-specific-source-port/
// https://groups.google.com/forum/#!msg/golang-nuts/nbmYWwHCgPc/ZBw2uH6Bdi4J
func getUDPConn(localIP string, localPort int, RemoteEPIP string, RemotePort int) net.UDPConn {
	LocalAddr, err := net.ResolveUDPAddr("udp", localIP+":50000")

	RemoteEP := net.UDPAddr{IP: net.ParseIP(RemoteEPIP), Port: 50000}

	conn, err := net.DialUDP("udp", LocalAddr, &RemoteEP)

	if err != nil {
        // handle error
	}
	return conn
}


// run go broadcasterInstance(...)
func broadcasterInstance(output chan, keepalive chan, synchronizer sync.WaitGroup){
	// WaitGroup updates for Goroutine synchronization
	synchronizer.Add(1)
	defer synchronizer.Done()

	
	localIP := getOutboundIP()
	conn := getUDPConn(localIP)
	defer conn.Close()



	watchdogKillTime := time.Now() + time.Second() * KEEPALIVE_TIME

	for {
		if (watchdogKillTime.Sub(time.Now) <= 0)
		{
			// All cleanup is done with defer
			return
		}
		select {
		case signal := <-keepalive:
			if (signal == KEEPALIVE)) {
				watchdogKillTime := time.Now() + time.Second() * KEEPALIVE_TIME
			}
			else if (signal == STOP) {
				// All cleanup is done with defer
				return
			}

		case msg := <- output
		default:
			continue
		}
	}


}