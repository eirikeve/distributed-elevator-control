package elevtimer

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

/*
 * The timer module contains functionality for creating timer instances with a given duration,
 * The instance signalizes a timeout after the given duration has passed.   
 */




////////////////////////////////
// Module variables
////////////////////////////////

// Map for updating the duration of timers
var timers = make(map[string]chan time.Duration)

// Mutex for thread safety. Used due to accessing timers, which can cause unexpected behaviour
var lock sync.Mutex


////////////////////////////////
// Interface
////////////////////////////////


/*Start creates a timer instance which signals true on signalTimeout channel after duration time has passed.
 * timerName is just an identifier for when logging info about the timer.
 * If duration is below 0, the timeout is signaled instantly.
 * @arg signalTimeout: channel where the timer signals when it timeouts.
 * @arg duration: the duration from the function call until the timer signals timeout
 * @arg timerName: timer identifier, can be passed as "" if desired. Only used for logging purposes.
 */
func Start(timerName string, duration time.Duration, signalTimeout chan bool) {
	lock.Lock()
	defer lock.Unlock()
	_, exists := timers[timerName]
	if exists {
		log.WithFields(log.Fields{"timerName": timerName, "duration": duration}).Warning("elevtimer Start: Called start on existing timer, updating its duration instead")

	}
	// Make buffered durationUpdateChan, and link it to the timerName
	timers[timerName] = make(chan time.Duration, 2)
	go timerInstance(timerName, duration, signalTimeout, timers[timerName])
	log.WithFields(log.Fields{"timerName": timerName, "duration": duration}).Debug("elevtimer Start: Timer start")

}



////////////////////////////////
// Auxiliary
////////////////////////////////


/*timerInstance is an instance which signals true to signalTimeout after duration time.
 * It is created in the Start function; you don't need to call this function directly.
 * @arg signalTimeout: channel where the timer signals when it timeouts.
 * @arg duration: the duration from the function call until the timerInstance signals timeout
 * @arg timerName: timer identifier
 * @arg durationUpdateChan: used to send new durations to the timerInstance, when calling Update or Stop
 */
func timerInstance(timerName string, duration time.Duration, signalTimeout chan<- bool, durationUpdateChan <-chan time.Duration) {

	defer removeTimerInstance(timerName)

	const trySendTimeoutLimit = 100
	var trySendTimeoutCount = 0
	var startTime = time.Now()

	// Counting until duration has passed. Listen for updates to duration.
	for time.Now().Sub(startTime) < duration {
		select {
		case newDuration := <-durationUpdateChan:
			if newDuration > time.Second*0 {
				startTime = time.Now()
				duration = newDuration
				log.WithField("timerName", timerName).Debug("elevtimer timerInstance: Timer duration updated")
			} else {
				log.WithField("timerName", timerName).Debug("elevtimer timerInstance: Timer stopped, exiting")
				return
			}

		default:
			time.Sleep(time.Millisecond * 50)
		}
	}

	// Timeout! Send message on signalTimeout channel
	for {
		select {
		// We try to send 101 times, if unsuccessful, exit the instance and issue Error
		case signalTimeout <- true:
			log.WithField("timerName", timerName).Debug("elevtimer timerInstance: Timer timeout signalled. Shutting down timer instance")
			return
		default:
			if trySendTimeoutCount > trySendTimeoutLimit {
				log.WithField("timerName", timerName).Error("elevtimer timerInstance: Unable to signal. Shutting down timer instance")
				return
			} else {
				trySendTimeoutCount++
				// Only print the error 5 times, as to not clog up the log.
				if (trySendTimeoutCount % 20) == 0 {
					log.WithField("timerName", timerName).Warning("elevtimer timerInstance: Unable to signal timeout, retrying")
				}

				time.Sleep(time.Millisecond * 10)
			}

		}
	}
}

/*removeTimerInstance deletes a timerName and its corresponding durationUpdateChan from the timers map
 * @arg timerName: identity of timer to be removed
 */
func removeTimerInstance(timerName string) {
	lock.Lock()
	defer lock.Unlock()
	_, exists := timers[timerName]
	if exists {
		delete(timers, timerName)
	} else {
		log.WithField("timerName", timerName).Error("elevtimer removeTimer: Cannot find timer")
	}
}
