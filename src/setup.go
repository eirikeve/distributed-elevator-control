package main

import (
	"flag"

	el "./elevlog"
	log "github.com/sirupsen/logrus"
)

var logLevel = log.DebugLevel
var logToFileInsteadOfBash = false

func parseCmdLineArgs() {

	isDebugEnvironment := flag.Bool("debug", false, "in debug, all log messages are recorded. Otherwise, only more important log messages are recorded.")
	doLog := flag.Bool("log", true, "if logging, log messages are displayed in bash or file, (where and which depending on logtobash flag and debug flag)")
	doLogToFile := flag.Bool("logtofile", false, "if true, logs to a timestamped file instead of to bash")
	flag.Parse()

	logToFileInsteadOfBash = *doLogToFile
	if *isDebugEnvironment {
		logLevel = log.DebugLevel
	}
	if !*doLog {
		logLevel = log.FatalLevel
		logToFileInsteadOfBash = false
	}
}

func setupLog() {
	el.InitLog(logLevel, logToFileInsteadOfBash)
	log.Debug("main setupLog: log initialized")
}
