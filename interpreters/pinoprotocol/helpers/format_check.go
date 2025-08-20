package helpers

import (
	"encoding/json"
	"fmt"
)

// ValidateJonoFormat checks if the Jono protocol output has a consistent format
// This can be called from both location and alarm packet handlers to ensure format consistency
func ValidateJonoFormat(jonoOutput string, source string) bool {
	var output map[string]interface{}
	if err := json.Unmarshal([]byte(jonoOutput), &output); err != nil {
		fmt.Printf("ERROR validating %s output: %v\n", source, err)
		return false
	}

	// Check required top-level fields
	requiredFields := []string{"IMEI", "ListPackets"}
	for _, field := range requiredFields {
		if _, exists := output[field]; !exists {
			fmt.Printf("ERROR: %s output is missing required field '%s'\n", source, field)
			return false
		}
	}

	// Check ListPackets format
	if listPackets, ok := output["ListPackets"].(map[string]interface{}); ok {
		if len(listPackets) == 0 {
			fmt.Printf("ERROR: %s output contains empty ListPackets\n", source)
			return false
		}

		// Check the first packet
		var packet map[string]interface{}
		for _, p := range listPackets {
			if pMap, ok := p.(map[string]interface{}); ok {
				packet = pMap
				break
			}
		}

		if packet == nil {
			fmt.Printf("ERROR: %s output has invalid packet format\n", source)
			return false
		}

		// Check for EventCode in the packet
		if eventCode, ok := packet["EventCode"].(map[string]interface{}); ok {
			// Check that Code and Name exist in EventCode
			if _, exists := eventCode["Code"]; !exists {
				fmt.Printf("ERROR: %s output is missing EventCode.Code\n", source)
				return false
			}
			if _, exists := eventCode["Name"]; !exists {
				fmt.Printf("ERROR: %s output is missing EventCode.Name\n", source)
				return false
			}
		} else {
			fmt.Printf("ERROR: %s output is missing EventCode structure\n", source)
			return false
		}

	} else {
		fmt.Printf("ERROR: %s output has invalid ListPackets format\n", source)
		return false
	}

	return true
}

// GetEventCodeFromJonoOutput extracts the EventCode from Jono protocol output
func GetEventCodeFromJonoOutput(jonoOutput string) (int, string, error) {
	var output map[string]interface{}
	if err := json.Unmarshal([]byte(jonoOutput), &output); err != nil {
		return 0, "", err
	}

	if listPackets, ok := output["ListPackets"].(map[string]interface{}); ok {
		// Get the first packet
		for _, p := range listPackets {
			if packet, ok := p.(map[string]interface{}); ok {
				if eventCode, ok := packet["EventCode"].(map[string]interface{}); ok {
					code := 0
					name := ""

					if codeVal, ok := eventCode["Code"].(float64); ok {
						code = int(codeVal)
					}

					if nameVal, ok := eventCode["Name"].(string); ok {
						name = nameVal
					}

					return code, name, nil
				}
			}
		}
	}

	return 0, "", fmt.Errorf("could not extract EventCode from Jono output")
}
