package recover

import (
	"testing"
)

/* Test created to check if the secondary is able to monitor the primary
 * If Primary crashes, secondary will become after 3 seconds and create a new
 * secondary.
 * To run test spawnbackup() must be modified to execute test and setup
 */
func TestSurveillanceProcess(*testing.T) {

	runSurveillanceProcess()
	for {
		//Do nothing
	}
}
