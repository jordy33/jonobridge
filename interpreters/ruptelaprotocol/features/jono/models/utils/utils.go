package utils

import "log"

var verbose *bool

// SetVerbose sets the verbose flag for logging
func SetVerbose(v *bool) {
	verbose = v
}

// VPrint prints verbose logs if enabled
func VPrint(format string, v ...interface{}) {
	if verbose != nil && *verbose {
		log.Printf(format, v...)
	}
}
