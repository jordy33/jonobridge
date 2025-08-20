package usecases

import (
	"encoding/json"
	"fmt"
	"skywaveprotocol/features/jono/models"
	"strconv"
	"strings"
)

// EventCodes maps event codes to their names
var EventCodes = map[int]string{
	1:  "SOS Pressed",
	2:  "Input 2 Active",
	3:  "Input 3 Active",
	4:  "Input 4 Active",
	5:  "Input 5 Active",
	9:  "Input 1 Inactive",
	10: "Input 2 Inactive",
	11: "Input 3 Inactive",
	12: "Input 4 Inactive",
	13: "Input 5 Inactive",
	17: "Low Battery",
	18: "Low External Battery",
	19: "Speeding",
	20: "Enter Geo-fence",
	21: "Exit Geo-fence",
	22: "External Battery On",
	23: "External Battery Cut",
	24: "GPS Signal Lost",
	25: "GPS Signal Recovery",
	26: "Enter Sleep",
	27: "Exit Sleep",
	28: "GPS Antenna Cut",
	29: "Device Reboot",
	31: "Heartbeat",
	32: "Cornering",
	33: "Track By Distance",
	34: "Reply Current (Passive)",
	35: "Track By Time Interval",
	36: "Tow",
	65: "Press Input 1 (SOS) to Call",
	66: "Press Input 2 to Call",
	67: "Press Input 3 to Call",
	68: "Press Input 4 to Call",
	69: "Press Input 5 to Call",
	70: "Reject Incoming Call",
	71: "Get Location by Call",
	72: "Auto Answer Incoming Call",
}

func GetDataJono(data string) (string, error) {
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &rawData); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Check for raw AAA data with analog inputs
	rawAnalogInputs, _ := rawData["RawAnalogInputs"].(string)

	// Check if this is AAA protocol data by examining Message field
	if message, exists := rawData["Message"].(string); exists && strings.Contains(message, "AAA") {
		parsedModel := models.ParsedModel{
			IMEI:        getStringPointer(rawData, "IMEI"),
			Message:     getStringPointer(rawData, "Message"),
			DataPackets: getIntPointer(rawData, "DataPackets"),
			ListPackets: make(map[string]models.Packet),
		}

		packet := createPacket(rawData)

		// Find analog inputs either in the message or directly provided
		analogInputs := findAAAnalogInputs(message, rawAnalogInputs)
		if analogInputs != nil {
			packet.AnalogInputs = analogInputs
		}

		// Find BaseStationInfo from AAA message
		baseStationInfo := findAAABaseStationInfo(message)
		if baseStationInfo != nil {
			packet.BaseStationInfo = baseStationInfo
		}

		// Find and parse IoPortStatus from AAA message
		ioPortStatus := findAAAIoPortStatus(message)
		if ioPortStatus != nil {
			packet.IoPortStatus = ioPortStatus
		}

		// Extract GSM signal strength from AAA message
		gsmSignalStrength := findAAAGSMSignalStrength(message)
		if gsmSignalStrength != nil {
			packet.GSMSignalStrength = gsmSignalStrength
		}

		// Extract Mileage from AAA message
		mileage := findAAAMileage(message)
		if mileage != nil {
			packet.Mileage = mileage
		}

		parsedModel.ListPackets["Packet1"] = packet

		// Convert to JSON and return
		jsonString, err := parsedModel.ToPrettyJSON()
		if err != nil {
			return "", fmt.Errorf("failed to convert to JSON: %w", err)
		}

		return jsonString, nil
	}

	// For non-AAA protocol data
	parsedModel := models.ParsedModel{
		IMEI:        getStringPointer(rawData, "IMEI"),
		Message:     getStringPointer(rawData, "Message"),
		DataPackets: getIntPointer(rawData, "DataPackets"),
		ListPackets: make(map[string]models.Packet),
	}

	// 📌 Verificar si "ListPackets" existe
	if packets, ok := rawData["ListPackets"].(map[string]interface{}); ok {
		// ✅ Si existe, procesamos cada paquete normalmente
		for key, packetData := range packets {
			packetMap, ok := packetData.(map[string]interface{})
			if !ok {
				continue
			}

			packet := createPacket(packetMap)
			parsedModel.ListPackets[key] = packet
		}
	} else {
		// ❌ Si no existe "ListPackets", creamos "ListPackets" con un solo paquete "Packet1"
		packet := createPacket(rawData) // Usamos directamente rawData
		parsedModel.ListPackets["Packet1"] = packet
	}

	// 📌 Convertir a JSON
	jsonString, err := parsedModel.ToPrettyJSON()
	if err != nil {
		return "", fmt.Errorf("failed to convert to JSON: %w", err)
	}

	return jsonString, nil
}

// Find analog inputs in an AAA message using the most reliable method available
func findAAAnalogInputs(message, rawAnalogInputs string) *models.AnalogInputs {
	// First try to find a part with the format "xxxx|yyyy|zzzz|aaaa|bbbb"
	msgParts := strings.Split(message, ",")
	for _, part := range msgParts {
		if strings.Contains(part, "|") && strings.Count(part, "|") >= 4 {
			analogParts := strings.Split(part, "|")
			if len(analogParts) >= 5 {
				return &models.AnalogInputs{
					AD1: formatAAAAnalogValue(analogParts[0]), // Updated to format as voltage
					AD2: formatAAAAnalogValue(analogParts[1]), // Updated to format as voltage
					AD3: formatAAAAnalogValue(analogParts[2]), // Updated to format as voltage
					AD4: formatAAAAnalogValue(analogParts[3]), // Updated to format as voltage
					AD5: formatAAAAnalogValue(analogParts[4]), // Updated to format as voltage
				}
			}
		}
	}

	// Fallback to raw analog inputs if provided
	if rawAnalogInputs != "" {
		analogParts := strings.Split(rawAnalogInputs, "|")
		if len(analogParts) >= 5 {
			return &models.AnalogInputs{
				AD1: formatAAAAnalogValue(analogParts[0]), // Updated to format as voltage
				AD2: formatAAAAnalogValue(analogParts[1]), // Updated to format as voltage
				AD3: formatAAAAnalogValue(analogParts[2]), // Updated to format as voltage
				AD4: formatAAAAnalogValue(analogParts[3]), // Updated to format as voltage
				AD5: formatAAAAnalogValue(analogParts[4]), // Updated to format as voltage
			}
		}
	}

	return nil
}

// Format analog values for AAA protocol (divide by 100 for voltage)
func formatAAAAnalogValue(value string) *string {
	// First try to parse the value as a hex number since some values might be in hex format
	if intVal, err := strconv.ParseInt(value, 16, 64); err == nil {
		// Convert to decimal and divide by 100 to get voltage
		voltage := fmt.Sprintf("%.2f", float64(intVal)/100.0)
		return &voltage
	}

	// If hex parsing fails, try to parse as a decimal integer
	if intVal, err := strconv.Atoi(value); err == nil {
		// Divide by 100 to get voltage
		voltage := fmt.Sprintf("%.2f", float64(intVal)/100.0)
		return &voltage
	}

	// If both parsing methods fail, just return the original value
	return &value
}

// Find BaseStationInfo in an AAA message (MCC|MNC|LAC|CellID format)
func findAAABaseStationInfo(message string) *models.BaseStationInfo {
	msgParts := strings.Split(message, ",")

	// In AAA protocol, BaseStationInfo is typically at position 14 (0-indexed)
	// But we'll search for it to be more robust
	for _, part := range msgParts {
		if strings.Contains(part, "|") && strings.Count(part, "|") >= 3 {
			// Skip parts that look like analog inputs (typically have 4+ pipe separators)
			if strings.Count(part, "|") >= 4 && len(part) >= 10 {
				continue
			}

			// This might be the base station info
			bsParts := strings.Split(part, "|")
			if len(bsParts) >= 4 {
				// Check if these look like BaseStationInfo values
				if isBaseStationInfo(bsParts) {
					// Return MCC and MNC directly without hex-to-decimal conversion
					// They are already decimal in the AAA protocol
					mcc := &bsParts[0]
					mnc := &bsParts[1]

					// Convert LAC and CellID from hex to decimal
					var lac, cellID *string
					if len(bsParts[2]) > 0 {
						lac = getStringPointerFromValue(bsParts[2])
					}
					if len(bsParts[3]) > 0 {
						cellID = getStringPointerFromValue(bsParts[3])
					}

					return &models.BaseStationInfo{
						MCC:    mcc,
						MNC:    mnc,
						LAC:    lac,
						CellID: cellID,
					}
				}
			}
		}
	}

	return nil
}

// Check if parts look like BaseStationInfo data
func isBaseStationInfo(parts []string) bool {
	// If first part is a numeric MCC (usually 3 digits like "334")
	if len(parts[0]) >= 2 && len(parts[0]) <= 4 {
		// And second part is a numeric MNC (usually 2-3 digits like "020")
		if len(parts[1]) >= 2 && len(parts[1]) <= 3 {
			return true
		}
	}
	return false
}

// Find and parse IoPortStatus in AAA message (typically after the BaseStationInfo)
func findAAAIoPortStatus(message string) *models.IoPortsStatus {
	msgParts := strings.Split(message, ",")

	// Look for the IoPortStatus in the message
	// It's typically found at position 15 (0-indexed), but we'll search for it
	// to be more robust
	var ioPortStatusHex string

	// First try position 15 if it's available
	if len(msgParts) > 15 {
		candidateValue := msgParts[15]
		if isHexIoPortStatus(candidateValue) {
			ioPortStatusHex = candidateValue
		}
	}

	// If not found at position 15, search for it in other positions
	if ioPortStatusHex == "" {
		for _, part := range msgParts { // Removed unused 'i' variable
			if isHexIoPortStatus(part) && !strings.Contains(part, "|") {
				ioPortStatusHex = part
				break
			}
		}
	}

	// If still not found, return nil
	if ioPortStatusHex == "" {
		return nil
	}

	// Parse the hex IoPortStatus value to get input and output port statuses
	return parseIoPortStatus(ioPortStatusHex)
}

// Check if a string looks like a valid IoPortStatus hex value
func isHexIoPortStatus(s string) bool {
	// IoPortStatus is typically a 4-digit hex value like "0421"
	if len(s) == 4 && isValidHexString(s) {
		return true
	}
	return false
}

// Check if a string is a valid hex value
func isValidHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// Parse IoPortStatus hex value into IoPortsStatus struct
func parseIoPortStatus(hexValue string) *models.IoPortsStatus {
	// Convert hex to integer
	portValue, err := strconv.ParseInt(hexValue, 16, 32)
	if err != nil {
		return nil
	}

	// Extract output ports (bits 0-7, LSB)
	outputPorts := int(portValue & 0xFF)

	// Map each bit to a port status (0 or 1)
	return &models.IoPortsStatus{
		Port1: (outputPorts >> 0) & 1, // Output port 1 (bit 0)
		Port2: (outputPorts >> 1) & 1, // Output port 2 (bit 1)
		Port3: (outputPorts >> 2) & 1, // Output port 3 (bit 2)
		Port4: (outputPorts >> 3) & 1, // Output port 4 (bit 3)
		Port5: (outputPorts >> 4) & 1, // Output port 5 (bit 4)
		Port6: (outputPorts >> 5) & 1, // Output port 6 (bit 5)
		Port7: (outputPorts >> 6) & 1, // Output port 7 (bit 6)
		Port8: (outputPorts >> 7) & 1, // Output port 8 (bit 7)
	}
}

// Find GSM signal strength in an AAA message (after number of satellites)
func findAAAGSMSignalStrength(message string) *int {
	msgParts := strings.Split(message, ",")

	// Analyzing the message structure:
	// $$A146,864352045580768,AAA,1,19.611106,-99.028335,250305225954,A,9,12,98,76...
	// Index:  0        1      2  3     4         5         6       7 8  9 10 11...
	// The GSM signal strength is at position 9 (0-indexed), not 6
	if len(msgParts) > 9 {
		if gsmStrength, err := strconv.Atoi(msgParts[9]); err == nil {
			return &gsmStrength
		}
	}

	return nil
}

// Find AAA message Mileage value (at position 14 - corrected from 11)
func findAAAMileage(message string) *int {
	msgParts := strings.Split(message, ",")

	// In AAA protocol, Mileage is at position 14 (0-indexed), not 11
	// According to documentation and matching the example: Mileage: 19655620
	if len(msgParts) > 14 {
		if mileage, err := strconv.Atoi(msgParts[14]); err == nil {
			return &mileage
		}
	}

	return nil
}

// Process AAA protocol data into Jono format
func processAAAData(rawData map[string]interface{}) (string, error) {
	message, _ := rawData["Message"].(string)

	// Parse analog inputs from the AAA message
	analogInputs := extractAAAAnalogInputs(message)

	// Create the ParsedModel
	parsedModel := models.ParsedModel{
		IMEI:        getStringPointer(rawData, "IMEI"),
		Message:     getStringPointer(rawData, "Message"),
		DataPackets: getIntPointer(rawData, "DataPackets"),
		ListPackets: make(map[string]models.Packet),
	}

	// Create a packet with the AAA data
	packet := createPacket(rawData)

	// Set the analog inputs we extracted
	if analogInputs != nil {
		packet.AnalogInputs = analogInputs
	}

	// Add the packet to the model
	parsedModel.ListPackets["Packet1"] = packet

	// Convert to JSON
	jsonString, err := parsedModel.ToPrettyJSON()
	if err != nil {
		return "", fmt.Errorf("failed to convert to JSON: %w", err)
	}

	return jsonString, nil
}

// Extract analog inputs from AAA message format
func extractAAAAnalogInputs(message string) *models.AnalogInputs {
	parts := strings.Split(message, ",")

	// AAA protocol has analog inputs at position 17 (0-indexed)
	if len(parts) < 18 {
		return nil
	}

	analogParts := strings.Split(parts[17], "|")
	if len(analogParts) < 5 {
		return nil
	}

	return &models.AnalogInputs{
		AD1: formatAAAAnalogValue(analogParts[0]), // Updated to format as voltage
		AD2: formatAAAAnalogValue(analogParts[1]), // Updated to format as voltage
		AD3: formatAAAAnalogValue(analogParts[2]), // Updated to format as voltage
		AD4: formatAAAAnalogValue(analogParts[3]), // Updated to format as voltage
		AD5: formatAAAAnalogValue(analogParts[4]), // Updated to format as voltage
		// Rest are nil as they might not be present in AAA format
	}
}

// Helper function to get string pointer from a value, converting hex to decimal if needed
func getStringPointerFromValue(value string) *string {
	// Try to parse as hex and convert to decimal - if it fails, use original value
	if decimalValue, err := strconv.ParseInt(value, 16, 64); err == nil {
		decimal := strconv.FormatInt(decimalValue, 10)
		return &decimal
	}
	return &value
}

// 📌 Función auxiliar para crear un `Packet`
func createPacket(packetMap map[string]interface{}) models.Packet {
	return models.Packet{
		Altitude:                     getIntValueOrDefault(packetMap, "Altitude", 0),
		Datetime:                     getStringPointer(packetMap, "Datetime"),
		EventCode:                    models.EventCode{Code: getCodePointer(packetMap, "EventCode"), Name: getNameCode(packetMap, "EventName")},
		Latitude:                     getFloatPointer(packetMap, "Latitude"),
		Longitude:                    getFloatPointer(packetMap, "Longitude"),
		Speed:                        getIntPointer(packetMap, "Speed"),
		RunTime:                      getIntPointer(packetMap, "RunTime"),
		Mileage:                      getIntPointer(packetMap, "Mileage"), // Make sure to extract Mileage field
		Direction:                    getIntPointer(packetMap, "Direction"),
		HDOP:                         getFloatPointer(packetMap, "HDOP"), // First try with standardized name
		PositioningStatus:            getStringPointer(packetMap, "PositioningStatus"),
		NumberOfSatellites:           getIntValueOrDefault(packetMap, "NumberOfSatellites", 0),
		GSMSignalStrength:            getIntPointer(packetMap, "GsmSignalStrength"),
		IoPortStatus:                 extractIoPortsStatus(packetMap),
		AnalogInputs:                 extractAnalogInputs(packetMap),
		BaseStationInfo:              extractBaseStationInfo(packetMap),
		OutputPortStatus:             extractOutputPortStatus(packetMap),
		InputPortStatus:              extractInputPortStatus(packetMap),
		SystemFlag:                   extractSystemFlag(packetMap),
		TemperatureSensor:            extractTemperatureSensor(packetMap),
		CameraStatus:                 extractCameraStatus(packetMap),
		CurrentNetworkInfo:           extractCurrentNetworkInfo(packetMap),
		FatigueDrivingInformation:    extractFatigueDrivingInformation(packetMap),
		AdditionalAlertInfoADASDMS:   extractAdditionalAlertInfoADASDMS(packetMap),
		BluetoothBeaconA:             extractBluetoothBeacon(packetMap, "BluetoothBeaconA"),
		BluetoothBeaconB:             extractBluetoothBeacon(packetMap, "BluetoothBeaconB"),
		TemperatureAndHumiditySensor: extractTemperatureAndHumidity(packetMap),
	}
}

// Add a new function to properly extract HDOP value, handling both field naming conventions
func getHDOPPointer(data map[string]interface{}) *float64 {
	// Try with standardized name first
	if value, exists := data["HDOP"]; exists {
		if floatValue, ok := value.(float64); ok {
			return &floatValue
		}
	}

	// Fall back to non-standardized name
	if value, exists := data["Hdop"]; exists {
		if floatValue, ok := value.(float64); ok {
			return &floatValue
		}
	}

	return nil
}

// 📌 Función para obtener un puntero a un string
func getStringPointer(data map[string]interface{}, key string) *string {
	if value, exists := data[key]; exists {
		strValue := fmt.Sprintf("%v", value)
		return &strValue
	}
	return nil
}
func getPositioningStatus(data map[string]interface{}, key string) string {
	if value, exists := data[key]; exists {
		strValue := ""
		if value == 0 || value == "V" || value == false {
			strValue = "V"
			return strValue
		} else if value == 1 || value == "A" || value == true {
			strValue = "A"
			return strValue
		} else {

			strValue = fmt.Sprintf("%v", value)
			return strValue
		}
	}
	return "L"
}

func getNameCode(data map[string]interface{}, key string) string {
	// First check if there's an EventName already in the data
	if value, exists := data[key]; exists {
		return fmt.Sprintf("%v", value)
	}

	// If no EventName, try to get the code and look it up in our map
	if codeValue, exists := data["EventCode"]; exists {
		// Debug what type of object we're getting
		fmt.Printf("DEBUG: EventCode value type: %T, value: %v\n", codeValue, codeValue)

		// Handle different possible types for EventCode
		switch cv := codeValue.(type) {
		case float64:
			code := int(cv)
			fmt.Printf("DEBUG: Looking up code %d in EventCodes map\n", code)
			if name, found := EventCodes[code]; found {
				return name
			} else {
				fmt.Printf("DEBUG: Code %d not found in EventCodes map\n", code)
			}
		case int:
			fmt.Printf("DEBUG: Looking up code %d in EventCodes map\n", cv)
			if name, found := EventCodes[cv]; found {
				return name
			} else {
				fmt.Printf("DEBUG: Code %d not found in EventCodes map\n", cv)
			}
		case map[string]interface{}:
			// Handle case where EventCode is a nested object with "Code" field
			fmt.Printf("DEBUG: EventCode is a map: %v\n", cv)
			if codeField, ok := cv["Code"]; ok {
				fmt.Printf("DEBUG: Found Code field: %T, %v\n", codeField, codeField)
				switch cf := codeField.(type) {
				case float64:
					code := int(cf)
					fmt.Printf("DEBUG: Looking up code %d in EventCodes map\n", code)
					if name, found := EventCodes[code]; found {
						return name
					} else {
						fmt.Printf("DEBUG: Code %d not found in EventCodes map\n", code)
					}
				case int:
					fmt.Printf("DEBUG: Looking up code %d in EventCodes map\n", cf)
					if name, found := EventCodes[cf]; found {
						return name
					} else {
						fmt.Printf("DEBUG: Code %d not found in EventCodes map\n", cf)
					}
				case string:
					// Handle case where Code is a string that needs to be converted to int
					if codeInt, err := strconv.Atoi(cf); err == nil {
						fmt.Printf("DEBUG: Looking up code %d in EventCodes map\n", codeInt)
						if name, found := EventCodes[codeInt]; found {
							return name
						} else {
							fmt.Printf("DEBUG: Code %d not found in EventCodes map\n", codeInt)
						}
					} else {
						fmt.Printf("DEBUG: Failed to convert Code string %s to int: %v\n", cf, err)
					}
				default:
					fmt.Printf("DEBUG: Code field has unhandled type: %T\n", codeField)
				}
			} else {
				fmt.Printf("DEBUG: Code field not found in EventCode map\n")
				// If the code field isn't found, check for "code" (lowercase)
				if codeField, ok := cv["code"]; ok {
					fmt.Printf("DEBUG: Found code field (lowercase): %T, %v\n", codeField, codeField)
					switch cf := codeField.(type) {
					case float64:
						code := int(cf)
						fmt.Printf("DEBUG: Looking up code %d in EventCodes map\n", code)
						if name, found := EventCodes[code]; found {
							return name
						} else {
							fmt.Printf("DEBUG: Code %d not found in EventCodes map\n", code)
						}
					case int:
						fmt.Printf("DEBUG: Looking up code %d in EventCodes map\n", cf)
						if name, found := EventCodes[cf]; found {
							return name
						} else {
							fmt.Printf("DEBUG: Code %d not found in EventCodes map\n", cf)
						}
					}
				}
			}
		default:
			fmt.Printf("DEBUG: EventCode has unhandled type: %T\n", codeValue)
		}
	}

	return "Unknown Event"
}

// 📌 Función para obtener un puntero a un int
func getIntPointer(data map[string]interface{}, key string) *int {
	if value, exists := data[key]; exists {
		if intValue, ok := value.(float64); ok {
			intVal := int(intValue)
			return &intVal
		}
	}
	return nil
}

// 📌 Función para obtener un puntero a un int
func getCodePointer(data map[string]interface{}, key string) int {
	// First check if the EventCode exists in the map
	if value, exists := data[key]; exists {
		// Handle different possible types for EventCode
		switch cv := value.(type) {
		case float64:
			return int(cv)
		case int:
			return cv
		case map[string]interface{}:
			// Handle map with "Code" field (CCE/CCF protocols often return this format)
			if codeField, ok := cv["Code"]; ok {
				switch cf := codeField.(type) {
				case float64:
					return int(cf)
				case int:
					return cf
				case string:
					if codeInt, err := strconv.Atoi(cf); err == nil {
						return codeInt
					}
				}
			}

			// Try lowercase "code" as well
			if codeField, ok := cv["code"]; ok {
				switch cf := codeField.(type) {
				case float64:
					return int(cf)
				case int:
					return cf
				case string:
					if codeInt, err := strconv.Atoi(cf); err == nil {
						return codeInt
					}
				}
			}
		}
	}
	// Unknown event code if nothing else matches
	return 9999
}

// 📌 Función para obtener un puntero a un float64
func getFloatPointer(data map[string]interface{}, key string) *float64 {
	if value, exists := data[key]; exists {
		if floatValue, ok := value.(float64); ok {
			return &floatValue
		}
	}
	return nil
}

// 📌 Función para extraer AnalogInputs
func extractAnalogInputs(packetMap map[string]interface{}) *models.AnalogInputs {
	// Check if we have analog inputs in raw AAA format
	if analogString, ok := packetMap["AnalogInputs"].(string); ok {
		parts := strings.Split(analogString, "|")
		if len(parts) >= 5 {
			return &models.AnalogInputs{
				AD1: formatAAAAnalogValue(parts[0]), // Updated to format as voltage
				AD2: formatAAAAnalogValue(parts[1]), // Updated to format as voltage
				AD3: formatAAAAnalogValue(parts[2]), // Updated to format as voltage
				AD4: formatAAAAnalogValue(parts[3]), // Updated to format as voltage
				AD5: formatAAAAnalogValue(parts[4]), // Updated to format as voltage
			}
		}
	}

	// Try to extract from a map representation (might be from CCC protocol)
	if analogMap, ok := packetMap["AnalogsInput"].(map[string]interface{}); ok {
		// For CCE protocol, analog values are divided by 100 for voltage
		// Format values to include decimal point if needed
		ad1 := formatAnalogValue(analogMap, "ad1")
		ad2 := formatAnalogValue(analogMap, "ad2")
		// Get ad3 value properly formatted
		ad3 := formatAnalogValue(analogMap, "ad3")
		ad4 := formatAnalogValue(analogMap, "ad4")
		ad5 := formatAnalogValue(analogMap, "ad5")

		return &models.AnalogInputs{
			AD1: &ad1,
			AD2: &ad2,
			AD3: &ad3, // Use the formatted ad3 value instead of calling getStringPointer
			AD4: &ad4,
			AD5: &ad5,
		}
	}

	// Standard extraction from map for regular Jono format
	return &models.AnalogInputs{
		AD1:  getStringPointer(packetMap, "AD1"),
		AD2:  getStringPointer(packetMap, "AD2"),
		AD3:  getStringPointer(packetMap, "AD3"),
		AD4:  getStringPointer(packetMap, "AD4"),
		AD5:  getStringPointer(packetMap, "AD5"),
		AD6:  getStringPointer(packetMap, "AD6"),
		AD7:  getStringPointer(packetMap, "AD7"),
		AD8:  getStringPointer(packetMap, "AD8"),
		AD9:  getStringPointer(packetMap, "AD9"),
		AD10: getStringPointer(packetMap, "AD10"),
	}
}

// Helper function to format analog values correctly (divide by 100 for voltage)
func formatAnalogValue(analogMap map[string]interface{}, key string) string {
	if value, ok := analogMap[key]; ok {
		// Try to get the numeric value
		var numericValue float64
		switch v := value.(type) {
		case float64:
			numericValue = v
		case int:
			numericValue = float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				numericValue = f
			}
		}

		// Format with 2 decimal places for voltage
		return fmt.Sprintf("%.2f", numericValue/100)
	}
	return "0.00"
}

// Helper function to safely get values from a map
func getValueOrEmpty(data map[string]interface{}, key string) interface{} {
	if value, ok := data[key]; ok {
		return value
	}
	return ""
}

// 📌 Función para extraer BaseStationInfo
func extractBaseStationInfo(packetMap map[string]interface{}) *models.BaseStationInfo {
	// Try to get BaseStationInfo from a map
	if bsInfo, ok := packetMap["BaseStationInfo"].(map[string]interface{}); ok {
		// Format the cell ID properly to avoid exponential notation
		var cellId *string
		if cellIdVal, exists := bsInfo["cellId"]; exists && cellIdVal != nil {
			// For CCE protocol, ensure cellId is properly formatted as a string
			cellIdStr := fmt.Sprintf("%v", cellIdVal)
			// If it contains 'e' or 'E' (exponential), convert to regular string
			if strings.Contains(cellIdStr, "e") || strings.Contains(cellIdStr, "E") {
				if f, err := strconv.ParseFloat(cellIdStr, 64); err == nil {
					cellIdStr = fmt.Sprintf("%.0f", f)
				}
			}
			cellId = &cellIdStr
		}

		// Get other fields normally
		return &models.BaseStationInfo{
			MCC:    getStringPointer(bsInfo, "mcc"),
			MNC:    getStringPointer(bsInfo, "mnc"),
			LAC:    getStringPointer(bsInfo, "lac"),
			CellID: cellId,
		}
	}

	// Standard direct field extraction
	return &models.BaseStationInfo{
		MCC:    getStringPointer(packetMap, "MCC"),
		MNC:    getStringPointer(packetMap, "MNC"),
		LAC:    getStringPointer(packetMap, "LAC"),
		CellID: getStringPointer(packetMap, "CellID"),
	}
}

// 📌 Funciones para extraer otras entidades
func extractOutputPortStatus(packetMap map[string]interface{}) *models.OutputPortStatus {
	return &models.OutputPortStatus{
		Output1: getStringPointer(packetMap, "Output1"),
		Output2: getStringPointer(packetMap, "Output2"),
		Output3: getStringPointer(packetMap, "Output3"),
		Output4: getStringPointer(packetMap, "Output4"),
		Output5: getStringPointer(packetMap, "Output5"),
		Output6: getStringPointer(packetMap, "Output6"),
		Output7: getStringPointer(packetMap, "Output7"),
		Output8: getStringPointer(packetMap, "Output8"),
	}
}

func extractInputPortStatus(packetMap map[string]interface{}) *models.InputPortStatus {
	return &models.InputPortStatus{
		Input1: getStringPointer(packetMap, "Input1"),
		Input2: getStringPointer(packetMap, "Input2"),
		Input3: getStringPointer(packetMap, "Input3"),
		Input4: getStringPointer(packetMap, "Input4"),
		Input5: getStringPointer(packetMap, "Input5"),
		Input6: getStringPointer(packetMap, "Input6"),
		Input7: getStringPointer(packetMap, "Input7"),
		Input8: getStringPointer(packetMap, "Input8"),
	}
}

func extractSystemFlag(packetMap map[string]interface{}) *models.SystemFlag {
	return &models.SystemFlag{
		EEP2:                getStringPointer(packetMap, "EEP2"),
		ACC:                 getStringPointer(packetMap, "ACC"),
		AntiTheft:           getStringPointer(packetMap, "AntiTheft"),
		VibrationFlag:       getStringPointer(packetMap, "VibrationFlag"),
		MovingFlag:          getStringPointer(packetMap, "MovingFlag"),
		ExternalPowerSupply: getStringPointer(packetMap, "ExternalPowerSupply"),
		Charging:            getStringPointer(packetMap, "Charging"),
		SleepMode:           getStringPointer(packetMap, "SleepMode"),
		FMS:                 getStringPointer(packetMap, "FMS"),
		FMSFunction:         getStringPointer(packetMap, "FMSFunction"),
		SystemFlagExtras:    getStringPointer(packetMap, "SystemFlagExtras"),
	}
}

// Extract temperature sensor data
func extractTemperatureSensor(packetMap map[string]interface{}) *models.TemperatureSensor {
	return &models.TemperatureSensor{
		SensorNumber: getStringPointer(packetMap, "SensorNumber"),
		Value:        getStringPointer(packetMap, "Value"),
	}
}

// Extract temperature and humidity sensor data
func extractTemperatureAndHumidity(packetMap map[string]interface{}) *models.TemperatureAndHumidity {
	return &models.TemperatureAndHumidity{
		DeviceName:           getStringPointer(packetMap, "DeviceName"),
		MAC:                  getStringPointer(packetMap, "MAC"),
		BatteryPower:         getStringPointer(packetMap, "BatteryPower"),
		Temperature:          getStringPointer(packetMap, "Temperature"),
		Humidity:             getStringPointer(packetMap, "Humidity"),
		AlertHighTemperature: getStringPointer(packetMap, "AlertHighTemperature"),
		AlertLowTemperature:  getStringPointer(packetMap, "AlertLowTemperature"),
		AlertHighHumidity:    getStringPointer(packetMap, "AlertHighHumidity"),
		AlertLowHumidity:     getStringPointer(packetMap, "AlertLowHumidity"),
	}
}

// Extract Bluetooth beacon data
func extractBluetoothBeacon(packetMap map[string]interface{}, beaconKey string) *models.BluetoothBeacon {
	if beaconMap, ok := packetMap[beaconKey].(map[string]interface{}); ok {
		return &models.BluetoothBeacon{
			Version:        getStringPointer(beaconMap, "Version"),
			DeviceName:     getStringPointer(beaconMap, "DeviceName"),
			MAC:            getStringPointer(beaconMap, "MAC"),
			BatteryPower:   getStringPointer(beaconMap, "BatteryPower"),
			SignalStrength: getStringPointer(beaconMap, "SignalStrength"),
		}
	}

	return &models.BluetoothBeacon{
		Version:        getStringPointer(packetMap, beaconKey+"_Version"),
		DeviceName:     getStringPointer(packetMap, beaconKey+"_DeviceName"),
		MAC:            getStringPointer(packetMap, beaconKey+"_MAC"),
		BatteryPower:   getStringPointer(packetMap, beaconKey+"_BatteryPower"),
		SignalStrength: getStringPointer(packetMap, beaconKey+"_SignalStrength"),
	}
}

// 📌 Función para extraer CameraStatus
func extractCameraStatus(packetMap map[string]interface{}) *models.CameraStatus {
	return &models.CameraStatus{
		CameraNumber: getStringPointer(packetMap, "CameraNumber"),
		Status:       getStringPointer(packetMap, "Status"),
	}
}

// 📌 Función para extraer CurrentNetworkInfo
func extractCurrentNetworkInfo(packetMap map[string]interface{}) *models.CurrentNetworkInfo {
	return &models.CurrentNetworkInfo{
		Version:    getStringPointer(packetMap, "CurrentNetworkInfo_Version"),
		Type:       getStringPointer(packetMap, "CurrentNetworkInfo_Type"),
		Descriptor: getStringPointer(packetMap, "CurrentNetworkInfo_Descriptor"),
	}
}

// 📌 Función para extraer FatigueDrivingInformation
func extractFatigueDrivingInformation(packetMap map[string]interface{}) *models.FatigueDrivingInformation {
	return &models.FatigueDrivingInformation{
		Version:    getStringPointer(packetMap, "FatigueDrivingInformation_Version"),
		Type:       getStringPointer(packetMap, "FatigueDrivingInformation_Type"),
		Descriptor: getStringPointer(packetMap, "FatigueDrivingInformation_Descriptor"),
	}
}

// 📌 Función para extraer AdditionalAlertInfoADASDMS
func extractAdditionalAlertInfoADASDMS(packetMap map[string]interface{}) *models.AdditionalAlertInfoADASDMS {
	return &models.AdditionalAlertInfoADASDMS{
		AlarmProtocol: getStringPointer(packetMap, "AlarmProtocol"),
		AlarmType:     getStringPointer(packetMap, "AlarmType"),
		PhotoName:     getStringPointer(packetMap, "PhotoName"),
	}
}

// 📌 Constructor de IoPortsStatus que asigna valores por defecto en 0 si no existen en el JSON
func extractIoPortsStatus(packetMap map[string]interface{}) *models.IoPortsStatus {
	return &models.IoPortsStatus{
		Port1: getIntValueOrDefault(packetMap, "Port1", 0),
		Port2: getIntValueOrDefault(packetMap, "Port2", 0),
		Port3: getIntValueOrDefault(packetMap, "Port3", 0),
		Port4: getIntValueOrDefault(packetMap, "Port4", 0),
		Port5: getIntValueOrDefault(packetMap, "Port5", 0),
		Port6: getIntValueOrDefault(packetMap, "Port6", 0),
		Port7: getIntValueOrDefault(packetMap, "Port7", 0),
		Port8: getIntValueOrDefault(packetMap, "Port8", 0),
	}
}

// 📌 Función auxiliar para obtener un int o asignar el valor por defecto si no está presente
func getIntValueOrDefault(data map[string]interface{}, key string, defaultValue int) int {
	if value, exists := data[key]; exists {
		if intValue, ok := value.(float64); ok {
			return int(intValue)
		}
	}
	return defaultValue
}
