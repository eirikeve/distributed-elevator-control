package localip

import (
	et "../../elevtype"
	"net"
	"strconv"
	"strings"
)

var localIP string
var localID int32

func LocalIP() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func LocalID() (int32, error) {
	_, err := LocalIP()
	if localID == 0 {
		if localIP == "" {
			return 0, err
		} else {

			splitIP := strings.Split(localIP, ".")
			v, _ := strconv.Atoi(splitIP[len(splitIP)-1])
			localID = int32(v)
			v2, _ := strconv.Atoi(et.SystemIpPort)
			localID += int32(v2)

			if localID == 0 {
				localID = 256
			}
		}
	}
	return localID, nil
}

func LocalIPWithPort() (string, error) {
	_, err := LocalIP()
	if localIP == "" {
		return "", err
	}
	return localIP + ":" + et.SystemIpPort, nil
}
