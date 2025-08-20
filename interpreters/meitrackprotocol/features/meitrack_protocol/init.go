package meitrack_protocol

import (
	"encoding/json"
	"fmt"
	"meitrackprotocol/features/meitrack_protocol/models"
	"meitrackprotocol/features/meitrack_protocol/usecases"
	"strconv"
	"strings"
)

func Initialize(data string) (string, error) {
	fields, err := models.ParseGeneralFields(data)

	if err != nil {
		return "", fmt.Errorf("error: general fields - %v - %s", err, data)
	}

	if fields.CommandType == models.CommandAAA {
		aaaFields := models.AAAModel{GeneralModel: fields}
		data, err := usecases.ParseAAAFields(&aaaFields)
		if err != nil {
			return "", fmt.Errorf("error: aaa - %v - data %s", err, data)
		}

		// Extract raw analog inputs from message to preserve formatting
		rawAnalogInputs := ""
		parts := strings.Split(fields.Rest, ",")

		// Search for analog inputs pattern in all parts
		for _, part := range parts {
			if strings.Contains(part, "|") && strings.Count(part, "|") >= 4 {
				rawAnalogInputs = part
				break
			}
		}

		// Add raw analog inputs to the JSON data
		var aaaData map[string]interface{}
		if err := json.Unmarshal([]byte(data), &aaaData); err == nil {
			aaaData["RawAnalogInputs"] = rawAnalogInputs

			// Make sure Mileage is properly extracted from AAA protocol from position 14
			if len(parts) > 14 {
				if mileage, err := strconv.Atoi(parts[14]); err == nil {
					aaaData["Mileage"] = mileage
				}
			}
		}

		// Re-marshal with the added data
		if enrichedData, err := json.Marshal(aaaData); err == nil {
			data = string(enrichedData)
		}

		return data, nil
	} else if fields.CommandType == models.CommandCCE || fields.CommandType == models.CommandCFF {
		// Handle both CCE and CFF and E01 protocols the same way
		cceFields := models.CCEModel{GeneralModel: fields}
		dataParsed, err := usecases.ParseCCEFields(&cceFields)
		if err != nil {
			return "", fmt.Errorf("error: %s - %v - data %s", fields.CommandType, err, data)
		}
		return dataParsed, nil
	} else if fields.CommandType == models.CommandCCC {
		cccFields := models.CCCModel{GeneralModel: fields}
		dataParsed, err := usecases.ParseCCCFields(&cccFields)
		if err != nil {
			return "", fmt.Errorf("error: ccc - %v - data %s", err, data)
		}

		return dataParsed, nil
	} else {
		jsonData, err := json.Marshal(fields)
		if err != nil {
			return "", fmt.Errorf("Error converting to JSON:", err)
		}
		return string(jsonData), nil
	}
}
