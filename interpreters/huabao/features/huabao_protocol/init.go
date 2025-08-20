package huabao_protocol

import (
	"fmt"
)

// Initialize parses the input data string according to the Huabao protocol
func Initialize(data string) (string, error) {
	// Use the Parse function to process the data
	parsedData, err := Parse(data)
	if err != nil {
		return "", fmt.Errorf("error parsing Huabao data: %v", err)
	}

	return parsedData, nil
}
