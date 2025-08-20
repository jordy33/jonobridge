package models

import "strings"

// IdentifyModel determines which Suntech model the data is from
func IdentifyModel(data string) string {
	if data == "" {
		return ""
	}

	// Split data by the delimiter
	parts := strings.Split(data, ";")
	if len(parts) == 0 {
		return ""
	}

	// Check header prefix
	header := parts[0]

	if strings.HasPrefix(header, "ST300") {
		return "ST300"
	} else if strings.HasPrefix(header, "ST4300") {
		return "ST4300"
	}

	return ""
}
