package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"queclinkprotocol/features/queclink_protocol"

	"github.com/MaddSystems/jonobridge/common/utils"
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Check if we have at least one argument
	if len(flag.Args()) < 1 {
		utils.VPrint("Usage: queclink [-v] <queclink_data>")
		os.Exit(1)
	}

	// Process the data from arguments
	data := strings.Join(flag.Args(), " ")
	result, err := queclink_protocol.Initialize(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing data: %v\n", err)
		os.Exit(1)
	}

	utils.VPrint(result)
}
