/*package utils

import (
	"flag"
	"fmt"
)

// Verbose controls debug output
var Verbose bool

func init() {
	flag.BoolVar(&Verbose, "v", false, "verbose output")
}

// VPrint prints a message if verbose mode is enabled
func VPrint(format string, a ...interface{}) {
	if Verbose {
		fmt.Printf(format+"\n", a...)
	}
}
*/

package utils

import (
	"flag"
	"fmt"
	"log"
)

// Verbose controls debug output
var Verbose bool

func init() {
	flag.BoolVar(&Verbose, "v", false, "verbose output")
	// Configure the standard logger to include date and time but not the prefix
	log.SetFlags(log.LstdFlags)
}

// VPrint prints a message if verbose mode is enabled
func VPrint(format string, a ...interface{}) {
	if Verbose {
		// Format the message using fmt.Sprintf first
		message := fmt.Sprintf(format, a...)
		// Then use log.Println to add the timestamp
		log.Println(message)
	}
}
