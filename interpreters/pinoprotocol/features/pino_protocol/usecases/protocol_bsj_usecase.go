package usecases

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings" // Add this import for strings.Repeat
	"time"

	"github.com/MaddSystems/jonobridge/common/utils"
	"golang.org/x/exp/rand"
)

// Add this global variable at the top of the file
var verbose = true // Set to true for detailed debug output

func ParseLocationData(data []byte, imei string, original []byte) string {
	if len(data) < 28 { // Asegurarse de que hay datos suficientes
		utils.VPrint("Trama de localización inválida: longitud insuficiente")
		return ""
	}

	utils.VPrint("alarmSign slice: %X", data[:4])     // Imprime el slice en formato hexadecimal
	_, alarmSign, _ := BytesToHexAndDecimal(data[:4]) // DWORD (4 bytes)

	utils.VPrint("status slice: %X", data[4:8]) // Imprime el slice en formato hexadecimal
	status := data[4:8]                         // DWORD (4 bytes)

	// Convert status to integer for bit checking
	statusInt := bytesToInt(status)
	// Bit 2: 0 = North latitude, 1 = South latitude
	isNorthLatitude := (statusInt & (1 << 2)) == 0
	// Bit 3: 0 = East longitude, 1 = West longitude
	isEastLongitude := (statusInt & (1 << 3)) == 0

	utils.VPrint("statusInt: %d, North Latitude: %v, East Longitude: %v",
		statusInt, isNorthLatitude, isEastLongitude)

	utils.VPrint("latitude slice: %X", data[8:12]) // Imprime el slice en formato hexadecimal

	// Calculate raw latitude value
	latitudeInt := bytesToInt(data[8:12])
	latitude := float64(latitudeInt) / 1000000.0 // DWORD (4 bytes)

	// Apply south latitude if indicated in status
	if !isNorthLatitude {
		latitude = -latitude
	}

	// Calculate raw longitude value
	longitudeInt := bytesToInt(data[12:16])
	longitude := float64(longitudeInt) / 1000000.0 // DWORD (4 bytes)

	// Apply west longitude if indicated in status
	if !isEastLongitude {
		longitude = -longitude
	}

	utils.VPrint("Raw latitude integer: %d, Raw longitude integer: %d", latitudeInt, longitudeInt)
	utils.VPrint("Final calculated coordinates: Latitude=%f, Longitude=%f", latitude, longitude)

	// Use the longitude without converting through string
	longitudeFloat := longitude

	// Fix elevation/altitude (bytes 16-18, WORD value in meters)
	utils.VPrint("elevation slice: %X", data[16:18])
	elevation := bytesToInt(data[16:18]) // WORD (2 bytes) for altitude in meters

	// Fix speed (bytes 18-20, WORD value in 1/10 km/h)
	utils.VPrint("speed slice: %X", data[18:20])
	speed := bytesToInt(data[18:20])

	// Fix direction (bytes 20-22, WORD value 0-359° where 0=true north)
	utils.VPrint("direction slice: %X", data[20:22])
	direction := bytesToInt(data[20:22]) // WORD (2 bytes) for direction

	utils.VPrint("extendedData slice: %X", data[29:]) // Imprime el slice en formato hexadecimal
	extendedData := fetchLongIdsExtendedData(data[29:])
	extendedData = formattedExtendedData(extendedData)
	timeFormatted, err := FormatToISO8601(DecodeBCD(data[22:28]))
	if err != nil {
		timeFormatted, _ = FormatToISO8601(DecodeBCD(data[23:29]))
	}

	mileage := 0
	if mileageStr, exists := extendedData["Mileage"]; exists {
		// Fix: Handle both string and int types for mileage
		switch v := mileageStr.(type) {
		case string:
			if mileageInt, err := hexToIntMileage(v); err == nil {
				mileage = mileageInt / 10 // Convert to km from 1/10 km
			}
		case int:
			mileage = v / 10 // Convert to km from 1/10 km
		case *big.Int:
			mileage = int(v.Int64()) / 10
		default:
			utils.VPrint("Warning: Unexpected mileage type: %T", mileageStr)
		}
	}

	// Debug the satellites info from extended data
	if satValue, exists := extendedData["NumberOfSatellites"]; exists {
		utils.VPrint("DEBUG BSJ: NumberOfSatellites from extended data: %v (type: %T)",
			satValue, satValue)
	} else {
		utils.VPrint("DEBUG BSJ: NumberOfSatellites not found in extended data")
	}

	// Debug the IMEI parameter
	utils.VPrint("DEBUG BSJ: Using IMEI from parameter: %s", imei)

	// Add IMEI to the map before merging extendedData to ensure it's preserved
	locationData := map[string]interface{}{
		"AlarmSign":          alarmSign,
		"Status":             fmt.Sprintf("%X", status),
		"Latitude":           latitude,
		"Longitude":          longitudeFloat,
		"Datetime":           timeFormatted,
		"Speed":              float64(speed) / 10.0, // Convertir a km/h
		"Direction":          direction,             // Direction in degrees (0-359)
		"Elevation":          elevation,             // Altitude in meters
		"Altitude":           elevation,             // Also include as Altitude for compatibility
		"Mileage":            mileage,               // Mileage in km
		"Message":            original,
		"NumberOfSatellites": extendedData["NumberOfSatellites"], // Ensure this is included
		"IMEI":               imei,                               // Add the IMEI to ensure it's included in the output
	}

	// Copy extendedData to locationData, but prevent overwriting existing values
	for k, v := range extendedData {
		// Skip if key is "IMEI", as we want to preserve the parameter value
		if k == "IMEI" {
			utils.VPrint("DEBUG BSJ: Skipping extendedData IMEI to preserve parameter IMEI: %s", imei)
			continue
		}
		locationData[k] = v
	}

	// Final debug check for IMEI
	utils.VPrint("DEBUG BSJ: Final IMEI in locationData: %v", locationData["IMEI"])

	jsonData, err := json.Marshal(locationData)

	// Final debug check
	utils.VPrint("DEBUG BSJ: Final NumberOfSatellites in locationData: %v",
		locationData["NumberOfSatellites"])

	if err != nil {
		log.Printf("Error al convertir datos a JSON: %v", err)
		return ""
	}

	return string(jsonData)

}

func bytesToInt(data []byte) int {
	var result int
	for _, b := range data {
		result = (result << 8) | int(b)
	}
	return result
}

func parseExtendedData(data []byte) map[string]interface{} {
	extendedData := make(map[string]interface{})
	offset := 0

	for offset < len(data) {
		if offset+2 > len(data) {
			log.Printf("Datos extendidos inválidos: offset fuera de rango")
			break
		}

		id := data[offset]            // ID
		length := int(data[offset+1]) // Longitud
		offset += 2

		if offset+length > len(data) {
			log.Printf("Datos extendidos inválidos: longitud fuera de rango para ID %X", id)
			break
		}

		value := data[offset : offset+length]
		offset += length

		// Decodificar valores según ID
		switch id {
		case 0x01:
			_, extendedData["Mileage"], _ = BytesToHexAndDecimal(value[:4])
		case 0x30:
			extendedData["GsmSignalStrength"] = int(value[0])
		case 0x31:
			// Fix: For satellites, use only the lower 4 bits as per common GNSS implementations
			satellites := int(value[0] & 0x0F)
			utils.VPrint("DEBUG BSJ: Raw satellites byte: 0x%02X, Parsed value: %d", value[0], satellites)
			extendedData["NumberOfSatellites"] = satellites
		case 0xEB: // Código de evento o estado adicional
			extendedData["ExtendedData"] = value

		default:
			extendedData[fmt.Sprintf("UnknownID_%X", id)] = hex.EncodeToString(value)
		}
	}

	if extendedValue, ok := extendedData["AlarmStatus"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["AlarmStatus"] = parseAlarmStatus(stringValue)
		}
	}

	if extendedValue, ok := extendedData["BaseStation"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["BaseStation"] = parseBaseStation(stringValue)
		}
	}

	if extendedValue, ok := extendedData["AlarmStatus2"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["AlarmStatus2"] = parseAlarmStatus2(stringValue)
		}
	}

	if extendedValue, ok := extendedData["ExternalVoltage"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["ExternalVoltage"], _ = hexToInt(stringValue)
		}
	}

	if extendedValue, ok := extendedData["IMEI"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["IMEI"], _ = hexToUTF8(stringValue)
		}
	}

	if extendedValue, ok := extendedData["WiFi"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["WiFi"], _ = hexToUTF8(stringValue)
		}
	}

	return extendedData
}

func formattedExtendedData(extendedData map[string]interface{}) map[string]interface{} {
	if extendedValue, ok := extendedData["AlarmStatus"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["AlarmStatus"] = parseAlarmStatus(stringValue)
		}
	}

	if extendedValue, ok := extendedData["BaseStation"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["BaseStation"] = parseBaseStation(stringValue)
		}
	}

	if extendedValue, ok := extendedData["AlarmStatus2"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["AlarmStatus2"] = parseAlarmStatus2(stringValue)
		}
	}

	if extendedValue, ok := extendedData["ExternalVoltage"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["ExternalVoltage"], _ = hexToInt(stringValue)
		}
	}

	if extendedValue, ok := extendedData["IMEI"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["IMEI"], _ = hexToUTF8(stringValue)
		}
	}

	if extendedValue, ok := extendedData["WiFi"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["WiFi"], _ = hexToUTF8(stringValue)
		}
	}
	if extendedValue, ok := extendedData["Mileage"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["Mileage"], _ = hexToIntMileage(stringValue)
		}
	}
	if extendedValue, ok := extendedData["GsmSignalStrength"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["GsmSignalStrength"], _ = hexToInt(stringValue)
		}
	}
	if extendedValue, ok := extendedData["NumberOfSatellites"]; ok {
		if stringValue, valid := extendedValue.(string); valid {
			extendedData["NumberOfSatellites"], _ = hexToInt(stringValue)
		}
	}
	return extendedData
}
func hexToInt(hexString string) (int, error) {
	// Convertir el string hexadecimal a un número entero
	value, err := strconv.ParseInt(hexString, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting hex to int: %v", err)
	}
	return int(value), nil
}
func parseAlarmStatus2(alarmHex string) map[string]interface{} {
	// Convertir el valor hexadecimal a un número entero (32 bits)
	alarmValue, err := strconv.ParseUint(alarmHex, 16, 32)
	if err != nil {
		return map[string]interface{}{
			"Error": fmt.Sprintf("Error parsing alarmHex: %v", err),
		}
	}

	// Definir el mapa de alarmas según los bits
	alarmStatus := make(map[string]interface{})

	// Bits 3~4: GPS positioning
	bit34 := (alarmValue >> 3) & 0b11
	switch bit34 {
	case 0b00:
		alarmStatus["Positioning"] = "No positioning"
	case 0b10:
		alarmStatus["Positioning"] = "GPS positioning"
	case 0b11:
		alarmStatus["Positioning"] = "Default"
	default:
		alarmStatus["Positioning"] = "Unknown positioning state"
	}

	// Bit 6: Vibration alarm
	if (alarmValue & (1 << 6)) == 0 {
		alarmStatus["VibrationAlarm"] = "Active"
	} else {
		alarmStatus["VibrationAlarm"] = "Normal"
	}

	// Bit 14: Alarm when exposed to light
	if (alarmValue & (1 << 14)) == 0 {
		alarmStatus["LightExposureAlarm"] = "Active"
	} else {
		alarmStatus["LightExposureAlarm"] = "Normal"
	}

	return alarmStatus
}

func parseBaseStation(baseStationHex string) map[string]interface{} {
	// Decodificar el string hexadecimal a bytes
	decodedBytes, err := hex.DecodeString(baseStationHex)
	if err != nil {
		return map[string]interface{}{
			"Error": fmt.Sprintf("Error decoding baseStationHex: %v", err),
		}
	}

	// Verificar que la longitud sea suficiente (2 + 1 + 2 + 4 bytes = 9 bytes)
	if len(decodedBytes) < 9 {
		return map[string]interface{}{
			"Error": "Invalid length for base station data. Expected at least 9 bytes.",
		}
	}

	// Extraer los campos según la documentación
	countryCode := int(decodedBytes[0])<<8 | int(decodedBytes[1]) // 2 bytes
	operatorNumber := int(decodedBytes[2])                        // 1 byte
	areaCode := int(decodedBytes[3])<<8 | int(decodedBytes[4])    // 2 bytes
	towerNumber := int(decodedBytes[5])<<24 | int(decodedBytes[6])<<16 |
		int(decodedBytes[7])<<8 | int(decodedBytes[8]) // 4 bytes

	// Retornar el resultado en un mapa
	return map[string]interface{}{
		"CountryCode":    countryCode,    // Código del país
		"OperatorNumber": operatorNumber, // Número del operador
		"AreaCode":       areaCode,       // Código del área
		"TowerNumber":    towerNumber,    // Número de torre
	}
}
func parseAlarmStatus(alarmHex string) map[string]string {
	// Convertir el valor hexadecimal a un número entero (32 bits)
	alarmValue, err := strconv.ParseUint(alarmHex, 16, 32)
	if err != nil {
		return map[string]string{
			"Error": fmt.Sprintf("Error parsing alarmHex: %v", err),
		}
	}

	// Si el valor es FFFFFFFF, todas las alarmas están en estado normal
	if alarmHex == "FFFFFFFF" {
		return map[string]string{
			"BatterySwitch":          "Normal",
			"TerminalState":          "Normal",
			"CollisionAlarm":         "Normal",
			"RapidAccelerationAlarm": "Normal",
			"RapidDecelerationAlarm": "Normal",
			"IllegalRemoval":         "Normal",
			"SharpTurnAlarm":         "Normal",
			"PseudoBaseStation":      "Normal",
		}
	}

	// Definir las alarmas y sus bits asociados
	alarms := map[int]string{
		0:  "BatterySwitch",          // Bit 0
		1:  "TerminalState",          // Bit 1
		4:  "CollisionAlarm",         // Bit 4
		8:  "RapidAccelerationAlarm", // Bit 8
		9:  "RapidDecelerationAlarm", // Bit 9
		12: "IllegalRemoval",         // Bit 12
		25: "SharpTurnAlarm",         // Bit 25
		30: "PseudoBaseStation",      // Bit 30
		31: "PseudoBaseStationAlarm", // Bit 31
	}

	// Crear el mapa con el estado de las alarmas
	alarmStatus := make(map[string]string)
	for bit, name := range alarms {
		// Verificar si el bit está activo (estado 0) o normal (estado 1)
		if (alarmValue & (1 << bit)) == 0 {
			alarmStatus[name] = "Active" // Estado de alarma
		} else {
			alarmStatus[name] = "Normal" // Estado normal
		}
	}

	return alarmStatus
}

func fetchLongIdsExtendedData(data []byte) map[string]interface{} {
	extendedData := make(map[string]interface{})

	// Debug the complete data
	utils.VPrint("DEBUG BSJ: Raw extended data bytes: %X", data)

	// Direct parsing approach - scan for specific ID patterns first
	// Look for satellite count (0x31) and mileage (0x01) directly in the data
	for i := 0; i < len(data)-2; i++ {
		// Search for NumberOfSatellites (ID 0x31)
		if data[i] == 0x31 && i+2 < len(data) {
			length := int(data[i+1])
			if i+2+length <= len(data) && length > 0 {
				satellites := int(data[i+2])
				utils.VPrint("DEBUG BSJ: Found NumberOfSatellites ID 0x31 at position %d, value: 0x%02X (%d)",
					i, satellites, satellites)
				extendedData["NumberOfSatellites"] = satellites
			}
		}

		// Search for Mileage (ID 0x01)
		if data[i] == 0x01 && i+5 < len(data) {
			length := int(data[i+1])
			if i+2+length <= len(data) && length >= 4 {
				mileageBytes := data[i+2 : i+6] // Read 4 bytes for mileage
				mileageInt := int(mileageBytes[0])<<24 | int(mileageBytes[1])<<16 |
					int(mileageBytes[2])<<8 | int(mileageBytes[3])
				utils.VPrint("DEBUG BSJ: Found Mileage ID 0x01 at position %d, bytes: %X, value: %d",
					i, mileageBytes, mileageInt)
				extendedData["Mileage"] = mileageInt
			}
		}
	}

	// Keep the rest of the function to process other data
	offset := 0
	for offset < len(data) {
		// Need at least 2 more bytes for ID and length
		if offset+2 >= len(data) {
			break
		}

		id := data[offset]
		length := int(data[offset+1])
		offset += 2

		utils.VPrint("DEBUG BSJ: Processing ID: 0x%02X, Length: %d at offset %d", id, length, offset-2)

		// Check if we have enough bytes for the value
		if offset+length > len(data) {
			utils.VPrint("DEBUG BSJ: Not enough data for ID 0x%02X, needed %d bytes", id, length)
			break
		}

		value := data[offset : offset+length]

		// Process specific IDs - but skip 0x01 and 0x31 as we've already processed them
		switch id {
		case 0x30: // GSM signal strength
			if length > 0 {
				gsm := int(value[0])
				utils.VPrint("DEBUG BSJ: Found GSM signal strength (0x30): %d", gsm)
				extendedData["GsmSignalStrength"] = gsm
			}
		}

		offset += length
	}

	// Final debug log for extracted values
	if mileage, ok := extendedData["Mileage"]; ok {
		utils.VPrint("DEBUG BSJ: Final extracted Mileage: %v", mileage)
	}
	if sats, ok := extendedData["NumberOfSatellites"]; ok {
		utils.VPrint("DEBUG BSJ: Final extracted NumberOfSatellites: %v", sats)
	}

	return extendedData
}

// ParseExtendedDataForIMEI specifically looks for IMEI in extended data
func ParseExtendedDataForIMEI(data []byte) map[string]interface{} {
	result := make(map[string]interface{})

	// Debug the data we're searching through
	if verbose {
		utils.VPrint("DEBUG BSJ: Searching for IMEI in %d bytes of extended data: %X", len(data), data)
	}

	// According to BSJ-EG01 protocol, IMEI is stored with ID 0x00D5
	for i := 0; i < len(data)-4; i++ {
		// Look for the exact pattern specified in BSJ-EG01:
		// Length (2 bytes) + ID (0x00D5) + IMEI data (15 bytes)
		if i+4 < len(data) && data[i+2] == 0x00 && data[i+3] == 0xD5 {
			// Extract length (first 2 bytes)
			length := int(data[i])<<8 | int(data[i+1])

			// The length should be around 17 (15 for IMEI + 2 for ID)
			if verbose {
				utils.VPrint("DEBUG BSJ: Found potential IMEI structure at position %d, length: %d", i, length)
			}

			// Make sure we have enough data and that the length makes sense
			if i+4+length-2 <= len(data) && length >= 15 {
				// Extract the IMEI (skipping the ID bytes)
				imeiStart := i + 4        // Skip length (2) and ID (2)
				imeiEnd := imeiStart + 15 // IMEI is 15 digits

				if imeiEnd <= len(data) {
					// Try to convert to string - IMEI can be stored as ASCII or binary
					imeiBytes := data[imeiStart:imeiEnd]

					// First try ASCII representation
					imei := string(imeiBytes)

					// Check if this looks like a valid IMEI (should contain only digits)
					isValid := true
					for _, c := range imei {
						if c < '0' || c > '9' {
							isValid = false
							break
						}
					}

					if isValid && len(imei) == 15 {
						utils.VPrint("DEBUG BSJ: Found valid IMEI (ASCII): %s", imei)
						result["IMEI"] = imei
						return result
					}

					// If not valid as ASCII, try to decode as BCD or other binary format
					// (Uncommon but supported by some devices)
					var imeiStr strings.Builder
					for _, b := range imeiBytes {
						imeiStr.WriteString(fmt.Sprintf("%02d", b))
					}

					// Trim to 15 digits (standard IMEI length)
					imeiCandidate := imeiStr.String()
					if len(imeiCandidate) >= 15 {
						imeiCandidate = imeiCandidate[:15]
						utils.VPrint("DEBUG BSJ: Found IMEI (binary): %s", imeiCandidate)
						result["IMEI"] = imeiCandidate
						return result
					}

					utils.VPrint("DEBUG BSJ: Invalid IMEI format: %X", imeiBytes)
				}
			}
		}
	}

	// Secondary scan strategy - look for the complete sequence
	// This handles the case where the IMEI appears as: 00 11 00 D5 38 36...
	// Where 38 36... is the ASCII representation of "86..."
	for i := 0; i < len(data)-5; i++ {
		if data[i] == 0x00 && data[i+1] >= 0x0F && data[i+1] <= 0x12 && // Length around 15-18
			data[i+2] == 0x00 && data[i+3] == 0xD5 { // IMEI identifier 0x00D5

			// Calculate appropriate length based on protocol
			length := int(data[i+1])

			if verbose {
				utils.VPrint("DEBUG BSJ: Found secondary IMEI marker at position %d, length: %d",
					i, length)
			}

			// Ensure we have enough data
			if i+4+length <= len(data) {
				// Extract the complete IMEI data
				imeiBytes := data[i+4 : i+4+length]
				imei := string(imeiBytes)

				// Validate IMEI - should be 15 digits
				if len(imei) >= 15 {
					// Take first 15 chars if longer
					imei = imei[:15]
					utils.VPrint("DEBUG BSJ: Extracted IMEI from extended format: %s", imei)
					result["IMEI"] = imei
					return result
				}
			}
		}
	}

	if verbose {
		utils.VPrint("DEBUG BSJ: No valid IMEI found in extended data")
	}

	return result
}

// Convertir número de teléfono a formato BCD according to BSJ protocol
func buildPhoneBCD(phoneNumber string) []byte {
	// Ensure the phone number has an even number of digits by padding if necessary
	if len(phoneNumber) == 0 {
		// Return a default value for empty strings (6 bytes of zeros)
		return []byte{0, 0, 0, 0, 0, 0}
	}

	// According to BSJ protocol specification:
	// - If less than 12 digits, add padding in front
	// - Mainland China numbers should be padded with '0'

	// First ensure even length for BCD encoding
	if len(phoneNumber)%2 != 0 {
		phoneNumber = "0" + phoneNumber
	}

	// Ensure we meet the BSJ protocol requirement of having 12 digits (6 bytes)
	if len(phoneNumber) > 12 {
		// If longer than 12 digits, truncate (take last 12 digits)
		phoneNumber = phoneNumber[len(phoneNumber)-12:]
		if verbose {
			utils.VPrint("DEBUG BSJ: Phone number too long, truncated to %s", phoneNumber)
		}
	} else if len(phoneNumber) < 12 {
		// If shorter than 12 digits, pad with zeros in front (per protocol)
		padding := strings.Repeat("0", 12-len(phoneNumber))
		phoneNumber = padding + phoneNumber
		if verbose {
			utils.VPrint("DEBUG BSJ: Phone number padded to %s", phoneNumber)
		}
	}

	// Now convert to BCD (6 bytes for 12 digits)
	phoneBCD := make([]byte, 6)
	for i := 0; i < 12 && i/2 < 6; i += 2 {
		// Convert each pair of digits to a BCD byte
		highNibble := phoneNumber[i] - '0'
		lowNibble := phoneNumber[i+1] - '0'
		phoneBCD[i/2] = (highNibble << 4) | lowNibble
	}

	if verbose {
		utils.VPrint("DEBUG BSJ: Converted phone %s to BCD: %X", phoneNumber, phoneBCD)
	}

	return phoneBCD
}

func getIdName(id string) string {
	idNameMap := map[string]string{
		"01":       "Mileage",
		"30":       "GsmSignalStrength",
		"31":       "NumberOfSatellites",
		"000c00b2": "SIM",
		"00060089": "AlarmStatus",
		"000b00d8": "BaseStation",
		"000600c5": "AlarmStatus2",
		"0006002d": "ExternalVoltage",
		"001100d5": "IMEI",
		"00b9":     "WiFi",
	}

	if name, found := idNameMap[id]; found {
		return name
	}
	return fmt.Sprintf("UnknownID_%s", id) // Para IDs desconocidos
}

func GenerateAuthenticationResponse(serialNumber []byte, phoneNumber string) string {
	messageID := []byte{0x80, 0x01} // ID del mensaje para respuesta de autenticación

	// Convertir número de teléfono a BCD
	phoneBCD := buildPhoneBCD(phoneNumber)

	// Construir el cuerpo del mensaje
	body := buildBodyAuthResponse(serialNumber, 0x00)

	// Calcular la longitud del cuerpo
	bodyLength := make([]byte, 2)
	bodyLength[0] = byte(len(body) >> 8) // Byte alto
	bodyLength[1] = byte(len(body))      // Byte bajo

	// Construir el encabezado
	messageSerialNumber := generateRandomSerialNumber()
	header := buildHeader(messageID, bodyLength, phoneBCD, messageSerialNumber)

	// Concatenar encabezado y cuerpo
	data := append(header, body...)

	// Calcular el checksum
	checksum := CalculateChecksum(data)

	// Formar la trama completa
	return fmt.Sprintf("7e%s%02x7e", hex.EncodeToString(data), checksum)
}

func buildBodyAuthResponse(serialNumber []byte, result byte) []byte {
	messageID := []byte{0x01, 0x02}            // ID del mensaje de autenticación
	body := append(serialNumber, messageID...) // Agregar serialNumber seguido del messageID
	return append(body, result)                // Agregar el resultado (0x00 para éxito)
}

// Construir el cuerpo del mensaje de registro (serial, resultado, auth code)
func buildRegistrationBody(serialNumber []byte, result byte, authCode []byte) []byte {
	body := append(serialNumber, result) // Serial Number + Resultado
	return append(body, authCode...)     // Agregar Código de Autenticación
}

// Calcular la longitud del cuerpo
func calculateBodyLength(body []byte) []byte {
	return []byte{byte(len(body) >> 8), byte(len(body))}
}

// Construir el encabezado del mensaje
func buildHeader(messageID []byte, bodyLength []byte, phoneBCD []byte, messageSerialNumber []byte) []byte {
	header := append(messageID, bodyLength...)    // Agrega Message ID y Longitud
	header = append(header, phoneBCD...)          // Agrega el número de teléfono en BCD
	return append(header, messageSerialNumber...) // Agrega Message Serial Number
}

// Generar la trama completa del registro
func GenerateRegistrationResponse(serialNumber []byte, phoneNumber string) string {
	messageID := []byte{0x81, 0x00}                        // ID de respuesta
	phoneBCD := buildPhoneBCD(phoneNumber)                 // Número de teléfono en BCD
	authCode := []byte{0x62, 0x73, 0x6A, 0x67, 0x70, 0x73} // Código de autenticación
	result := byte(0x00)                                   // Éxito

	// Construcción por partes
	body := buildRegistrationBody(serialNumber, result, authCode) // Cuerpo
	bodyLength := calculateBodyLength(body)                       // Longitud del cuerpo
	messageSerialNumber := generateRandomSerialNumber()
	header := buildHeader(messageID, bodyLength, phoneBCD, messageSerialNumber) // Encabezado
	data := append(header, body...)                                             // Encabezado + Cuerpo

	// Calcular el checksum
	checksum := CalculateChecksum(data)

	// Generar trama completa
	return fmt.Sprintf("7e%s%02x7e", hex.EncodeToString(data), checksum)
}

func CalculateChecksum(data []byte) byte {
	var checksum byte
	for _, b := range data {
		checksum ^= b
	}
	return checksum
}

func generateRandomSerialNumber() []byte {
	randValue := rand.Intn(0xFFFF) // Generar un valor entre 0x0000 y 0xFFFF
	return []byte{byte(randValue >> 8), byte(randValue & 0xFF)}
}

// Decodifica un número BCD a una cadena
func DecodeBCD2(data []byte) string {
	var decoded strings.Builder
	for _, b := range data {
		decoded.WriteString(fmt.Sprintf("%d", b>>4))
		decoded.WriteString(fmt.Sprintf("%d", b&0x0F))
	}
	return decoded.String()
}

// DecodeBCD converts BCD encoded bytes to string
func DecodeBCD(bcd []byte) string {
	var result strings.Builder
	for _, b := range bcd {
		// Each byte in BCD contains two decimal digits
		highNibble := (b >> 4) & 0x0F
		lowNibble := b & 0x0F
		// Only add digits 0-9 (ignore invalid BCD values)
		if highNibble <= 9 {
			result.WriteByte('0' + highNibble)
		}
		if lowNibble <= 9 {
			result.WriteByte('0' + lowNibble)
		}
	}
	return result.String()
}

// DecodeTerminalMobileNumber properly handles the terminal mobile number according to BSJ protocol
func DecodeTerminalMobileNumber(bcd []byte) string {
	if len(bcd) != 6 {
		utils.VPrint("DEBUG BSJ: Invalid BCD length for mobile number: %d, expected 6", len(bcd))
		return "invalid-bcd-length"
	}

	// Decode all 12 digits from the 6 BCD bytes
	mobileNumber := DecodeBCD(bcd)

	// Remove any padding (F) that might be present
	mobileNumber = strings.ReplaceAll(mobileNumber, "f", "")

	// Remove leading zeros (padding) according to protocol
	mobileNumber = strings.TrimLeft(mobileNumber, "0")

	// Debug the raw mobile number
	utils.VPrint("DEBUG BSJ: Decoded mobile number (after removing padding): %s", mobileNumber)

	// Check if this looks like a truncated IMEI (usually starting with '99' or other patterns)
	if len(mobileNumber) >= 11 && (strings.HasPrefix(mobileNumber, "99") || strings.HasPrefix(mobileNumber, "86")) {
		// This is likely a truncated IMEI - reconstruct the full IMEI
		if strings.HasPrefix(mobileNumber, "86") {
			// If it starts with 86, this may be a China-format IMEI
			utils.VPrint("DEBUG BSJ: Detected IMEI in China format: %s", mobileNumber)
			return mobileNumber
		} else {
			// For other devices, add standard prefix if needed to reach 15 digits
			if len(mobileNumber) < 15 {
				prefix := "86"
				fullIMEI := prefix + mobileNumber
				utils.VPrint("DEBUG BSJ: Reconstructed IMEI with prefix: %s", fullIMEI)
				return fullIMEI
			}
			utils.VPrint("DEBUG BSJ: Using IMEI as is: %s", mobileNumber)
			return mobileNumber
		}
	}

	// If we can't determine a specific pattern, return the number as is
	// This ensures backward compatibility with existing devices
	utils.VPrint("DEBUG BSJ: Using mobile number as is: %s", mobileNumber)
	return mobileNumber
}

func FormatToISO8601(decodedTime string) (string, error) {
	// Validar la longitud de la cadena
	if len(decodedTime) != 12 {
		return "", fmt.Errorf("Error: Invalid decodedTime length (%d), expected 12", len(decodedTime))
	}

	// Extraer los componentes de la fecha y hora
	year := "20" + decodedTime[:2]
	month := decodedTime[2:4]
	day := decodedTime[4:6]
	hour := decodedTime[6:8]
	minute := decodedTime[8:10]
	second := decodedTime[10:12]

	// Parsear la fecha y hora
	parsedTime, err := time.Parse("20060102150405", year+month+day+hour+minute+second)
	if err != nil {
		return "", fmt.Errorf("Error parsing time: %v", err)
	}

	// Retornar en formato ISO 8601
	return parsedTime.Format("2006-01-02T15:04:05Z"), nil
}

func BytesToHexAndDecimal(data []byte) (string, *big.Int, error) {
	// Convertir a hexadecimal
	hexValue := hex.EncodeToString(data)

	// Convertir hexadecimal a decimal
	decimalValue := new(big.Int)
	_, ok := decimalValue.SetString(hexValue, 16) // Base 16 (hexadecimal)
	if !ok {
		return "", nil, fmt.Errorf("error converting hex to decimal")
	}
	return hexValue, decimalValue, nil
}

func hexToUTF8(hexString string) (string, error) {
	// Decodificar el string hexadecimal a bytes
	decodedBytes, err := hex.DecodeString(hexString)
	if err != nil {
		return "", fmt.Errorf("error decoding hex string: %v", err)
	}

	// Construir la representación UTF-8 a partir de los bytes
	utf8String := string(decodedBytes)
	return utf8String, nil
}

func hexToIntMileage(hexString string) (int, error) {
	// Parse hex string to int64 since mileage can be large
	value, err := strconv.ParseInt(hexString, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting mileage hex to int: %v", err)
	}
	// Return the value in 1/10 km units
	return int(value), nil
}

// GenerateHeartbeatResponse creates a response for a heartbeat message
func GenerateHeartbeatResponse(serialNumber []byte, phoneNumber string) string {
	// For heartbeat, respond with general response (8001)
	messageID := []byte{0x80, 0x01} // Platform general response ID

	// Convert terminal heartbeat message ID (0002) to bytes for the response
	terminalMsgID := []byte{0x00, 0x02}

	// Prepare the response body:
	// - serialNumber (2 bytes from the original message)
	// - terminalMsgID (the ID of the original message we're responding to)
	// - result (0x00 = success)
	body := append(serialNumber, terminalMsgID...)
	body = append(body, 0x00) // Result: 0 = Success/Confirmation

	// Convert phone number to BCD
	phoneBCD := buildPhoneBCD(phoneNumber)

	// Calculate the body length
	bodyLength := calculateBodyLength(body)

	// Generate random message serial number
	messageSerialNumber := generateRandomSerialNumber()

	// Build header
	header := buildHeader(messageID, bodyLength, phoneBCD, messageSerialNumber)

	// Combine header and body
	data := append(header, body...)

	// Calculate checksum
	checksum := CalculateChecksum(data)

	// Format complete response with flag bits and checksum
	return fmt.Sprintf("7e%s%02x7e", hex.EncodeToString(data), checksum)
}
