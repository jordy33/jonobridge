package helpers

// AlarmCodeMap maps alarm byte codes to human-readable descriptions
var AlarmCodeMap = map[byte]string{
	0x01: "SOS",
	0x02: "Power Cut Alarm",
	0x03: "Shock Alarm",
	0x04: "Fence In Alarm",
	0x05: "Fence Out Alarm",
	0x14: "Normal",
	// Add other alarm codes as needed
}

// AlarmEventCodeMap maps alarm byte codes to event codes
var AlarmEventCodeMap = map[byte]string{
	0x01: "1",  // SOS
	0x02: "23", // Power Cut
	0x03: "79", // Shock
	0x04: "20", // Fence In
	0x05: "21", // Fence Out
	0x14: "35", // Normal
	// Add other mappings as needed
}

// LanguageMap maps language byte codes to language descriptions
var LanguageMap = map[byte]string{
	0x01: "Chinese",
	0x02: "English",
	// Add other languages as needed
}

// EventCodeToName maps numeric event codes to their human-readable names
var EventCodeToName = map[string]string{
	"1":  "SOS",
	"20": "Fence In Alarm",
	"21": "Fence Out Alarm",
	"23": "Power Cut Alarm",
	"35": "Normal Location",
	"50": "Alarm",
	"79": "Shock Alarm",
}

// GetEventCodeFromAlarmType returns the event code for a given alarm description
func GetEventCodeFromAlarmType(alarmType string) string {
	switch alarmType {
	case "SOS":
		return "1"
	case "Power Cut Alarm":
		return "23"
	case "Shock Alarm":
		return "79"
	case "Fence In Alarm":
		return "20"
	case "Fence Out Alarm":
		return "21"
	case "Normal":
		return "35"
	default:
		return "50" // Generic alarm
	}
}

// GetAlarmTypeFromEventCode returns the alarm type for a given event code
func GetAlarmTypeFromEventCode(eventCode string) string {
	if name, exists := EventCodeToName[eventCode]; exists {
		return name
	}
	return "Unknown Alarm"
}
