package utils

import "log"

// Debug controls verbose logging output
var Debug bool

// SetVerbose sets the verbose flag for logging
func SetVerbose(v bool) {
	Debug = v
}

// VPrint prints verbose logs if enabled
func VPrint(format string, v ...interface{}) {
	if Debug {
		log.Printf(format, v...)
	}
}
