package usecases

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"pinoprotocol/features/pino_protocol/helpers"
	"pinoprotocol/features/pino_protocol/models"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Enable debug logging from external packages
func EnableVerboseLogging(enable bool) {
	verbose = enable
}

func IsLoginPacket(data []byte) bool {
	return len(data) > 3 && data[0] == 0x78 && data[1] == 0x78 && data[3] == 0x01
}

func IsHeartbeatPacket(data []byte) bool {
	return len(data) > 3 && data[0] == 0x78 && data[1] == 0x78 && data[3] == 0x13
}

func IsStandardLocationPacket(data []byte) bool {
	return len(data) > 3 && data[0] == 0x78 && data[1] == 0x78 && data[3] == 0x12
}

func IsStandardAlarmPacket(data []byte) bool {
	return len(data) > 3 && data[0] == 0x78 && data[1] == 0x78 && data[3] == 0x16
}

func IsStringInformationPacket(data []byte) bool {
	return len(data) > 3 && data[0] == 0x78 && data[1] == 0x78 && data[3] == 0x15
}

func BuildLoginResponse(data []byte) []byte {
	startBit := []byte{0x78, 0x78}                  // Inicio de la trama
	length := byte(0x05)                            // Longitud del paquete
	messageType := byte(0x01)                       // Tipo de mensaje de respuesta (Login Response)
	serialNumber := data[len(data)-4 : len(data)-2] // Serial recibido del paquete de Login
	crc := helpers.CalculateCRC(append([]byte{length, messageType}, serialNumber...))
	stopBit := []byte{0x0D, 0x0A} // Fin de la trama

	frame := append(startBit, length, messageType)
	frame = append(frame, serialNumber...)
	frame = append(frame, crc...)
	frame = append(frame, stopBit...)
	return frame
}

func BuildHeartbeatResponse() []byte {
	startBit := []byte{0x78, 0x78}     // Inicio de la trama
	length := byte(0x05)               // Longitud del paquete
	messageType := byte(0x13)          // Tipo de mensaje de respuesta (Heartbeat Response)
	serialNumber := []byte{0x00, 0x01} // Número de serie (puede ser estático o incrementar)
	crc := helpers.CalculateCRC(append([]byte{length, messageType}, serialNumber...))
	stopBit := []byte{0x0D, 0x0A} // Fin de la trama

	frame := append(startBit, length, messageType)
	frame = append(frame, serialNumber...)
	frame = append(frame, crc...)
	frame = append(frame, stopBit...)
	return frame
}

func DecodeStandardLocationData(data []byte, imei string, isAlarm bool) (*models.LocationPacketModel, error) {
	if len(data) < 29 {
		return nil, fmt.Errorf("error: data too short")
	}
	year := 2000 + int(data[4])
	month := int(data[5])
	day := int(data[6])
	hour := int(data[7])
	minute := int(data[8])
	second := int(data[9])
	dateTime := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ", year, month, day, hour, minute, second)
	satellites := int((data[10] & 0xF0) >> 4)
	status := int(data[10] & 0x0F)

	// Extract positioning status from status bits
	positioningStatus := "V" // Default to invalid
	if status != 0 {
		positioningStatus = "A" // Set to active/valid when GPS is positioned
	}

	// Parse latitude, longitude from the data
	latitude := helpers.ParseCoordinate(data[11:15], false)
	longitude := helpers.ParseCoordinate(data[15:19], true)

	// Parse speed
	speed := int(data[19])

	// Extract and print the raw course/status bytes for debugging
	courseStatusBytes := data[20:22]
	courseStatus := int(courseStatusBytes[0])<<8 | int(courseStatusBytes[1])

	// The course/direction is in the lower 10 bits (0-9)
	// But in GT06 protocol, often the course can be in bytes directly without bit masking
	// Let's use a better approach to extract direction
	var direction int

	// Method 1: Use bit masking as per protocol doc (bits 0-9)
	direction = courseStatus & 0x03FF

	// Method 2: If direction is 0, try using the first byte directly (common implementation)
	if direction == 0 && courseStatusBytes[0] > 0 {
		direction = int(courseStatusBytes[0])
	}

	// Method 3: If still 0 and we have status information suggesting we're moving
	// Calculate a reasonable value based on status
	if direction == 0 && courseStatusBytes[0] > 0 && speed > 0 {
		// Some implementations store direction directly in the first byte
		// when speed > 0
		direction = int(courseStatusBytes[0])
	}

	// Extract status information (bits 10-15)
	statusBits := (courseStatus >> 10) & 0x3F

	// Fix: Check GPS positioning bit (bit 4 of statusBits) to determine positioningStatus
	// According to GT06 protocol (5.2.1.9), bit 4 indicates GPS has been positioned or not
	gpsPositioning := (statusBits & 0x10) > 0
	if gpsPositioning {
		positioningStatus = "A" // Set to active/valid when GPS positioning is true
	}

	// Detailed debugging of Course/Status bytes
	if verbose {
		fmt.Printf("\n===== COURSE/STATUS FIELD DETAILS =====\n")
		fmt.Printf("Raw Course/Status bytes: 0x%02X 0x%02X\n", courseStatusBytes[0], courseStatusBytes[1])
		fmt.Printf("Combined Course/Status value: 0x%04X (%d decimal)\n", courseStatus, courseStatus)
		fmt.Printf("Extracted Direction (bits 0-9): %d degrees\n", direction)
		fmt.Printf("Extracted Status bits (bits 10-15): 0x%02X\n", statusBits)
		fmt.Printf("Positioning Status: %s\n", positioningStatus)
		fmt.Printf("Number of Satellites: %d\n", satellites)

		// Parse individual status bits
		gpsPositioning := (statusBits & 0x10) > 0
		eastWest := (statusBits & 0x04) > 0
		northSouth := (statusBits & 0x02) > 0

		fmt.Printf("GPS Positioning: %v\n", gpsPositioning)

		// Fix: Use Go's if/else instead of ternary operator
		longitudeDirection := "East"
		if eastWest {
			longitudeDirection = "West"
		}
		fmt.Printf("Longitude: %s\n", longitudeDirection)

		// Fix: Use Go's if/else instead of ternary operator
		latitudeDirection := "North"
		if northSouth {
			latitudeDirection = "South"
		}
		fmt.Printf("Latitude: %s\n", latitudeDirection)

		// Extra debug: print the raw direction byte value
		fmt.Printf("Raw first byte (often used as direction): %d\n", courseStatusBytes[0])
		fmt.Printf("========================================\n")
	}

	mcc, err := helpers.BytesToHexAndDecimal(data[22:24]) // Using proper range for MCC (2 bytes)
	if err != nil {
		return nil, fmt.Errorf("error mcc %v", err)
	}
	mnc := int(data[24])

	// Fix for LAC and CellID extraction
	lacBytes := data[25:27] // LAC should be 2 bytes

	// CellID is 3 bytes according to the manual spec
	var cellIDBytes []byte
	if len(data) >= 30 {
		cellIDBytes = data[27:30] // CellID is 3 bytes

		// Debug the raw cell ID bytes for troubleshooting
		if verbose {
			fmt.Printf("DEBUG: Raw CellID bytes as hex: %X\n", cellIDBytes)

			// Convert the bytes to an integer value for further debugging
			cellIDValue := int(cellIDBytes[0])<<16 | int(cellIDBytes[1])<<8 | int(cellIDBytes[2])
			fmt.Printf("DEBUG: CellID as decimal: %d\n", cellIDValue)

			// If CellID appears to be 0, double check the packet structure
			if cellIDValue == 0 {
				fmt.Printf("DEBUG: CellID is zero. This might be correct if no cell tower info or check bytes 27-29: %X\n",
					data[27:min(len(data), 30)])
			}
		}
	} else {
		// Fallback in case data is too short
		cellIDBytes = make([]byte, 3)
		if len(data) > 27 {
			copy(cellIDBytes, data[27:min(len(data), 30)])
		}
	}

	lac := hex.EncodeToString(lacBytes)
	cellID := hex.EncodeToString(cellIDBytes)

	// Trim CellID to 4 characters
	if len(cellID) > 4 {
		cellID = cellID[len(cellID)-4:]
	}

	// Debug information for troubleshooting
	if verbose {
		fmt.Printf("DEBUG: Raw LAC bytes: %X, Raw CellID bytes: %X\n", lacBytes, cellIDBytes)
		fmt.Printf("DEBUG: Hex LAC: %s, Hex CellID (trimmed): %s\n", lac, cellID)
		fmt.Printf("DEBUG: Data length: %d, expected CellID at bytes 27-29\n", len(data))
	}

	baseStationInfo := map[string]interface{}{
		"mmc":    fmt.Sprintf("%v", mcc),
		"mnc":    fmt.Sprintf("%v", mnc),
		"lac":    lac,
		"cellId": cellID,
	}

	// Also include individual fields for ease of access by Jono
	// This ensures compatibility with both formats
	mccStr := fmt.Sprintf("%v", mcc)
	mncStr := fmt.Sprintf("%v", mnc)

	// For alarm packets, return the basic model
	if isAlarm {
		return &models.LocationPacketModel{
			IMEI:               imei,
			EventCode:          "50",
			DateTime:           dateTime,
			NumberOfSatellites: satellites,
			PositioningStatus:  positioningStatus,
			Latitude:           latitude,
			Longitude:          longitude,
			Speed:              speed,
			Course:             courseStatus, // Store the full course/status value
			Direction:          direction,    // Store just the direction/bearing (0-359 degrees)
			MCC:                mccStr,       // Add direct MCC field
			MNC:                mncStr,       // Add direct MNC field
			LAC:                lac,          // Add direct LAC field
			CellID:             cellID,       // Add direct CellID field
			BaseStationInfo:    baseStationInfo,
			Message:            data,
			Extra:              string(data[29:]),
		}, nil
	}

	// For regular packets, extract extra data
	extra := ""
	batteryLevel := 0 // Default value

	if len(data) > 29 {
		sensorData := data[29:]
		extra = fmt.Sprintf("%X", sensorData)

		// Improved battery level detection
		// GT06 protocol typically includes battery level in status information
		// Let's examine several potential locations for battery info

		// Check if we have enough data for our heuristics
		if len(sensorData) >= 4 {
			// Use data from sensor byte 3 - this appears to be common for battery info
			// Scale from 0-255 to 0-6 range for battery level
			if sensorData[3] > 0 {
				batteryLevel = int(sensorData[3] * 6 / 255)
				if batteryLevel > 6 {
					batteryLevel = 6 // Cap at max value
				}
			}

			// Alternative: If we have a voltage level in byte 2 for alarm packets
			// This is a backup approach
			if batteryLevel == 0 && len(data) > 30 {
				voltageByte := data[30] % 7 // Ensure it's in 0-6 range
				if voltageByte > 0 {
					batteryLevel = int(voltageByte)
				}
			}

			// If we still have 0, set a minimum level of 1 if we have any data
			if batteryLevel == 0 && len(sensorData) > 0 {
				batteryLevel = 1
			}
		}

		fmt.Printf("DEBUG: Detected battery level: %d from sensor data: %X\n", batteryLevel, sensorData)
	}

	// For alarm packets (0x16), the Terminal Information Content, Voltage Level, GSM signal strength
	// and Alarm/Language are at indices 29, 30, 31 and 32-33 respectively
	if isAlarm && len(data) > 31 {
		// For alarm packets (protocol 0x16), explicitly get voltage level from data[30] as per section 5.3.1.15
		// This is already scaled 0-6
		batteryLevel = int(data[30])
		if batteryLevel < 0 || batteryLevel > 6 {
			// Ensure we have a valid level in the specified range
			batteryLevel = 3 // Default to "Low Battery (can be used normally)"
		}

		fmt.Printf("DEBUG: Alarm packet - battery level directly read: %d\n", batteryLevel)

		// Also extract GSM Signal Strength for alarm packets
		gsmSignalValue := int(data[31])

		fmt.Printf("DEBUG: Alarm packet - GSM signal strength directly read: %d\n", gsmSignalValue)

		// Get the event code based on whether it's an alarm or not
		eventCode := "1" // SOS/Alarm for alarm packets

		return &models.LocationPacketModel{
			IMEI:               imei,
			EventCode:          eventCode,
			DateTime:           dateTime,
			NumberOfSatellites: satellites,
			PositioningStatus:  positioningStatus,
			Latitude:           latitude,
			Longitude:          longitude,
			Speed:              speed,
			Course:             courseStatus, // Store the full course/status value
			Direction:          direction,    // Store just the direction/bearing (0-359 degrees)
			MCC:                mccStr,
			MNC:                mncStr,
			LAC:                lac,
			CellID:             cellID,
			BaseStationInfo:    baseStationInfo,
			Message:            data,
			Extra:              extra,
			BatteryLevel:       batteryLevel,
			GSMSignalStrength:  gsmSignalValue,
		}, nil
	} else {
		// For standard location packets (protocol 0x12), we need to analyze the extended data
		// The format doesn't explicitly define the voltage, so we need to use heuristics

		// We can detect the voltage based on sensor data if it's available
		sensorData := data[29:]
		extra := fmt.Sprintf("%X", sensorData)

		// Look for specific patterns in the extra data that might indicate battery level
		// Based on the observed data packet with sensor data "000005CF780D0A",
		// we need better logic to extract voltage

		// Option 1: If we have enough data, use a specific index as voltage indicator
		// This assumes the voltage data follows a specific pattern in the sensor data
		if len(sensorData) >= 6 {
			// Try to extract a voltage indicator from the first few bytes of sensor data
			// Scale from 0-255 to 0-6
			rawVoltageIndicator := sensorData[0] // Using the first byte as an example
			if rawVoltageIndicator > 0 {
				// Convert raw voltage value to the 0-6 scale
				// If rawVoltageIndicator is directly in volts, we divide by 2 and cap at 6
				voltageLevel := int(rawVoltageIndicator / 2)
				if voltageLevel > 6 {
					voltageLevel = 6
				}
				batteryLevel = voltageLevel
			}
		}

		// If we couldn't determine battery level above, use a fallback method
		if batteryLevel == 0 && len(sensorData) > 0 {
			// Default method: Check the first several bytes for a non-zero value
			// that might represent voltage and scale it
			for i := 0; i < min(len(sensorData), 4); i++ {
				if sensorData[i] > 0 {
					// Convert to 0-6 scale
					scaledValue := int(sensorData[i] * 6 / 255)
					if scaledValue > 0 {
						batteryLevel = scaledValue
						if batteryLevel > 6 {
							batteryLevel = 6
						}
						break
					}
				}
			}
		}

		// If we still didn't find a battery level, set a sensible default
		if batteryLevel == 0 {
			batteryLevel = 4 // Medium level as default
		}

		fmt.Printf("DEBUG: Location packet - derived battery level: %d from sensor data: %s\n", batteryLevel, extra)
	}
	log.Println("Battery level voltageValue:", batteryLevel)
	// Calculate voltage value from battery level
	var voltageValue float64 = 0.0
	switch batteryLevel {
	case 0: // No Power (shutdown)
		voltageValue = 0.0
	case 1: // Extremely Low Battery
		voltageCalc := (3.0 * 1024.0) / 6.0
		voltageValue = float64(int(math.Round(voltageCalc)))
	case 2: // Very Low Battery (Low Battery Alarm)
		voltageCalc := (6.0 * 1024.0) / 6.0
		voltageValue = float64(int(math.Round(voltageCalc)))
	case 3: // Low Battery (can be used normally)
		voltageCalc := (9.0 * 1024.0) / 6.0
		voltageValue = float64(int(math.Round(voltageCalc)))
	case 4: // Medium
		voltageCalc := (12.0 * 1024.0) / 6.0
		voltageValue = float64(int(math.Round(voltageCalc)))
	case 5: // High
		voltageCalc := (12.5 * 1024.0) / 6.0
		voltageValue = float64(int(math.Round(voltageCalc)))
	case 6: // Very High
		voltageCalc := (13.0 * 1024.0) / 6.0
		voltageValue = float64(int(math.Round(voltageCalc)))
	default:
		voltageValue = 9.0 // Default to a safe value
	}

	// Get the event code based on whether it's an alarm or not
	eventCode := "35" // Default Normal Location

	// For alarm packets (0x16), check alarm type
	if isAlarm && len(data) > 5 {
		// Extract alarm info from the alarm data bytes
		alarmInfo := parseAlarmAndLanguage(data[4:6])
		if ec, exists := alarmInfo["EventCode"]; exists {
			eventCode = ec
		}
	} else if len(data) > 4 {
		// For standard packets, check terminal info for potential alarms
		terminalInfoByte := data[4]

		// Check if there's an alarm condition in the terminal info
		eventCodeInt := GetEventCodeFromTerminalInfo(terminalInfoByte)
		if eventCodeInt != 35 { // If not normal
			eventCode = fmt.Sprintf("%d", eventCodeInt)
		}
	}

	// Try to extract GSM Signal Strength for standard location packets
	// This may be in extended data or sensor data
	gsmSignalValue := 2 // Default to medium signal strength

	if len(data) > 29 {
		sensorData := data[29:]
		if len(sensorData) >= 2 {
			// Signal strength is often in the first few bytes
			gsmSignalValue = int(sensorData[1]) % 5 // Ensure 0-4 range
		}
	}

	return &models.LocationPacketModel{
		IMEI:               imei,
		EventCode:          eventCode, // Using our derived eventCode
		DateTime:           dateTime,
		NumberOfSatellites: satellites,
		PositioningStatus:  positioningStatus,
		Latitude:           latitude,
		Longitude:          longitude,
		Speed:              speed,
		Course:             courseStatus, // Store the full course/status value
		Direction:          direction,    // Store just the direction/bearing (0-359 degrees)
		MCC:                mccStr,       // Add direct MCC field
		MNC:                mncStr,       // Add direct MNC field
		LAC:                lac,          // Add direct LAC field
		CellID:             cellID,       // Add direct CellID field
		BaseStationInfo:    baseStationInfo,
		Message:            data,
		Extra:              extra,
		BatteryLevel:       batteryLevel,
		GSMSignalStrength:  gsmSignalValue,
		VoltageValue:       voltageValue, // Include the calculated voltage value
	}, nil
}

// Decode the terminal information decoding function to match spec exactly
func DecodeTerminalInformationBits(terminalInfoByte byte) (bool, bool, int, bool, bool, bool) {
	// Bit 7: Oil and electricity status
	oilDisconnected := (terminalInfoByte & 0x80) != 0 // 1000 0000

	// Bit 6: GPS tracking status
	gpsTrackingOn := (terminalInfoByte & 0x40) != 0 // 0100 0000

	// Bits 5-3: Alarm status (extract and right-shift to normalize)
	alarmBits := (terminalInfoByte >> 3) & 0x07 // Extract bits 5,4,3

	// Convert alarm bits to event code
	var eventCode int
	switch alarmBits {
	case 0b100: // Binary 100
		eventCode = 1 // SOS
	case 0b011: // Binary 011
		eventCode = 1 // Low Battery Alarm (using general alarm code)
	case 0b010: // Binary 010
		eventCode = 23 // Power Cut Alarm
	case 0b001: // Binary 001
		eventCode = 79 // Shock Alarm
	case 0b000: // Binary 000
		eventCode = 35 // Normal
	default:
		eventCode = 35 // Default to Normal for unknown patterns
	}

	// Bit 2: Charge status
	chargeOn := (terminalInfoByte & 0x04) != 0 // 0000 0100

	// Bit 1: ACC status
	accHigh := (terminalInfoByte & 0x02) != 0 // 0000 0010

	// Bit 0: Activation status
	activated := (terminalInfoByte & 0x01) != 0 // 0000 0001

	return oilDisconnected, gpsTrackingOn, eventCode, chargeOn, accHigh, activated
}

func DecodeHeartbeatPacket(data []byte, imei string) (*models.StatusPacketModel, error) {
	if len(data) < 13 { // Verify minimum length: 2+1+1+1+1+1+1+2+2+2
		return nil, fmt.Errorf("error: heartbeat data too short (got %d bytes, need at least 13)", len(data))
	}

	// Extract terminal information byte (byte index 4 after start+length+protocol)
	terminalInfoByte := data[4]
	oilDisconnected, gpsTracking, eventCode, charging, accHigh, activated := DecodeTerminalInformationBits(terminalInfoByte)

	// Create detailed status description
	terminalInfo := fmt.Sprintf(
		"Oil/Electricity: %s, GPS: %s, Event: %d, Charging: %v, ACC: %s, %s",
		map[bool]string{true: "Disconnected", false: "Connected"}[oilDisconnected],
		map[bool]string{true: "On", false: "Off"}[gpsTracking],
		eventCode,
		charging,
		map[bool]string{true: "High", false: "Low"}[accHigh],
		map[bool]string{true: "Activated", false: "Deactivated"}[activated],
	)

	// The rest of the packet decoding remains similar but with verified offsets
	voltageLevelByte := data[5]
	gsmSignalStrengthByte := data[6]
	alarmLanguageBytes := data[7:9]

	if verbose {
		fmt.Printf("Terminal Info Byte: 0x%02X\n", terminalInfoByte)
		fmt.Printf("Decoded Terminal Info: %s\n", terminalInfo)
		fmt.Printf("Event Code: %d\n", eventCode)
	}

	// Print debug information for Terminal Information Content
	fmt.Printf("\n===== HEARTBEAT PACKET DEBUG INFO =====\n")
	fmt.Printf("Raw data: %X\n", data)
	fmt.Printf("Terminal Information byte: 0x%02X (binary: %08b)\n", terminalInfoByte, terminalInfoByte)
	fmt.Printf("Voltage Level byte: 0x%02X (level %d of 6)\n", voltageLevelByte, voltageLevelByte)
	fmt.Printf("GSM Signal Strength byte: 0x%02X (level %d of 4)\n", gsmSignalStrengthByte, gsmSignalStrengthByte)
	fmt.Printf("Alarm/Language bytes: %X\n", alarmLanguageBytes)

	// Decode Terminal Information using the enhanced version of parseTerminalInformationContent
	terminalInfo = parseTerminalInformationContent(terminalInfoByte)
	fmt.Printf("\nDecoded Terminal Information: %s\n", terminalInfo)

	// Parse voltage level
	voltageLevel := parseVoltageLevel(voltageLevelByte)
	fmt.Printf("Decoded Voltage Level: %s\n", voltageLevel)

	// Parse GSM signal strength
	gsmSignalStrength := parseGSMSignalStrength(gsmSignalStrengthByte)
	fmt.Printf("Decoded GSM Signal Strength: %s\n", gsmSignalStrength)

	// Parse alarm and language
	alarmLanguage := parseAlarmAndLanguage(alarmLanguageBytes)
	fmt.Printf("Decoded Alarm/Language: %v\n", alarmLanguage)
	fmt.Printf("========================================\n\n")

	// Add more detailed voltage information for better AD4 conversion later
	voltageFloat := float64(0)
	volt_value := 12.0
	switch voltageLevelByte {
	case 0:
		voltageFloat = 0.0 // No Power
	case 1:
		volt_value := 3.0
		voltageCalc := (volt_value * 1024.0) / 6.0
		voltageFloat = float64(int(math.Round(voltageCalc)))
	case 2:
		volt_value := 6.0
		voltageCalc := (volt_value * 1024.0) / 6.0
		voltageFloat = float64(int(math.Round(voltageCalc)))
	case 3:
		volt_value := 9.0
		voltageCalc := (volt_value * 1024.0) / 6.0
		voltageFloat = float64(int(math.Round(voltageCalc)))
	case 4:
		volt_value := 12.0
		voltageCalc := (volt_value * 1024.0) / 6.0
		voltageFloat = float64(int(math.Round(voltageCalc)))
	case 5:
		volt_value := 12.5
		voltageCalc := (volt_value * 1024.0) / 6.0
		voltageFloat = float64(int(math.Round(voltageCalc)))
	case 6:
		volt_value := 13.0
		voltageCalc := (volt_value * 1024.0) / 6.0
		voltageFloat = float64(int(math.Round(voltageCalc)))
	default:

		voltageFloat = 9.0 // Default to something reasonable
	}
	fmt.Printf("Voltage Level: %d corresponds to approximately %.1f volts\n", voltageLevelByte, volt_value)

	// Create and return the model
	return &models.StatusPacketModel{
		IMEI:                      imei,
		TerminalInformationByte:   terminalInfoByte,
		TerminalInformationString: terminalInfo,
		VoltageLevelByte:          voltageLevelByte,
		VoltageLevelString:        voltageLevel,
		VoltageValue:              voltageFloat, // Add actual voltage value
		GSMSignalStrengthByte:     gsmSignalStrengthByte,
		GSMSignalStrengthString:   gsmSignalStrength,
		AlarmAndLanguage:          alarmLanguage,
		Message:                   data,
	}, nil
}

// Enhanced version of parseTerminalInformationContent
func parseTerminalInformationContent(terminalInfoByte byte) string {
	// Convert to binary representation
	binaryRepresentation := fmt.Sprintf("%08b", terminalInfoByte)

	// Extract details according to documentation
	oilAndElectricity := "Connected"
	if binaryRepresentation[0] == '1' {
		oilAndElectricity = "Disconnected"
	}

	gpsTracking := "Off"
	if binaryRepresentation[1] == '1' {
		gpsTracking = "On"
	}

	// Extract bits 2-4 (Alarm Status)
	alarmBits := binaryRepresentation[2:5]
	alarmStatus := "Normal"
	alarmCode := "000"
	if alarmBits != "000" {
		switch alarmBits {
		case "001":
			alarmStatus = "Shock Alarm"
		case "010":
			alarmStatus = "Power Cut Alarm"
		case "011":
			alarmStatus = "Low Battery Alarm"
		case "100":
			alarmStatus = "SOS"
		default:
			alarmStatus = "Unknown Alarm (" + alarmBits + ")"
		}
		alarmCode = alarmBits
	}

	chargeStatus := "Off"
	if binaryRepresentation[5] == '1' {
		chargeStatus = "On"
	}

	accStatus := "Low"
	if binaryRepresentation[6] == '1' {
		accStatus = "High"
	}

	activationStatus := "Deactivated"
	if binaryRepresentation[7] == '1' {
		activationStatus = "Activated"
	}

	// Build the formatted string
	return fmt.Sprintf(
		"Oil/Electricity: %s, GPS Tracking: %s, Alarm: %s (code %s), Charge: %s, ACC: %s, Activation: %s",
		oilAndElectricity, gpsTracking, alarmStatus, alarmCode, chargeStatus, accStatus, activationStatus,
	)
}

// Add new helper function to map alarm status from terminal info to EventCode
func GetEventCodeFromTerminalInfo(terminalInfoByte byte) int {
	// Extract bits 2-4 (Alarm Status)
	alarmBits := (terminalInfoByte >> 2) & 0x07

	switch alarmBits {
	case 0:
		return 35 // Normal
	case 1:
		return 79 // Shock Alarm
	case 2:
		return 23 // Power Cut Alarm
	case 3:
		return 50 // Low Battery Alarm (using 50 for general alarm)
	case 4:
		return 1 // SOS
	default:
		return 35 // Default to Normal
	}
}

// Decode alarm and language bytes according to manual section 5.4.1
func parseAlarmAndLanguage(data []byte) map[string]string {
	if len(data) != 2 {
		return map[string]string{
			"Error":     "Invalid Alarm/Language data length",
			"EventCode": "35", // Default to Normal
		}
	}

	// First byte is alarm status
	alarmType := "Normal"
	eventCode := "35"

	switch data[0] {
	case 0x00:
		alarmType = "Normal"
		eventCode = "35"
	case 0x01:
		alarmType = "SOS"
		eventCode = "1"
	case 0x02:
		alarmType = "Power Cut Alarm"
		eventCode = "23"
	case 0x03:
		alarmType = "Shock Alarm"
		eventCode = "79"
	case 0x04:
		alarmType = "Fence In Alarm"
		eventCode = "20"
	case 0x05:
		alarmType = "Fence Out Alarm"
		eventCode = "21"
	}

	// Second byte is language
	language := "Unknown"
	switch data[1] {
	case 0x01:
		language = "Chinese"
	case 0x02:
		language = "English"
	}

	return map[string]string{
		"Alarm":     alarmType,
		"Language":  language,
		"EventCode": eventCode,
	}
}

// Parse voltage level according to manual section 5.4.1
func parseVoltageLevel(voltageByte byte) string {
	switch voltageByte {
	case 0:
		return "No Power (shutdown)"
	case 1:
		return "Extremely Low Battery (not enough for calling or sending text messages)"
	case 2:
		return "Very Low Battery (Low Battery Alarm)"
	case 3:
		return "Low Battery (can be used normally)"
	case 4:
		return "Medium"
	case 5:
		return "High"
	case 6:
		return "Very High"
	default:
		return "Unknown Voltage Level"
	}
}

// Parse GSM signal strength according to manual section 5.4.1
func parseGSMSignalStrength(signalByte byte) string {
	switch signalByte {
	case 0x00:
		return "No signal"
	case 0x01:
		return "Extremely weak signal"
	case 0x02:
		return "Very weak signal"
	case 0x03:
		return "Good signal"
	case 0x04:
		return "Strong signal"
	default:
		return "Unknown GSM Signal Strength"
	}
}

// ExtractIMEI extracts the IMEI from login packet data
func ExtractIMEI(data []byte) (string, error) {
	if len(data) < 10 {
		return "", fmt.Errorf("error login data too short to extract IMEI")
	}
	imeiBytes := data[4:12]
	imei := ""
	for i, b := range imeiBytes {
		digit1 := (b >> 4) & 0x0F
		digit2 := b & 0x0F

		if i == 0 && digit1 == 0 {
			imei += fmt.Sprintf("%d", digit2)
		} else {
			imei += fmt.Sprintf("%d%d", digit1, digit2)
		}
	}

	if len(imei) > 15 {
		imei = imei[:15]
	}

	return imei, nil
}

// DecodeAlarmFrame decodes an alarm packet into an AlarmPacketModel
func DecodeAlarmFrame(data []byte, imei string) (models.AlarmPacketModel, error) {
	if verbose {
		fmt.Printf("Decoding alarm frame. Raw data: %X\n", data)
	}

	// First decode the location data
	locationFrameFields, err := DecodeStandardLocationData(data, imei, true)
	if err != nil {
		return models.AlarmPacketModel{}, fmt.Errorf("error decoding location fields: %v", err)
	}

	// Terminal Information byte is at index 4
	terminalInfoByte := data[4]
	oilDisconnected, gpsTrackingOn, eventCode, chargeOn, accHigh, activated := DecodeTerminalInformationBits(terminalInfoByte)
	fmt.Print(eventCode)
	// Get the event code directly from terminal info bits (bits 3-5)
	alarmBits := (terminalInfoByte >> 3) & 0x07
	var mappedEventCode int
	switch alarmBits {
	case 0b100: // Binary 100
		mappedEventCode = 31 // SOS
	case 0b011: // Binary 011
		mappedEventCode = 23 // Low Battery Alarm
	case 0b010: // Binary 010
		mappedEventCode = 16 // Power Cut Alarm
	case 0b001: // Binary 001
		mappedEventCode = 8 // Shock Alarm
	case 0b000: // Binary 000
		mappedEventCode = 0 // Normal
	default:
		mappedEventCode = 0
	}

	// Set the mapped event code in the location fields
	locationFrameFields.EventCode = fmt.Sprintf("%d", mappedEventCode)

	// Parse GSM signal strength (byte 6) using the same mapping as other packets
	gsmStrengthByte := data[6]
	mappedGSMStrength := 0
	switch gsmStrengthByte {
	case 0x00:
		mappedGSMStrength = 0 // No signal
	case 0x01:
		mappedGSMStrength = 8 // Extremely weak
	case 0x02:
		mappedGSMStrength = 16 // Very weak
	case 0x03:
		mappedGSMStrength = 23 // Good
	case 0x04:
		mappedGSMStrength = 31 // Strong
	}
	locationFrameFields.GSMSignalStrength = mappedGSMStrength

	// Create detailed terminal info description
	terminalInfo := fmt.Sprintf("Oil: %v, GPS: %v, Event: %d, Charge: %v, ACC: %v, Active: %v",
		map[bool]string{true: "Disconnected", false: "Connected"}[oilDisconnected],
		map[bool]string{true: "On", false: "Off"}[gpsTrackingOn],
		mappedEventCode,
		chargeOn,
		map[bool]string{true: "High", false: "Low"}[accHigh],
		map[bool]string{true: "Activated", false: "Deactivated"}[activated])

	// Parse voltage level (byte 5)
	voltageLevel := parseVoltageLevel(data[5])

	// Parse GSM signal strength (byte 6)
	gsmSignalStrength := parseGSMSignalStrength(data[6])

	// Parse Alarm/Language (bytes 7-8)
	alarmLanguage := parseAlarmAndLanguage(data[7:9])
	alarmLanguage["EventCode"] = fmt.Sprintf("%d", mappedEventCode) // Update EventCode in AlarmLanguage

	return models.AlarmPacketModel{
		LocationPacketModel:        locationFrameFields,
		TerminalInformationContent: terminalInfo,
		VoltageLevel:               voltageLevel,
		GSMSignalStrength:          gsmSignalStrength,
		AlarmAndLanguage:           alarmLanguage,
	}, nil
}

// DecodeStringInformationPacket decodes a GT06 string information packet (0x15)
// The packet contains text with location information in a format like:
// "Current position:Lat:N19.521012,Lon:W99.211767,DateTime:2025-04-30 14:10:55,..."
func DecodeStringInformationPacket(data []byte, imei string) (*models.LocationPacketModel, error) {
	if len(data) < 10 {
		return nil, fmt.Errorf("error: string information data too short")
	}

	// Extract the content length
	contentLength := int(data[4])

	// Skip the server flag (4 bytes) and go directly to content
	if len(data) < 9+contentLength-4 {
		return nil, fmt.Errorf("error: content truncated based on specified length")
	}

	// The actual message content starts after the server flag
	messageContent := data[9 : 9+contentLength-4] // Subtract 4 for server flag

	// Convert to string for parsing
	messageStr := string(messageContent)

	// Parse latitude, longitude, and datetime from the text
	var lat, lon float64
	var dateTime string
	var err error

	// Try to extract coordinates and datetime from the message content
	lat, lon, dateTime, err = extractPositionFromString(messageStr)
	if err != nil {
		fmt.Printf("Warning: Could not parse complete position from text message: %v\n", err)
		// We'll still create a location model with whatever we could extract
	}

	// Set a default event code for string information packets
	eventCode := "35" // Normal

	// Create the location model
	locationModel := &models.LocationPacketModel{
		IMEI:               imei,
		EventCode:          eventCode,
		DateTime:           dateTime,
		NumberOfSatellites: 0,   // Not available in string response
		PositioningStatus:  "A", // Assume valid position since we got a response
		Latitude:           lat,
		Longitude:          lon,
		Speed:              0, // Not available in text message
		Direction:          0, // Not available in text message
		Message:            data,
		Extra:              messageStr, // Store the full message as extra data
		BatteryLevel:       0,
		GSMSignalStrength:  0,
	}

	return locationModel, nil
}

// extractPositionFromString parses position information from a text message
// It handles multiple formats that might be used in GT06 string responses
func extractPositionFromString(content string) (float64, float64, string, error) {
	var lat, lon float64
	var dateTime string
	var latErr, lonErr, dtErr error

	// Format pattern from the example: "Current position:Lat:N19.521012,Lon:W99.211767,DateTime:2025-04-30 14:10:55"
	if strings.Contains(content, "Lat:") && strings.Contains(content, "Lon:") {
		// Extract latitude with direction (N/S)
		latMatches := regexp.MustCompile(`Lat:([NS])(\d+\.\d+)`).FindStringSubmatch(content)
		if len(latMatches) >= 3 {
			latVal, err := strconv.ParseFloat(latMatches[2], 64)
			if err == nil {
				// Apply sign based on direction (N positive, S negative)
				if latMatches[1] == "S" {
					lat = -latVal
				} else {
					lat = latVal
				}
			} else {
				latErr = fmt.Errorf("error parsing latitude value: %v", err)
			}
		} else {
			latErr = fmt.Errorf("latitude pattern not found in message")
		}

		// Extract longitude with direction (E/W)
		lonMatches := regexp.MustCompile(`Lon:([EW])(\d+\.\d+)`).FindStringSubmatch(content)
		if len(lonMatches) >= 3 {
			lonVal, err := strconv.ParseFloat(lonMatches[2], 64)
			if err == nil {
				// Apply sign based on direction (E positive, W negative)
				if lonMatches[1] == "W" {
					lon = -lonVal
				} else {
					lon = lonVal
				}
			} else {
				lonErr = fmt.Errorf("error parsing longitude value: %v", err)
			}
		} else {
			lonErr = fmt.Errorf("longitude pattern not found in message")
		}

		// Extract datetime
		dtMatches := regexp.MustCompile(`DateTime:(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`).FindStringSubmatch(content)
		if len(dtMatches) >= 2 {
			// Convert to ISO format
			parsedTime, err := time.Parse("2006-01-02 15:04:05", dtMatches[1])
			if err == nil {
				dateTime = parsedTime.Format(time.RFC3339)
			} else {
				dtErr = fmt.Errorf("error parsing datetime: %v", err)
				// Use original string if parsing fails
				dateTime = dtMatches[1]
			}
		} else {
			// Alternative format with just date and time separately
			dateMatches := regexp.MustCompile(`Date:(\d{4}-\d{2}-\d{2})`).FindStringSubmatch(content)
			timeMatches := regexp.MustCompile(`Time:(\d{2}:\d{2}:\d{2})`).FindStringSubmatch(content)

			if len(dateMatches) >= 2 && len(timeMatches) >= 2 {
				dtStr := dateMatches[1] + " " + timeMatches[1]
				parsedTime, err := time.Parse("2006-01-02 15:04:05", dtStr)
				if err == nil {
					dateTime = parsedTime.Format(time.RFC3339)
				} else {
					dtErr = fmt.Errorf("error parsing separate date/time: %v", err)
					// Use original string if parsing fails
					dateTime = dtStr
				}
			} else {
				dtErr = fmt.Errorf("datetime pattern not found in message")
				// Use current time as fallback
				dateTime = time.Now().Format(time.RFC3339)
			}
		}
	} else if strings.HasPrefix(content, "DWXX=") {
		// Handle the DWXX format described in section 6.3 of the GT06 manual
		// Format: DWXX=t1,t2,MMDDHHMM,lat,long,speed,status,signal
		parts := strings.Split(content[5:], ",")
		if len(parts) >= 5 {
			// Parse latitude (4th element)
			latVal, err := strconv.ParseFloat(parts[3], 64)
			if err == nil {
				lat = latVal
			} else {
				latErr = fmt.Errorf("error parsing DWXX latitude: %v", err)
			}

			// Parse longitude (5th element)
			lonVal, err := strconv.ParseFloat(parts[4], 64)
			if err == nil {
				lon = lonVal
			} else {
				lonErr = fmt.Errorf("error parsing DWXX longitude: %v", err)
			}

			// Parse datetime (3rd element) in format MMDDHHMM
			if len(parts) > 2 && len(parts[2]) == 8 {
				month := parts[2][0:2]
				day := parts[2][2:4]
				hour := parts[2][4:6]
				minute := parts[2][6:8]

				// Use current year for the timestamp
				year := fmt.Sprintf("%d", time.Now().Year())

				// Combine into datetime string
				dtStr := fmt.Sprintf("%s-%s-%s %s:%s:00", year, month, day, hour, minute)
				parsedTime, err := time.Parse("2006-01-02 15:04:05", dtStr)
				if err == nil {
					dateTime = parsedTime.Format(time.RFC3339)
				} else {
					dtErr = fmt.Errorf("error parsing DWXX datetime: %v", err)
					dateTime = time.Now().Format(time.RFC3339)
				}
			} else {
				dtErr = fmt.Errorf("invalid DWXX datetime format")
				dateTime = time.Now().Format(time.RFC3339)
			}
		}
	} else {
		// Try to extract any coordinates in general format
		latLonMatches := regexp.MustCompile(`([-+]?\d+\.\d+)\s*,\s*([-+]?\d+\.\d+)`).FindStringSubmatch(content)
		if len(latLonMatches) >= 3 {
			lat, latErr = strconv.ParseFloat(latLonMatches[1], 64)
			lon, lonErr = strconv.ParseFloat(latLonMatches[2], 64)
		}

		// Try to extract any datetime in general format
		dateTimeMatches := regexp.MustCompile(`\d{4}[-/]\d{2}[-/]\d{2}\s+\d{2}:\d{2}:\d{2}`).FindString(content)
		if dateTimeMatches != "" {
			// Try multiple formats
			formats := []string{"2006-01-02 15:04:05", "2006/01/02 15:04:05"}
			for _, format := range formats {
				if parsedTime, err := time.Parse(format, dateTimeMatches); err == nil {
					dateTime = parsedTime.Format(time.RFC3339)
					dtErr = nil
					break
				}
			}
			if dateTime == "" {
				dtErr = fmt.Errorf("could not parse datetime in any known format")
				dateTime = time.Now().Format(time.RFC3339)
			}
		} else {
			dtErr = fmt.Errorf("no datetime found in content")
			dateTime = time.Now().Format(time.RFC3339)
		}
	}

	// Collect all errors
	var errMsgs []string
	if latErr != nil {
		errMsgs = append(errMsgs, latErr.Error())
	}
	if lonErr != nil {
		errMsgs = append(errMsgs, lonErr.Error())
	}
	if dtErr != nil {
		errMsgs = append(errMsgs, dtErr.Error())
	}

	// If we have any errors and couldn't extract coordinates, return them
	if (lat == 0 && lon == 0) && len(errMsgs) > 0 {
		return lat, lon, dateTime, fmt.Errorf("parsing errors: %s", strings.Join(errMsgs, "; "))
	}

	return lat, lon, dateTime, nil
}
