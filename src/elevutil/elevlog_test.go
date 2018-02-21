package elevutil

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestLog(*testing.T) {
	initLog()
	var b = 5
	var a [10]byte
	a[2] = 255

	log.WithFields(log.Fields{
		"variable b": 10,
		"variable a": 10,
	}).Debug("This should not show up. This is info.")

	log.WithFields(log.Fields{
		"variable b": b,
		"variable a": a,
	}).Info("This was logged on purpose. This is info.")

	log.WithFields(log.Fields{
		"variable b": a,
		"variable a": b,
	}).Warning("This is a warning.")
}
