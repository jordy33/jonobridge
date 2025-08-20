package models

import "strings"

// IdentifyModel determines which queclink model the data is from

func IdentifyModel(data string) string {
	if data == "" {
		return ""
	}

	// Split data by comma as shown in the example code
	parts := strings.Split(data, ",")
	if len(parts) < 2 {
		return ""
	}

	// Check message type first
	messageType := parts[0]
	if strings.HasPrefix(messageType, "+RESP") || strings.HasPrefix(messageType, "+BUFF") {
		// Extract device type from first 2 characters of the second part
		if len(parts[1]) >= 2 {
			deviceType := parts[1][:2]

			// Check for valid device types
			switch deviceType {
			case "30": // For model 300
				return "300"
			case "32": // For model 320
				return "320"
			case "35": // For model 350
				return "350"
			}
		}
	}

	return ""
}
