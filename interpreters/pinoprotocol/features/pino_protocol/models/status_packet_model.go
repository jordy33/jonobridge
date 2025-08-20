package models

import "encoding/json"

// StatusPacketModel represents the decoded data from a status/heartbeat packet
type StatusPacketModel struct {
	IMEI                      string            `json:"IMEI"`
	TerminalInformationByte   byte              `json:"TerminalInformationByte"`
	TerminalInformationString string            `json:"TerminalInformationString"`
	VoltageLevelByte          byte              `json:"VoltageLevelByte"`
	VoltageLevelString        string            `json:"VoltageLevelString"`
	VoltageValue              float64           `json:"VoltageValue"`
	GSMSignalStrengthByte     byte              `json:"GSMSignalStrengthByte"`
	GSMSignalStrengthString   string            `json:"GSMSignalStrengthString"`
	AlarmAndLanguage          map[string]string `json:"AlarmAndLanguage"`
	Message                   []byte            `json:"-"`
}

// ToJSON converts the status packet model to a JSON string
func (s StatusPacketModel) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
