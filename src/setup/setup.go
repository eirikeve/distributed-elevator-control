package setup

import (
	"flag"
	"os"
	"strconv"
	"time"

	et "../elevtype"
	px "../phoenix"
	sb "../sysbackup"
	ss "../sysstate"
	log "github.com/sirupsen/logrus"
)

/*

setup is used when initializing the system.
It is used to perform necessary setup for starting the main processes (nethandler, elevhandler)

*/

////////////////////////////////
// Module variables
////////////////////////////////

// Used to check if we managed to recover a backup of the local elevator, in CheckBackupForExistingElevator().
var setupStartTime int64

// Controls which log messages display. Set from cmd line args.
var logLevel = log.WarnLevel

// Set from cmd line args
var logToFileInsteadOfBash = false

var localElevInitialState *et.Elevator

////////////////////////////////
// Interface
////////////////////////////////

/*Setup performs all necessary setup on the system.
 * After calling setup, the system is ready to start nethandler and elevhandler
 */
func Setup() {
	setupStartTime = time.Now().Unix()
	handleCmdLineArgs()
	setupLog()
	setupPhoenix()
	setupSysState()
	systemStates := recoverBackup()
	setSystemStates(systemStates)
	log.Info("setup: Completed setup")
}

/*CheckBackupForExistingElevator checks if we managed to recover our elevator state from a backup,
 * and if so, returns it.
 * @return: nil if no backup was recovered, else the most recent backup of the local elevator
 */
func CheckBackupForExistingElevator() *et.Elevator {
	localSys := ss.GetLocalSystem()

	log.WithField("states", ss.GetLocalSystem().E).Debug("main run: Done w/ setup of sysstates")
	if setupStartTime > localSys.StartupTime {
		return &localSys.E
	} else {
		return nil
	}
}

////////////////////////////////
// Auxiliary functions
////////////////////////////////

/*handleCmdLineArgs parses command line arguments and sets the module variables accordingly.
 */
func handleCmdLineArgs() {

	isDebugEnvironment := flag.Bool("debug", false, "in debug, all log messages are recorded. Otherwise, only more important log messages are recorded.")
	doLog := flag.Bool("log", true, "if logging, log messages are displayed in bash or file, (where and which depending on logtobash flag and debug flag)")
	doLogToFile := flag.Bool("logtofile", false, "if true, logs to a timestamped file instead of to bash")
	ipPort := flag.String("port", "15657", "port nr for local ip address")
	backupPort := flag.String("backupPort", "23003", "port nr for monitoring primary")

	flag.Parse()

	logToFileInsteadOfBash = *doLogToFile

	if *isDebugEnvironment {
		logLevel = log.DebugLevel
	}
	if !*doLog {
		logLevel = log.FatalLevel
		logToFileInsteadOfBash = false
	}
	et.SystemIpPort = *ipPort
	et.BackupPort = *backupPort
}

/*setupLog initializes the log module with the module variables.
 */
func setupLog() {
	initLog(logLevel, logToFileInsteadOfBash)
	log.Debug("main setupLog: log initialized")
}

/*initLog sets log output and log level.
 * @arg level: minimum importance level of log messages to be logged
 * @arg toFile: flag on whether to log to file or not (-> bash)
 */
func initLog(level log.Level, toFile bool) {
	if toFile {
		// Set name of current log output to the current UNIX timestamp
		filename := strconv.FormatInt(time.Now().Unix(), 10) + ".elevlog"
		logFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			// Could not open file
			log.WithFields(log.Fields{
				"Error": err.Error(),
			}).Error("initLog: Could not open log output file. Defaulting to bash output.")
		} else {
			log.SetOutput(logFile)
		}
	}
	if log.PanicLevel <= level && level <= log.DebugLevel {
		log.SetLevel(level)
	} else {
		// Invalid level input
		log.SetLevel(log.InfoLevel)
		log.WithFields(log.Fields{
			"Log level": level,
		}).Warning("initLog: Invalid log level input. Defaulting to log.InfoLevel")
	}

}

/*setupPhoenix starts the phoenix service, which restarts our system upon crash
 */
func setupPhoenix() {
	px.StartSurveillanceOfPrimary()
}

/*setupSysState initializes our sysstate module, which stores the state of the local & remote systems (including orders).
 */
func setupSysState() {
	ss.InitSysState()
}

/*recoverBackup returns the most recent backup of the sysstate module
 */
func recoverBackup() []et.ElevState {
	systemStates, _ := sb.Recover(time.Now().Add(et.BackupRecoverInterval))
	return systemStates
}

/*setSystemStates sets sysstate's systems to systemStates
 */
func setSystemStates(systemStates []et.ElevState) {
	log.WithField("states", systemStates).Debug("main run: Setup sysstates")
	ss.SetSystemsStates(systemStates)

}
