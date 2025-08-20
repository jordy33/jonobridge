package ruptela_protocol

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"ruptelaprotocol/features/ruptela_protocol/usecases"
)

func Initialize(data string) (string, error) {

	testTrama, err := hex.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("error decoding data (required ruptela protocol): %v", err)
	}

	conversionData, _, err := usecases.Conversion(testTrama)
	if err != nil {
		return "", fmt.Errorf("error decoding data: %v", err)
	}

	if len(conversionData) == 0 {
		return "", fmt.Errorf("conversionData is empty")
	}

	firstObject := conversionData[0]

	jsonData, err := json.Marshal(firstObject)
	if err != nil {
		return "", fmt.Errorf("error converting map to JSON: %v", err)
	}

	//fmt.Println("JSON resultante:", string(jsonData))
	return string(jsonData), nil
}
