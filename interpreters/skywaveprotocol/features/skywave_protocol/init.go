package skywave_protocol

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"skywaveprotocol/features/skywave_protocol/models"
)

func Initialize(data string) (string, error) {
	var skywaveResult models.GetReturnMessagesResult
	err := xml.Unmarshal([]byte(data), &skywaveResult)
	if err != nil {
		dataJSON, err := json.Marshal(skywaveResult)
		if err != nil {
			return "", fmt.Errorf("error marshaling ST300 data: %v", err)
		}
		return string(dataJSON), nil
	} else {
		fmt.Println("Could not identify model from data")
	}
	return "", fmt.Errorf("no valid model identified from data: %s", data)
}
