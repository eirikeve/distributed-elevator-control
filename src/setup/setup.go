package setup

import (
	"flag"
	"os"
	"strconv"
	"time"

	et "../elevtype"
	log "github.com/sirupsen/logrus"
)

var logLevel = log.DebugLevel
var logToFileInsteadOfBash = false

func ParseCmdLineArgs() {

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

func SetupLog() {
	initLog(logLevel, logToFileInsteadOfBash)
	log.Debug("main setupLog: log initialized")
}

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
