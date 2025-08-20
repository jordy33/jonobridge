package jono

import (
	"encoding/json"
	"fmt"
	"log"
	"pinoprotocol/features/jono/models"
	"pinoprotocol/features/jono/usecases"
	"strconv"
	"strings"
)

// Debug logging flag
var verbose = false

// EnableVerboseLogging sets the verbose flag for logging
func EnableVerboseLogging(enable bool) {
	verbose = enable
	usecases.EnableVerboseLogging(enable)
	if verbose {
		log.Println("Jono parser verbose logging enabled")
	}
}

// Initialize processes input data and returns a normalized protocol string
func Initialize(rawData string) (string, error) {
	if verbose {
		//log.Printf("Jono initializing with data: %s", rawData)
		// Add debug for mileage
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(rawData), &data); err == nil {
			if mileage, exists := data["Mileage"]; exists {
				log.Printf("DEBUG Jono: Found mileage in raw data: %v (type: %T)", mileage, mileage)
			} else {
				log.Printf("DEBUG Jono: No mileage found in raw data")
				// Debug satellites info in raw data
				if sats, exists := data["NumberOfSatellites"]; exists {
					log.Printf("DEBUG Jono: Found NumberOfSatellites in raw data: %v (type: %T)",
						sats, sats)
				}
			}
		}
	}

	// Check if this is an alarm packet from the GT06 protocol
	if containsAlarmPacket(rawData) {
		return handleAlarmPacket(rawData)
	}

	// Otherwise process as a standard location packet using existing code
	parsedData, err := usecases.GetDataJono(rawData)
	if err != nil {
		return "", fmt.Errorf("error parsed data %s", err)
	}
	return parsedData, nil
}

// AlarmLocation matches the incoming JSON structure
type AlarmLocation struct {
	IMEI               string                 `json:"IMEI"`
	DateTime           string                 `json:"Datetime"`
	NumberOfSatellites int                    `json:"NumberOfSatellites"`
	PositioningStatus  string                 `json:"PositioningStatus"`
	Latitude           float64                `json:"Latitude"`
	Longitude          float64                `json:"Longitude"`
	Speed              int                    `json:"Speed"`
	Direction          int                    `json:"Direction"`
	MCC                string                 `json:"MCC"`
	MNC                string                 `json:"MNC"`
	VoltageValue       float64                `json:"VoltageValue"`
	GSMSignalStrength  int                    `json:"GSMSignalStrength"`
	BaseStationInfo    map[string]interface{} `json:"BaseStationInfo"`
}

// AlarmData matches the incoming alarm packet JSON
type AlarmData struct {
	LocationPacketModel        AlarmLocation     `json:"locationPacketModel"`
	TerminalInformationContent string            `json:"terminalInformationContent"`
	AlarmAndLanguage           map[string]string `json:"alarmAndLanguage"`
	AlarmType                  string            `json:"AlarmType"`
	EventCode                  string            `json:"EventCode"`
	Message                    string            `json:"Message"`
	GsmSignalStrength          string            `json:"gsmSignalStrength"`
}

func handleAlarmPacket(rawData string) (string, error) {
	if verbose {
		log.Printf("Processing alarm packet: %s", rawData)
	}

	var alarmData AlarmData
	if err := json.Unmarshal([]byte(rawData), &alarmData); err != nil {
		return "", fmt.Errorf("error parsing alarm data: %v", err)
	}

	// Create the Jono model
	jonoModel := &models.ParsedModel{
		ListPackets: make(map[string]models.Packet),
	}

	// Set IMEI
	imeiVal := alarmData.LocationPacketModel.IMEI
	jonoModel.IMEI = &imeiVal

	// Set message with alarm type
	message := fmt.Sprintf("Alarm event: %s", alarmData.AlarmType)
	jonoModel.Message = &message

	// Set data packets count
	packets := 1
	jonoModel.DataPackets = &packets

	// Debug satellites info
	if verbose {
		fmt.Printf("DEBUG Jono: Incoming NumberOfSatellites: %d\n",
			alarmData.LocationPacketModel.NumberOfSatellites)
	}

	// Create base packet
	packet := models.Packet{
		Altitude: 0,
		EventCode: models.EventCode{
			Code: extractEventCode(alarmData.EventCode, alarmData.AlarmType),
			Name: alarmData.AlarmType,
		},
		NumberOfSatellites: alarmData.LocationPacketModel.NumberOfSatellites,
		AnalogInputs:       &models.AnalogInputs{},
		BaseStationInfo:    &models.BaseStationInfo{},
	}

	// Calculate voltage value for AD4
	var voltageValue float64
	if alarmData.LocationPacketModel.VoltageValue > 0 {
		voltageValue = alarmData.LocationPacketModel.VoltageValue
		if verbose {
			log.Printf("Using provided voltage value: %.1f", voltageValue)
		}
	} else {
		voltageValue = 9.0 // Default value if none provided
		if verbose {
			log.Printf("Warning: Using default voltage value: %.1f", voltageValue)
		}
	}

	// Convert voltage to hex string for AD4
	voltageInt := int(voltageValue)
	voltageHex := fmt.Sprintf("%X", voltageInt) // Convert to uppercase hex without "0x" prefix
	fmt.Println("DEBUG Jono: Voltage value in hex:", voltageHex)
	// Set the AD4 value
	packet.AnalogInputs = &models.AnalogInputs{
		AD1:  nil,
		AD2:  nil,
		AD3:  nil,
		AD4:  &voltageHex, // Set AD4 to hex voltage value
		AD5:  nil,
		AD6:  nil,
		AD7:  nil,
		AD8:  nil,
		AD9:  nil,
		AD10: nil,
	}

	// Set location data
	datetime := alarmData.LocationPacketModel.DateTime
	packet.Datetime = &datetime

	lat := alarmData.LocationPacketModel.Latitude
	packet.Latitude = &lat

	lon := alarmData.LocationPacketModel.Longitude
	packet.Longitude = &lon

	speed := alarmData.LocationPacketModel.Speed
	packet.Speed = &speed

	direction := alarmData.LocationPacketModel.Direction
	packet.Direction = &direction

	status := alarmData.LocationPacketModel.PositioningStatus
	packet.PositioningStatus = &status

	// Set GSM signal strength using the mapping function
	if alarmData.GsmSignalStrength != "" {
		// Convert string GSM value to int and map it
		gsmValue := 0
		switch alarmData.GsmSignalStrength {
		case "No signal":
			gsmValue = 0
		case "Extremely weak signal":
			gsmValue = 12
		case "Very weak signal":
			gsmValue = 15
		case "Good signal":
			gsmValue = 18
		case "Strong signal":
			gsmValue = 31
		}
		gsm := mapGSMValue(gsmValue)
		packet.GSMSignalStrength = &gsm
	}

	// Set base station info
	mcc := alarmData.LocationPacketModel.MCC
	mnc := alarmData.LocationPacketModel.MNC
	packet.BaseStationInfo = &models.BaseStationInfo{
		MCC: &mcc,
		MNC: &mnc,
	}

	// Add mileage if it exists
	fmt.Printf("DEBUG Jono: Attempting to extract mileage from: %s\n", rawData)
	if mileage, exists := extractMileageFromData(rawData); exists {
		packet.Mileage = &mileage
		fmt.Printf("DEBUG Jono: Successfully extracted mileage: %d\n", mileage)
	} else {
		fmt.Printf("DEBUG Jono: Failed to extract mileage\n")
	}

	// Store the packet
	jonoModel.ListPackets["Packet1"] = packet

	// Debug final packet
	if verbose {
		if packet.Mileage != nil {
			fmt.Printf("DEBUG Jono: Final packet mileage: %d\n", *packet.Mileage)
		} else {
			fmt.Printf("DEBUG Jono: Final packet mileage is nil\n")
		}
		fmt.Printf("DEBUG Jono: Final packet NumberOfSatellites: %d\n",
			packet.NumberOfSatellites)
	}

	// Convert to JSON with error handling
	jsonData, err := jonoModel.ToJSON()
	if err != nil {
		return "", fmt.Errorf("error converting to JSON: %v", err)
	}
	return jsonData, nil
}

// Helper function to extract event code from alarm data
func extractEventCode(eventCode string, alarmType string) int {
	// First try explicit event code
	if code, err := strconv.Atoi(eventCode); err == nil {
		return code
	}

	// Otherwise map from alarm type
	switch alarmType {
	case "SOS":
		return 1
	case "Power Cut Alarm":
		return 23
	case "Shock Alarm":
		return 79
	case "Fence In Alarm":
		return 20
	case "Fence Out Alarm":
		return 21
	default:
		return 35 // Default to normal
	}
}

func containsAlarmPacket(data string) bool {
	return strings.Contains(data, "\"alarmAndLanguage\"")
}

// Helper function to map GSM values to standard scale
func mapGSMValue(value int) int {
	switch value {
	case 0:
		return 0 // No signal
	case 1:
		return 8 // Extremely weak
	case 2:
		return 16 // Very weak
	case 3:
		return 23 // Good
	case 4:
		return 31 // Strong
	default:
		return 16 // Default to very weak
	}
}

// Add new helper function to extract mileage
func extractMileageFromData(rawData string) (int, bool) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(rawData), &data); err != nil {
		return 0, false
	}

	if mileageVal, exists := data["Mileage"]; exists {
		switch v := mileageVal.(type) {
		case float64:
			return int(v), true
		case int:
			return v, true
		case string:
			if mInt, err := strconv.Atoi(v); err == nil {
				return mInt, true
			}
		}
	}
	return 0, false
}
