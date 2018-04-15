package localip

import (
	"net"
	"strconv"
	"strings"

	et "../../elevtype"
)

/*
 * Contains functionality for abstracting a unique identification
 * from the the local system.
 */

////////////////////////////////
// Module variables
////////////////////////////////
var localIP string
var localID int32

////////////////////////////////
// Interface
////////////////////////////////

//LocalIP (.) returns the systems local ip
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

/*LocalID (.) returns a unique ID from the local system
 * based on the local ip.
 */
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
