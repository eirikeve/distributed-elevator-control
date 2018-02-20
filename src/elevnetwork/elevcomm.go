package elevnetwork

import (
	"log"
	"net"
	"sync"
	"time"
)

const (
	KEEPALIVE = iota
	STOP      = iota
)
const KEEPALIVE_TIME = 2   // [seconds]
const MESSAGE_DEADLINE = 2 // [seconds]
const LISTEN_DURATION = 2  // [seconds]

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
// https://groups.google.com/fo errrum/#!msg/golang-nuts/nbmYWwHCgPc/ZBw2uH6Bdi4J
func getUDPConn(localUDPAddr net.UDPAddr, remoteUDPAddr net.UDPAddr) *net.UDPConn {
	conn, err := net.DialUDP("udp", &localUDPAddr, &remoteUDPAddr)

	if err != nil {
		// handle error
	}
	return conn
}

// Use unbuffered channel for Keepalive, so that we don't get msg congestion.
type Pinger struct {
	WG                sync.WaitGroup
	Messages          chan []byte
	Keepalive         chan int
	LocalUDPAddr      net.UDPAddr
	RemoteUDPAddr     net.UDPAddr
	StayAliveDuration time.Duration
	StayAliveTimeout  time.Time
	MessageDeadline   time.Duration
}

// Use unbuffered channel for Keepalive, so that we don't get msg congestion.
type Listener struct {
	WG                sync.WaitGroup
	Messages          chan []byte
	Keepalive         chan int
	LocalUDPAddr      net.UDPAddr
	StayAliveDuration time.Duration
	StayAliveTimeout  time.Time
	ListenDuration    time.Duration
}

func (p Pinger) updateStayAlive() {
	p.StayAliveTimeout = time.Now().Add(p.StayAliveDuration)
}
func (l Listener) updateStayAlive() {
	l.StayAliveTimeout = time.Now().Add(l.StayAliveDuration)
}
func (p Pinger) hasTimedOut() bool {
	return (p.StayAliveTimeout.Sub(time.Now()) <= 0)
}
func (l Listener) hasTimedOut() bool {
	return (l.StayAliveTimeout.Sub(time.Now()) <= 0)
}

func GeneralBroadcaster(s sync.WaitGroup, m chan []byte, k chan int) Pinger {
	localaddr, err := net.ResolveUDPAddr("udp", ":20008")
	remoteaddr := net.UDPAddr{IP: net.IPv4(255, 255, 255, 255), Port: 20008}
	// @todo: handle localaddr error

	b := Pinger{
		WG:                s,
		Messages:          m,
		Keepalive:         k,
		LocalUDPAddr:      *localaddr,
		RemoteUDPAddr:     remoteaddr,
		StayAliveDuration: time.Second * KEEPALIVE_TIME,
		StayAliveTimeout:  time.Now(), // Updates upon calling b.Instance()
		MessageDeadline:   time.Second * MESSAGE_DEADLINE}
	return b
}

func GeneralListener(s sync.WaitGroup, m chan []byte, k chan int) Listener {
	localaddr, err := net.ResolveUDPAddr("udp", ":20008")
	// @todo: Handle localaddr error

	l := Listener{
		WG:                s,
		Messages:          m,
		Keepalive:         k,
		LocalUDPAddr:      *localaddr,
		StayAliveDuration: time.Second * KEEPALIVE_TIME,
		StayAliveTimeout:  time.Now(), // Updates upon calling l.Instance()
		ListenDuration:    time.Second * LISTEN_DURATION}
	return l
}

// run go broadcasterInstance(...)
func (p Pinger) Instance() {
	// WaitGroup updates for Goroutine synchronization
	p.WG.Add(1)
	defer p.WG.Done()

	conn, err := net.DialUDP("udp", &p.LocalUDPAddr, &p.RemoteUDPAddr)
	if err != nil {
		// handle error
		// @todo
	}
	defer conn.Close()

	p.updateStayAlive()

	for {
		select {
		case signal := <-p.Keepalive:
			if signal == KEEPALIVE {
				p.updateStayAlive()
			} else if signal == STOP {
				// All cleanup is done with defer
				return
			}

		case msg := <-p.Messages:
			{
				conn.SetDeadline(time.Now().Add(p.MessageDeadline))
				_, err := conn.Write( /* Only for testing. Replace with msg ! ! ! ->*/ []byte(msg))
				if err != nil {
					// @todo handle error
				}
			}
		default:
			continue
		}

		if p.hasTimedOut() {
			// All cleanup is done with defer
			return
		}
	}
}

func (l Listener) Instance() {
	// WaitGroup updates for Goroutine synchronization
	l.WG.Add(1)
	defer l.WG.Done()

	conn, err := net.ListenUDP("udp", &l.LocalUDPAddr)
	if err != nil {
		// @todo handle error
	}
	defer conn.Close()

	var buf [1024]byte
	var msgBytes int = 1024

	var msgPassedToChannel = true

	l.updateStayAlive()

	for {
		// The order of the select, timeoutcheck, and listen is not arbitrary
		// If it times out while listening, it will check for keepalive
		select {
		case signal := <-l.Keepalive:
			if signal == KEEPALIVE {
				l.updateStayAlive()
			} else if signal == STOP {
				// All cleanup is done with defer
				return
			}
		default:
			continue
		}
		if l.hasTimedOut() {
			// All cleanup done with defer
			return
		}
		if msgPassedToChannel {
			conn.SetDeadline(time.Now().Add(l.ListenDuration))
			// Unsure if n is number of bytes. Unsure if addr refers to sender.
			n, addr, err := conn.ReadFromUDP(buf[0:])
			if err != nil {
				// @todo handle errror
			} else {
				msgPassedToChannel = false
			}
		} else {
			/*

				Kom så langt i dag.
				Utfordring nå:
				- Problemer med buffer / slice / arrays
					* Får ikke sendt returmeldingen. Tror dette er
					  pga. at vi ikke kan sende referanse til minneområdet? Litt usikker.
				- Har ikke testet om noe fungerer




			*/
			msg := make([]byte, 1024)
			select {
			case l.Messages <- msg:
				// Managed to send buffer, so we clear it
				buf.clear() // is it necessary?
			default:
				// @todo log error here: THe message could not send!
			}
		}
	}
}
