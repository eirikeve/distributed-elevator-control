package elevdriver

import "sync"
import "../elevlog"
import log "github.com/sirupsen/logrus"
import "testing"

/*
TestLog tests basic log functionality, and shows how to use the log system
*/
func TestDriver(*testing.T) {
	elevlog.initLog(log.DebugLevel, true)
	log.WithField("a", a)
	wg = &sync.WaitGroup
	Driver()

}
