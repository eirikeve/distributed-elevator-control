package localip

import (
	"net"
	"strings"
	ed "../../elevdriver" //@SIM
)
var localIP string

func LocalIP() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]+":"+ed.PortNum // @SIM added '":"+ed.PortNUM"'
	}
	return localIP, nil
}
