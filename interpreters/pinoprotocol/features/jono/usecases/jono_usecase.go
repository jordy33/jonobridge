package usecases

import (
	"encoding/json"
	"fmt"
	"math"
	"pinoprotocol/features/jono/models"
	"sort"
	"strconv"
	"strings"
)

// Package-level verbose flag for debugging
var verbose = false

// EnableVerboseLogging enables detailed logging in this package
func EnableVerboseLogging(enable bool) {
	verbose = enable
}

func GetDataJono(data string) (string, error) {
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &rawData); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Debug the input data
	if verbose {
		// Check for alarm-specific fields that indicate this is an alarm packet
		alarmIndications := []string{}
		if msg, ok := rawData["Message"].(string); ok && strings.Contains(strings.ToLower(msg), "alarm") {
			alarmIndications = append(alarmIndications, fmt.Sprintf("Message contains 'alarm': %s", msg))
		}
		if ec, ok := rawData["EventCode"].(string); ok && ec != "35" {
			alarmIndications = append(alarmIndications, fmt.Sprintf("Non-normal EventCode: %s", ec))
		}
		if alarm, ok := rawData["AlarmType"].(string); ok {
			alarmIndications = append(alarmIndications, fmt.Sprintf("AlarmType field: %s", alarm))
		}

		if len(alarmIndications) > 0 {
			fmt.Println("DEBUG: Input data appears to be an alarm packet:")
			for _, ind := range alarmIndications {
				fmt.Println("  - " + ind)
			}
		} else {
			fmt.Println("DEBUG: Input data appears to be a standard location packet")
		}
	}

	// 📌 Crear el modelo `ParsedModel`
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

	// After creating packets, debug the EventCode values to ensure consistency
	if verbose {
		for key, packet := range parsedModel.ListPackets {
			fmt.Printf("DEBUG: Packet %s has EventCode %d (%s)\n",
				key, packet.EventCode.Code, packet.EventCode.Name)
		}
	}

	// 📌 Convertir a JSON
	jsonString, err := parsedModel.ToPrettyJSON()
	if err != nil {
		return "", fmt.Errorf("failed to convert to JSON: %w", err)
	}

	//fmt.Println("\n\n", jsonString, "\n\n")
	return jsonString, nil
}

// 📌 Función auxiliar para crear un `Packet`
func createPacket(packetMap map[string]interface{}) models.Packet {
	// Debug details about the input packet to help diagnose issues
	if verbose {
		fmt.Println("\nDEBUG: Creating packet from input map with keys:", getKeysAsSortedString(packetMap))

		// Check for specific alarm-related fields
		keysToCheck := []string{"EventCode", "AlarmType", "Message", "AlarmAndLanguage", "terminalInformationContent"}
		for _, key := range keysToCheck {
			if val, exists := packetMap[key]; exists {
				fmt.Printf("DEBUG: Input map contains %s = %v\n", key, val)
			}
		}
	}

	// Create EventCode struct with proper types
	eventCode := models.EventCode{
		Code: 35,                // Default value for normal conditions
		Name: "Normal Location", // Default name
	}

	// Log alarm detection process
	if verbose {
		fmt.Println("DEBUG: Starting alarm detection process in createPacket")
	}

	// First priority - check for explicit EventCode field (this is most reliable)
	if ecValue, exists := packetMap["EventCode"]; exists {
		if verbose {
			fmt.Printf("DEBUG: Found explicit EventCode: %v (type: %T)\n", ecValue, ecValue)
		}

		if strValue, ok := ecValue.(string); ok {
			if intValue, err := strconv.Atoi(strValue); err == nil {
				eventCode.Code = intValue

				// Set corresponding name based on code
				switch intValue {
				case 1:
					eventCode.Name = "SOS"
				case 20:
					eventCode.Name = "Fence In Alarm"
				case 21:
					eventCode.Name = "Fence Out Alarm"
				case 23:
					eventCode.Name = "Power Cut Alarm"
				case 35:
					eventCode.Name = "Normal Location"
				case 50:
					eventCode.Name = "Alarm"
				case 79:
					eventCode.Name = "Shock Alarm"
				default:
					eventCode.Name = fmt.Sprintf("Event %d", intValue)
				}

				if verbose {
					fmt.Printf("DEBUG: Set EventCode from string: %d (%s)\n",
						eventCode.Code, eventCode.Name)
				}
			}
		} else if floatValue, ok := ecValue.(float64); ok {
			eventCode.Code = int(floatValue)
			// Set name based on code same as above
			switch eventCode.Code {
			case 1:
				eventCode.Name = "SOS"
			case 20:
				eventCode.Name = "Fence In Alarm"
			case 21:
				eventCode.Name = "Fence Out Alarm"
			case 23:
				eventCode.Name = "Power Cut Alarm"
			case 35:
				eventCode.Name = "Normal Location"
			case 50:
				eventCode.Name = "Alarm"
			case 79:
				eventCode.Name = "Shock Alarm"
			default:
				eventCode.Name = fmt.Sprintf("Event %d", eventCode.Code)
			}

			if verbose {
				fmt.Printf("DEBUG: Set EventCode from float: %d (%s)\n",
					eventCode.Code, eventCode.Name)
			}
		}
	}

	// Second priority - check explicit AlarmAndLanguage information
	if alarmInfo, exists := packetMap["AlarmAndLanguage"].(map[string]interface{}); exists {
		if verbose {
			fmt.Printf("DEBUG: Found AlarmAndLanguage: %v\n", alarmInfo)
		}

		// First check for EventCode in AlarmAndLanguage as it's most reliable
		if ec, exists := alarmInfo["EventCode"].(string); exists {
			if ecInt, err := strconv.Atoi(ec); err == nil {
				eventCode.Code = ecInt

				// Update name to match the code
				switch ecInt {
				case 1:
					eventCode.Name = "SOS"
				case 20:
					eventCode.Name = "Fence In Alarm"
				case 21:
					eventCode.Name = "Fence Out Alarm"
				case 23:
					eventCode.Name = "Power Cut Alarm"
				case 35:
					eventCode.Name = "Normal Location"
				case 50:
					eventCode.Name = "Alarm"
				case 79:
					eventCode.Name = "Shock Alarm"
				default:
					eventCode.Name = fmt.Sprintf("Event %d", ecInt)
				}

				if verbose {
					fmt.Printf("DEBUG: Set EventCode from AlarmAndLanguage: %d (%s)\n",
						eventCode.Code, eventCode.Name)
				}
			}
		}

		// If no EventCode found, try to determine from Alarm description
		if eventCode.Code == 35 && eventCode.Name == "Normal Location" {
			if alarm, ok := alarmInfo["Alarm"].(string); ok && alarm != "Normal" {
				if verbose {
					fmt.Printf("DEBUG: Found alarm string in AlarmAndLanguage: %s\n", alarm)
				}

				// Set the correct EventCode based on alarm type
				switch alarm {
				case "SOS":
					eventCode.Code = 1
					eventCode.Name = "SOS"
				case "Power Cut Alarm":
					eventCode.Code = 23
					eventCode.Name = "Power Cut Alarm"
				case "Shock Alarm":
					eventCode.Code = 79
					eventCode.Name = "Shock Alarm"
				case "Fence In Alarm":
					eventCode.Code = 20
					eventCode.Name = "Fence In Alarm"
				case "Fence Out Alarm":
					eventCode.Code = 21
					eventCode.Name = "Fence Out Alarm"
				default:
					eventCode.Code = 50 // Generic alarm if type not specifically known
					eventCode.Name = alarm
				}

				if verbose {
					fmt.Printf("DEBUG: Set EventCode from alarm string: %d (%s)\n",
						eventCode.Code, eventCode.Name)
				}
			}
		}
	}

	// Third priority - check if there's an AlarmType field
	if alarmType, exists := packetMap["AlarmType"].(string); exists && alarmType != "Normal" {
		if verbose {
			fmt.Printf("DEBUG: Found AlarmType: %s\n", alarmType)
		}

		// Only override if we're still on default Normal Location
		if eventCode.Code == 35 {
			switch alarmType {
			case "SOS":
				eventCode.Code = 1
				eventCode.Name = "SOS"
			case "Power Cut Alarm":
				eventCode.Code = 23
				eventCode.Name = "Power Cut Alarm"
			case "Shock Alarm":
				eventCode.Code = 79
				eventCode.Name = "Shock Alarm"
			case "Fence In Alarm":
				eventCode.Code = 20
				eventCode.Name = "Fence In Alarm"
			case "Fence Out Alarm":
				eventCode.Code = 21
				eventCode.Name = "Fence Out Alarm"
			default:
				eventCode.Code = 50 // Generic alarm
				eventCode.Name = alarmType
			}

			if verbose {
				fmt.Printf("DEBUG: Set EventCode from AlarmType: %d (%s)\n",
					eventCode.Code, eventCode.Name)
			}
		}
	}

	// Fourth priority - check for Message field indicating alarm
	if message, exists := packetMap["Message"].(string); exists &&
		(strings.Contains(message, "Alarm") || strings.Contains(message, "alarm") ||
			strings.Contains(message, "SOS")) {
		if verbose {
			fmt.Printf("DEBUG: Found alarm in Message: %s\n", message)
		}

		// Only override if we're still on default Normal Location
		if eventCode.Code == 35 {
			// Try to determine alarm type from message
			if strings.Contains(message, "SOS") {
				eventCode.Code = 1
				eventCode.Name = "SOS"
			} else if strings.Contains(message, "Power Cut") {
				eventCode.Code = 23
				eventCode.Name = "Power Cut Alarm"
			} else if strings.Contains(message, "Shock") {
				eventCode.Code = 79
				eventCode.Name = "Shock Alarm"
			} else if strings.Contains(message, "Fence In") {
				eventCode.Code = 20
				eventCode.Name = "Fence In Alarm"
			} else if strings.Contains(message, "Fence Out") {
				eventCode.Code = 21
				eventCode.Name = "Fence Out Alarm"
			} else {
				// Generic alarm
				eventCode.Code = 50
				eventCode.Name = "Alarm"
			}

			if verbose {
				fmt.Printf("DEBUG: Set EventCode from Message: %d (%s)\n",
					eventCode.Code, eventCode.Name)
			}
		}
	}

	// Fifth priority - check if we need to infer event code from terminal information
	if eventCode.Code == 35 {
		if termInfo, exists := packetMap["terminalInformationContent"].(string); exists {
			if verbose {
				fmt.Printf("DEBUG: Found terminalInformationContent: %s\n", termInfo)
			}

			if strings.Contains(termInfo, "SOS") {
				eventCode.Code = 1
				eventCode.Name = "SOS"
				if verbose {
					fmt.Printf("DEBUG: Inferred EventCode SOS from terminal info\n")
				}
			} else if strings.Contains(termInfo, "Shock Alarm") {
				eventCode.Code = 79
				eventCode.Name = "Shock Alarm"
				if verbose {
					fmt.Printf("DEBUG: Inferred EventCode Shock Alarm from terminal info\n")
				}
			} else if strings.Contains(termInfo, "Power Cut") {
				eventCode.Code = 23
				eventCode.Name = "Power Cut Alarm"
				if verbose {
					fmt.Printf("DEBUG: Inferred EventCode Power Cut from terminal info\n")
				}
			}
		}
	}

	// Final debug - show what EventCode we're using
	if verbose {
		fmt.Printf("DEBUG: Final EventCode: %d (%s)\n", eventCode.Code, eventCode.Name)
	}

	// Get altitude with default of 0
	altitude := 0
	if altPtr := getIntPointer(packetMap, "Altitude"); altPtr != nil {
		altitude = *altPtr
	}

	// Extract Direction/Course properly
	direction := 0
	if dirPtr := getIntPointer(packetMap, "Direction"); dirPtr != nil {
		direction = *dirPtr
	} else if coursePtr := getIntPointer(packetMap, "Course"); coursePtr != nil {
		direction = *coursePtr
	}

	// Extract GSM Signal Strength
	gsmSignalStrength := extractGSMSignalStrength(packetMap)

	// Set HDOP with default value of 1
	hdop := 1.0
	if hdopPtr := getFloatPointer(packetMap, "HDOP"); hdopPtr != nil {
		hdop = *hdopPtr
	}

	// Add debug for mileage extraction
	if verbose {
		if mileageVal, exists := packetMap["Mileage"]; exists {
			fmt.Printf("DEBUG: Found mileage in packet data: %v (type: %T)\n", mileageVal, mileageVal)
		}
	}

	// Extract mileage value
	var mileage *int
	if mileageVal, exists := packetMap["Mileage"]; exists {
		switch v := mileageVal.(type) {
		case float64:
			mi := int(v)
			mileage = &mi
			if verbose {
				fmt.Printf("DEBUG: Converted mileage from float64 %v to int %d\n", v, mi)
			}
		case int:
			mileage = &v
			if verbose {
				fmt.Printf("DEBUG: Using mileage directly as int %d\n", v)
			}
		case string:
			if mi, err := strconv.Atoi(v); err == nil {
				mileage = &mi
				if verbose {
					fmt.Printf("DEBUG: Converted mileage from string %s to int %d\n", v, mi)
				}
			}
		default:
			if verbose {
				fmt.Printf("DEBUG: Unexpected mileage type: %T\n", mileageVal)
			}
		}
	}

	// Extract NumberOfSatellites properly
	numberOfSatellites := 0
	if satsVal, exists := packetMap["NumberOfSatellites"]; exists {
		if verbose {
			fmt.Printf("DEBUG: Found NumberOfSatellites: %v (type: %T)\n", satsVal, satsVal)
		}
		switch v := satsVal.(type) {
		case float64:
			numberOfSatellites = int(v)
		case int:
			numberOfSatellites = v
		}
		if verbose {
			fmt.Printf("DEBUG: Converted NumberOfSatellites to: %d\n", numberOfSatellites)
		}
	}

	packet := models.Packet{
		Altitude:                     altitude,
		Datetime:                     getStringPointer(packetMap, "Datetime"),
		EventCode:                    eventCode,
		Latitude:                     getFloatPointer(packetMap, "Latitude"),
		Longitude:                    getFloatPointer(packetMap, "Longitude"),
		Speed:                        getIntPointer(packetMap, "Speed"),
		PositioningStatus:            getStringPointer(packetMap, "PositioningStatus"),
		Direction:                    &direction,
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
		GSMSignalStrength:            gsmSignalStrength,
		HDOP:                         &hdop,
		Mileage:                      mileage,            // Add mileage to the packet
		NumberOfSatellites:           numberOfSatellites, // Assign the extracted value
	}

	// Add final debug check
	if verbose {
		if packet.Mileage != nil {
			fmt.Printf("DEBUG: Final packet mileage value: %d\n", *packet.Mileage)
		} else {
			fmt.Printf("DEBUG: Final packet mileage is nil\n")
		}
		if verbose {
			fmt.Printf("DEBUG: Final packet NumberOfSatellites: %d\n", packet.NumberOfSatellites)
		}
	}

	return packet
}

// Updated function for mapping GSM signal strength values
func mapGSMSignalStrength(originalValue int) int {
	// Map according to the required specification
	switch originalValue {
	case 0x00:
		return 0 // No signal
	case 0x01:
		return 8 // Extremely weak signal
	case 0x02:
		return 16 // Very weak signal
	case 0x03:
		return 23 // Good signal
	case 0x04:
		return 31 // Strong signal
	default:
		// For out of range values, use appropriate mapping
		if originalValue < 0 {
			return 0 // No signal for negative values
		} else if originalValue > 4 {
			return 31 // Max signal for values above range
		}
		return 16 // Default to "very weak" (16) for unknown values
	}
}

// Updated function to properly handle string descriptions
func extractGSMSignalStrength(packetMap map[string]interface{}) *int {
	// First try to get the value directly as a numeric value
	if gsmPtr := getIntPointer(packetMap, "GSMSignalStrength"); gsmPtr != nil {
		if verbose {
			fmt.Printf("DEBUG: Original GSM signal strength: %d\n", *gsmPtr)
		}
		mappedValue := mapGSMSignalStrength(*gsmPtr)
		if verbose {
			fmt.Printf("DEBUG: Mapped GSM signal strength: %d\n", mappedValue)
		}
		return &mappedValue
	}

	// Try to get from GSMSignalStrengthString field
	if gsm, exists := packetMap["GSMSignalStrengthString"].(string); exists {
		var originalValue int
		switch gsm {
		case "No signal":
			originalValue = 0x00
		case "Extremely weak signal":
			originalValue = 0x01
		case "Very weak signal":
			originalValue = 0x02
		case "Good signal":
			originalValue = 0x03
		case "Strong signal":
			originalValue = 0x04
		default:
			originalValue = 0x02 // Default to "Very weak signal" if unknown
		}
		mappedValue := mapGSMSignalStrength(originalValue)
		if verbose {
			fmt.Printf("DEBUG: Mapped GSM signal from string '%s': %d -> %d\n",
				gsm, originalValue, mappedValue)
		}
		return &mappedValue
	}

	// Default value if no signal strength information found
	defaultValue := mapGSMSignalStrength(0x02) // Default to "Very weak signal"
	return &defaultValue
}

// 📌 Función para obtener un puntero a un string
func getStringPointer(data map[string]interface{}, key string) *string {
	if value, exists := data[key]; exists {
		strValue := fmt.Sprintf("%v", value)
		return &strValue
	}
	return nil
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

// 📌 Función para obtener un puntero a un float64
func getFloatPointer(data map[string]interface{}, key string) *float64 {
	if value, exists := data[key]; exists {
		if floatValue, ok := value.(float64); ok {
			return &floatValue
		}
	}
	return nil
}

// 📌 Función para extraer AnalogInputs y establecer AD4 al nivel de batería
func extractAnalogInputs(packetMap map[string]interface{}) *models.AnalogInputs {
	// Initialize analog inputs
	analogInputs := &models.AnalogInputs{
		AD1:  getStringPointer(packetMap, "AD1"),
		AD2:  getStringPointer(packetMap, "AD2"),
		AD3:  getStringPointer(packetMap, "AD3"),
		AD5:  getStringPointer(packetMap, "AD5"),
		AD6:  getStringPointer(packetMap, "AD6"),
		AD7:  getStringPointer(packetMap, "AD7"),
		AD8:  getStringPointer(packetMap, "AD8"),
		AD9:  getStringPointer(packetMap, "AD9"),
		AD10: getStringPointer(packetMap, "AD10"),
	}

	// Special handling for AD4 - should contain battery voltage as a hex value

	// Try to get VoltageValue first (top priority)
	var voltage float64 = 12.0 // Default value if we can't determine
	valueFound := false

	// First priority: explicit VoltageValue
	if vv, exists := packetMap["VoltageValue"]; exists {
		if floatValue, ok := vv.(float64); ok && floatValue >= 0 {
			voltage = floatValue
			valueFound = true
			if verbose {
				fmt.Printf("DEBUG: Found direct VoltageValue: %.1f\n", voltage)
			}
		}
	}

	// Second priority: BatteryLevel conversion
	if !valueFound {
		if bl, exists := packetMap["BatteryLevel"]; exists {
			batteryLevel := 0
			if floatValue, ok := bl.(float64); ok && floatValue >= 0 {
				batteryLevel = int(floatValue)
				valueFound = true
			} else if intValue, ok := bl.(int); ok && intValue >= 0 {
				batteryLevel = intValue
				valueFound = true
			}

			if valueFound {
				// Convert battery level to voltage according to the scale
				switch batteryLevel {
				case 0:
					voltage = 0.0 // No Power
				case 1:
					voltageCalc := (3.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 2:
					voltageCalc := (6.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 3:
					voltageCalc := (9.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 4:
					voltageCalc := (12.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 5:
					voltageCalc := (12.5 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 6:
					voltageCalc := (13.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				}
				if verbose {
					fmt.Printf("DEBUG: Converted battery level %d to voltage %.1f\n", batteryLevel, voltage)
				}
			}
		}
	}

	// Third priority: Parse from VoltageLevel string
	if !valueFound {
		if vl, exists := packetMap["VoltageLevel"]; exists {
			if strValue, ok := vl.(string); ok && strValue != "" {
				batteryLevel := 3 // Default to medium
				// Extract level from string description
				if strValue == "Very High" {
					batteryLevel = 6
				} else if strValue == "High" {
					batteryLevel = 5
				} else if strValue == "Medium" {
					batteryLevel = 4
				} else if strValue == "Low Battery" || strValue == "Low Battery (can be used normally)" {
					batteryLevel = 3
				} else if strValue == "Very Low Battery" || strValue == "Very Low Battery (Low Battery Alarm)" {
					batteryLevel = 2
				} else if strValue == "Extremely Low Battery" || strValue == "Extremely Low Battery (not enough for calling or sending text messages)" {
					batteryLevel = 1
				} else if strValue == "No Power" {
					batteryLevel = 0
				}
				// Convert level to voltage
				switch batteryLevel {
				case 0:
					voltage = 0.0 // No Power
				case 1:
					voltageCalc := (3.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 2:
					voltageCalc := (6.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 3:
					voltageCalc := (9.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 4:
					voltageCalc := (12.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 5:
					voltageCalc := (12.5 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 6:
					voltageCalc := (13 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				}
				// Convert to float64 since voltage is float64
				// High
				valueFound = true
				if verbose {
					fmt.Printf("DEBUG: Extracted battery level %d from description '%s', voltage: %.1f\n",
						batteryLevel, strValue, voltage)
				}
			}
		}
	}

	// Fourth priority: VoltageLevelByte
	if !valueFound {
		if vlb, exists := packetMap["VoltageLevelByte"]; exists {
			batteryLevel := 0
			if floatValue, ok := vlb.(float64); ok && floatValue >= 0 && floatValue <= 6 {
				batteryLevel = int(floatValue)
				valueFound = true
			} else if intValue, ok := vlb.(int); ok && intValue >= 0 && intValue <= 6 {
				batteryLevel = intValue
				valueFound = true
			}

			if valueFound {
				// Convert battery level to voltage
				switch batteryLevel {
				case 0:
					voltage = 0.0 // No Power
				case 1:
					voltageCalc := (3.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 2:
					voltageCalc := (6.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 3:
					voltageCalc := (12.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 4:
					voltageCalc := (12.3 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 5:
					voltageCalc := (12.5 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				case 6:
					voltageCalc := (13.0 * 1024.0) / 6.0
					voltage = float64(int(math.Round(voltageCalc)))
				}
				if verbose {
					fmt.Printf("DEBUG: Using VoltageLevelByte %d for voltage: %.1f\n", batteryLevel, voltage)
				}
			}
		}
	}

	// Convert voltage to integer and then to hex string without "0x" prefix
	voltageInt := int(voltage)                  // Convert to integer
	voltageHex := fmt.Sprintf("%X", voltageInt) // This will be uppercase hex without "0x" prefix
	analogInputs.AD4 = &voltageHex

	if verbose {
		fmt.Printf("DEBUG: Final AD4 voltage: %d (decimal) -> %s (hex)\n", voltageInt, voltageHex)
	}

	return analogInputs
}

// 📌 Función para extraer BaseStationInfo
func extractBaseStationInfo(packetMap map[string]interface{}) *models.BaseStationInfo {
	baseStationInfo := &models.BaseStationInfo{
		MCC:    getStringPointer(packetMap, "MCC"),
		MNC:    getStringPointer(packetMap, "MNC"),
		LAC:    getStringPointer(packetMap, "LAC"),
		CellID: getStringPointer(packetMap, "CellID"),
	}

	// If direct fields are not available, try to extract from BaseStationInfo object
	if baseStationInfo.MCC == nil || baseStationInfo.MNC == nil || baseStationInfo.LAC == nil || baseStationInfo.CellID == nil {
		if bsInfo, ok := packetMap["BaseStationInfo"].(map[string]interface{}); ok {
			// Try extracting from BaseStationInfo map
			if baseStationInfo.MCC == nil {
				if mcc, exists := bsInfo["mmc"]; exists {
					mccStr := fmt.Sprintf("%v", mcc)
					baseStationInfo.MCC = &mccStr
				}
			}
			if baseStationInfo.MNC == nil {
				if mnc, exists := bsInfo["mnc"]; exists {
					mncStr := fmt.Sprintf("%v", mnc)
					baseStationInfo.MNC = &mncStr
				}
			}
			if baseStationInfo.LAC == nil {
				if lac, exists := bsInfo["lac"]; exists {
					lacStr := fmt.Sprintf("%v", lac)
					baseStationInfo.LAC = &lacStr
				}
			}
			if baseStationInfo.CellID == nil {
				if cellID, exists := bsInfo["cellId"]; exists {
					cellIDStr := fmt.Sprintf("%v", cellID)
					baseStationInfo.CellID = &cellIDStr
				}
			}
		}
	}

	return baseStationInfo
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

// 📌 Función para extraer InputPortStatus
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

// 📌 Función para extraer SystemFlag
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

func extractTemperatureSensor(packetMap map[string]interface{}) *models.TemperatureSensor {
	return &models.TemperatureSensor{
		SensorNumber: getStringPointer(packetMap, "SensorNumber"),
		Value:        getStringPointer(packetMap, "Value"),
	}
}

func extractBluetoothBeacon(packetMap map[string]interface{}, key string) *models.BluetoothBeacon {
	return &models.BluetoothBeacon{
		Version:        getStringPointer(packetMap, key+"_Version"),
		DeviceName:     getStringPointer(packetMap, key+"_DeviceName"),
		MAC:            getStringPointer(packetMap, key+"_MAC"),
		BatteryPower:   getStringPointer(packetMap, key+"_BatteryPower"),
		SignalStrength: getStringPointer(packetMap, key+"_SignalStrength"),
	}
}

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

// Helper function to get a sorted string representation of map keys
func getKeysAsSortedString(m map[string]interface{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
