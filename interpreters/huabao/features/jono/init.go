package jono

import (
	"encoding/json"
	"fmt"
	"huabaoprotocol/features/jono/usecases"
)

func Initialize(data string) (string, error) {
	// First ensure we have valid JSON from the Huabao protocol parser
	var parsedData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &parsedData); err != nil {
		return "", fmt.Errorf("error parsing input JSON: %v", err)
	}

	// Re-encode to ensure we have well-formed JSON
	prettyData, err := json.Marshal(parsedData)
	if err != nil {
		return "", fmt.Errorf("error re-marshaling JSON: %v", err)
	}

	// Now convert to Jono protocol
	jonoData, err := usecases.GetDataJono(string(prettyData))
	if err != nil {
		return "", fmt.Errorf("error converting to Jono protocol: %v", err)
	}

	return jonoData, nil
}
