package models

import "encoding/json"

// AlarmPacketModel represents the decoded data from an alarm packet
type AlarmPacketModel struct {
	LocationPacketModel        *LocationPacketModel `json:"locationPacketModel"`
	TerminalInformationContent string               `json:"terminalInformationContent"`
	VoltageLevel               string               `json:"voltageLevel"`
	GSMSignalStrength          string               `json:"gsmSignalStrength"`
	AlarmAndLanguage           map[string]string    `json:"alarmAndLanguage"`
}

// ToJSON converts the alarm packet model to a JSON string
func (a AlarmPacketModel) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
