package usecases

import (
	"encoding/json"
	"fmt"
)

func GetDataJono(data string) (string, error) {

	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &rawData); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	parsedModel := map[string]interface{}{
		"IMEI":        rawData["IMEI"],
		"Message":     rawData["Message"],
		"DataPackets": getOrDefault(rawData, "DataPackets", 1),
		"ListPackets": make(map[string]interface{}),
	}

	packetFields := []string{"Altitude", "Datetime", "EventCode", "Latitude", "Longitude", "Speed"}

	if packets, ok := rawData["ListPackets"].(map[string]interface{}); ok {
		for key, packetData := range packets {
			packetMap := packetData.(map[string]interface{})
			packet := map[string]interface{}{
				"Altitude":  getOrDefault(packetMap, "Altitude", 0),
				"Datetime":  getOrDefault(packetMap, "Datetime", ""),
				"EventCode": getOrDefault(packetMap, "EventCode", nil),
				"Latitude":  getOrDefault(packetMap, "Latitude", 0.0),
				"Longitude": getOrDefault(packetMap, "Longitude", 0.0),
				"Speed":     getOrDefault(packetMap, "Speed", 0),
				"Extras":    extractExtras(packetMap, packetFields),
			}
			parsedModel["ListPackets"].(map[string]interface{})[key] = packet
		}
	} else {

		packet := map[string]interface{}{
			"Altitude":  getOrDefault(rawData, "Altitude", 0),
			"Datetime":  getOrDefault(rawData, "Datetime", ""),
			"EventCode": getOrDefault(rawData, "EventCode", nil),
			"Latitude":  getOrDefault(rawData, "Latitude", 0.0),
			"Longitude": getOrDefault(rawData, "Longitude", 0.0),
			"Speed":     getOrDefault(rawData, "Speed", 0),
			"Extras":    extractExtras(rawData, append(packetFields, "IMEI", "Message", "DataPackets")),
		}
		parsedModel["ListPackets"].(map[string]interface{})["packet_1"] = packet
	}
	// fmt.Print(parsedModel)
	processedData, err := json.Marshal(parsedModel)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(processedData), nil
}

func extractExtras(data map[string]interface{}, knownFields []string) map[string]interface{} {
	extras := make(map[string]interface{})
	for key, value := range data {
		if !contains(knownFields, key) {
			extras[key] = value
		}
	}
	return extras
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func getOrDefault(data map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if value, exists := data[key]; exists {
		return value
	}
	return defaultValue
}
