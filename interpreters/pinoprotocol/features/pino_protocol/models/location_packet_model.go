package models

// LocationPacketModel represents the decoded data from a location packet
type LocationPacketModel struct {
	IMEI               string                 `json:"IMEI"`
	EventCode          string                 `json:"EventCode"`
	DateTime           string                 `json:"Datetime"`
	NumberOfSatellites int                    `json:"NumberOfSatellites"`
	PositioningStatus  string                 `json:"PositioningStatus"`
	Latitude           float64                `json:"Latitude"`
	Longitude          float64                `json:"Longitude"`
	Speed              int                    `json:"Speed"`
	Course             int                    `json:"Course"`
	Direction          int                    `json:"Direction"`
	MCC                string                 `json:"MCC"`
	MNC                string                 `json:"MNC"`
	LAC                string                 `json:"LAC"`
	CellID             string                 `json:"CellID"`
	BaseStationInfo    map[string]interface{} `json:"BaseStationInfo"`
	Message            []byte                 `json:"-"`
	Extra              string                 `json:"Extra"`
	BatteryLevel       int                    `json:"BatteryLevel"`
	GSMSignalStrength  int                    `json:"GSMSignalStrength"`
	VoltageValue       float64                `json:"VoltageValue"`
}
