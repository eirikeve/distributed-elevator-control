package main

import (
	"flag"

	def "./elevdef"
	el "./elevlog"
	log "github.com/sirupsen/logrus"
)

var logLevel = log.DebugLevel
var logToFileInsteadOfBash = false

func parseCmdLineArgs() {

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
	def.SystemIpPort = *ipPort
	def.BackupPort = *backupPort
}

func setupLog() {
	el.InitLog(logLevel, logToFileInsteadOfBash)
	log.Debug("main setupLog: log initialized")
}
