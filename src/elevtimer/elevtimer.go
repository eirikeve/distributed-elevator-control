package elevtimer

import (
	"time"

	log "github.com/sirupsen/logrus"
)

/*Start creates a timer instance which signals true on signalTimeout channel after duration time has passed.
 * timerName is just an identifier for when logging info about the timer.
 * If duration is below 0, the timeout is signaled instantly.
 * @arg signalTimeout: channel where the timer signals when it timeouts.
 * @arg duration: the duration from the function call until the timer signals timeout
 * @arg timerName: timer identifier, can be passed as "" if desired. Only used for logging purposes.
 */
func Start(signalTimeout chan bool, duration time.Duration, timerName string) {
	go timerInstance(signalTimeout, duration, timerName)
	log.WithField("timerName", timerName).Debug("elevtimer Start: Timer start")
}

/*timerInstance is an instance which signals true to signalTimeout after duration time.
 * It is created in the Start function, you don't need to call this function directly.
 * @arg signalTimeout: channel where the timer signals when it timeouts.
 * @arg duration: the duration from the function call until the timerInstance signals timeout
 * @arg timerName: timer identifier, can be passed as "" if desired. Only used for logging purposes.
 */
func timerInstance(signalTimeout chan<- bool, duration time.Duration, timerName string) {
	var trySendTimeoutCount = 0
	const trySendTimeoutLimit = 100

	if duration > time.Second*0 {
		time.Sleep(duration)
	}

	for {
		select {
		case signalTimeout <- true:
			log.WithField("timerName", timerName).Debug("elevtimer timerInstance: Timer timeout signalled. Shutting down timer instance")
			return
		default:
			if trySendTimeoutCount > trySendTimeoutLimit {
				log.WithField("timerName", timerName).Error("elevtimer timerInstance: Unable to signal. Shutting down timer instance")
				return
			} else {
				trySendTimeoutCount++
				log.WithField("timerName", timerName).Warning("elevtimer timerInstance: Unable to signal timeout, retrying in 10ms")
				time.Sleep(time.Millisecond * 10)
			}

		}
	}
}
