package huabao_protocol

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"math"
	"huabaoprotocol/features/jono/models"
)

// FloatToString converts a float to a string with specified precision
func FloatToString(input_num float64, precision int) string {
	return strconv.FormatFloat(input_num, 'f', precision, 64)
}

// Parse parses the raw Huabao protocol data
func Parse(data string) (string, error) {
	// Check if data is empty
	if data == "" {
		return "", fmt.Errorf("empty data received")
	}

	// Try to parse as Huabao protocol
	if strings.Contains(data, "#") {
		// Handle DVR-like Huabao format
		return parseDvrFormat(data)
	}

	// Create ParsedModel for non-DVR format
	parsedModel := &models.ParsedModel{
		ListPackets: make(map[string]models.Packet),
	}
	
	message := data
	parsedModel.Message = &message
	
	dataPackets := 1
	parsedModel.DataPackets = &dataPackets

	// Create the packet
	packet := models.Packet{
		Altitude: 0,
		NumberOfSatellites: 0,
	}

	// Try to extract IMEI and other information
	imeiRegex := regexp.MustCompile(`IMEI:(\d+)`)
	if match := imeiRegex.FindStringSubmatch(data); len(match) > 1 {
		imei := match[1]
		parsedModel.IMEI = &imei
	} else {
		// If we can't find IMEI in standard format, look for it elsewhere
		parts := strings.Split(data, ",")
		if len(parts) > 3 {
			imei := parts[2]
			parsedModel.IMEI = &imei
		}
	}

	// Extract coordinates if present
	latRegex := regexp.MustCompile(`lat[^:]*:([0-9.-]+)`)
	lonRegex := regexp.MustCompile(`lon[^:]*:([0-9.-]+)`)
	
	if match := latRegex.FindStringSubmatch(data); len(match) > 1 {
		if lat, err := strconv.ParseFloat(match[1], 64); err == nil {
			packet.Latitude = &lat
		}
	}
	
	if match := lonRegex.FindStringSubmatch(data); len(match) > 1 {
		if lon, err := strconv.ParseFloat(match[1], 64); err == nil {
			packet.Longitude = &lon
		}
	}

	// Extract speed if present
	speedRegex := regexp.MustCompile(`speed:([0-9.]+)`)
	if match := speedRegex.FindStringSubmatch(data); len(match) > 1 {
		if speed, err := strconv.ParseFloat(match[1], 64); err == nil {
			speedInt := int(speed)
			packet.Speed = &speedInt
		}
	}

	// Extract heading/direction if present
	headingRegex := regexp.MustCompile(`(?:heading|direction):([0-9.]+)`)
	if match := headingRegex.FindStringSubmatch(data); len(match) > 1 {
		if heading, err := strconv.ParseFloat(match[1], 64); err == nil {
			headingInt := int(heading)
			packet.Direction = &headingInt
		}
	}

	// Default event code for regular tracking
	packet.EventCode = models.EventCode{
		Code: 35,
		Name: "Track By Time Interval",
	}
	
	// Set current time if not found
	datetimeStr := time.Now().UTC().Format(time.RFC3339)
	packet.Datetime = &datetimeStr
	
	// Set default positioning status
	posStatus := "A"
	packet.PositioningStatus = &posStatus

	// Add packet to the model
	parsedModel.ListPackets["0"] = packet

	// Convert to JSON using the model's method
	result, err := parsedModel.ToJSON()
	if err != nil {
		return "", fmt.Errorf("error marshaling Huabao data: %v", err)
	}
	
	return result, nil
}

// parseDvrFormat handles the DVR-specific format of Huabao protocol
func parseDvrFormat(data string) (string, error) {
	lines := strings.Split(data, "#")
	if len(lines) == 0 {
		return "", fmt.Errorf("invalid DVR format")
	}

	// Process the first line to extract data
	line := lines[0]
	fields := strings.Split(line, ",")
	
	if len(fields) < 15 {
		return "", fmt.Errorf("not enough fields in DVR format data")
	}
	
	// Basic validation - check for $$ prefix
	if !strings.HasPrefix(fields[0], "$$") {
		return "", fmt.Errorf("invalid packet format: missing $$ prefix")
	}
	
	// Debug: Print extracted values before processing
	fmt.Println("=== DEBUG: Extracted Values ===")
	if len(fields) > 2 {
		fmt.Printf("Event: %s\n", fields[2])
	}
	if len(fields) > 3 {
		fmt.Printf("IMEI: %s\n", fields[3])
	}
	if len(fields) > 5 {
		fmt.Printf("Datetime: %s\n", fields[5])
	}
	if len(fields) > 7 {
		fmt.Printf("Longitude Degrees: %s\n", fields[7])
	}
	if len(fields) > 8 {
		fmt.Printf("Longitude Minutes: %s\n", fields[8])
	}
	if len(fields) > 9 {
		fmt.Printf("Longitude Seconds: %s\n", fields[9])
	}
	if len(fields) > 10 {
		fmt.Printf("Latitude Degrees: %s\n", fields[10])
	}
	if len(fields) > 11 {
		fmt.Printf("Latitude Minutes: %s\n", fields[11])
	}
	if len(fields) > 12 {
		fmt.Printf("Latitude Seconds: %s\n", fields[12])
	}
	if len(fields) > 13 {
		fmt.Printf("Speed: %s\n", fields[13])
	}
	if len(fields) > 14 {
		fmt.Printf("Heading: %s\n", fields[14])
	}
	fmt.Println("===============================")
	
	// Create the ParsedModel
	parsedModel := &models.ParsedModel{
		ListPackets: make(map[string]models.Packet),
	}
	
	// Extract IMEI
	imei := fields[3]
	parsedModel.IMEI = &imei
	
	// Set message
	message := line
	parsedModel.Message = &message
	
	// Set data packets count
	dataPackets := 1
	parsedModel.DataPackets = &dataPackets
	
	// Create the packet
	packet := models.Packet{
		Altitude: 0, // Default altitude
		NumberOfSatellites: 0, // Default satellites
	}
	
	// Event code with special handling for V201 and V251
	eventCode := 35 // Default event code
	eventName := "Track By Time Interval" // Default event name
	if len(fields) > 2 {
		eventStr := fields[2]
		if strings.HasPrefix(eventStr, "V") && len(eventStr) > 1 {
			// Special handling for alarm events V201 and V251
			if eventStr == "V201" || eventStr == "V251" {
				eventCode = 1
				eventName = "Panic/Alarm"
			} else {
				// Try to extract numeric part of other V events (V101, V142 etc.)
				// if code, err := strconv.Atoi(eventStr[1:]); err == nil {
				// 	eventCode = code
				// }
				eventCode = 35 // Default to tracking event for non-alarm V events
	            eventName = "Track By Time Interval"
			}
			// Keep the original event string as name for non-alarm events
			if eventCode != 1 {
				eventName = eventStr
			}
		}
	}
	packet.EventCode = models.EventCode{
		Code: eventCode,
		Name: eventName,
	}
	
	// Timestamp
	datetime, err := parseDateTime(fields[5])
	if err != nil {
		datetime = time.Now().UTC()
	}
	datetimeStr := datetime.Format(time.RFC3339)
	packet.Datetime = &datetimeStr
	
	// Location data - using exact algorithm from original.txt
	// Parse longitude
	lonDeg, lonMin, lonSec := fields[7], fields[8], fields[9]
	
	// Parse latitude
	latDeg, latMin, latSec := fields[10], fields[11], fields[12]
	
	// Calculate longitude using original algorithm
	var longitude float64 = 0.0
	is_lon_negative := false
	
	if degrees, err := strconv.ParseFloat(lonDeg, 64); err == nil {
		if degrees < 0 {
			degrees = degrees * -1
			is_lon_negative = true
		}
		longitude = longitude + degrees
	} else {
		return "", fmt.Errorf("invalid longitude degrees: %s", lonDeg)
	}
	
	if minutes, err := strconv.ParseFloat(lonMin, 64); err == nil {
		minutes = minutes / 60
		longitude = longitude + minutes
	} else {
		return "", fmt.Errorf("invalid longitude minutes: %s", lonMin)
	}
	
	if seconds, err := strconv.ParseFloat(lonSec, 64); err == nil {
		seconds = (seconds / 10000000) // Scale factor for microseconds
		seconds = seconds / 3600       // Convert to degrees
		longitude = longitude + seconds
	} else {
		return "", fmt.Errorf("invalid longitude seconds: %s", lonSec)
	}
	
	if is_lon_negative {
		longitude = longitude * -1
	}
	
	// Round to 6 decimal places
	longitude = math.Round(longitude*1000000) / 1000000
	
	// Calculate latitude using original algorithm
	var latitude float64 = 0.0
	is_lat_negative := false
	
	if degrees, err := strconv.ParseFloat(latDeg, 64); err == nil {
		if degrees < 0 {
			degrees = degrees * -1
			is_lat_negative = true
		}
		latitude = latitude + degrees
	} else {
		return "", fmt.Errorf("invalid latitude degrees: %s", latDeg)
	}
	
	if minutes, err := strconv.ParseFloat(latMin, 64); err == nil {
		minutes = minutes / 60
		latitude = latitude + minutes
	} else {
		return "", fmt.Errorf("invalid latitude minutes: %s", latMin)
	}
	
	if seconds, err := strconv.ParseFloat(latSec, 64); err == nil {
		seconds = (seconds / 10000000) // Scale factor for microseconds
		seconds = seconds / 3600       // Convert to degrees
		latitude = latitude + seconds
	} else {
		return "", fmt.Errorf("invalid latitude seconds: %s", latSec)
	}
	
	if is_lat_negative {
		latitude = latitude * -1
	}
	
	// Round to 6 decimal places
	latitude = math.Round(latitude*1000000) / 1000000
	
	// Debug: Print calculated coordinates
	fmt.Println("=== DEBUG: Calculated Coordinates ===")
	fmt.Printf("Calculated Longitude: %.6f\n", longitude)
	fmt.Printf("Calculated Latitude: %.6f\n", latitude)
	fmt.Println("====================================")
	
	packet.Longitude = &longitude
	packet.Latitude = &latitude
	packet.NumberOfSatellites = 12
	// Speed - keep as float like original, then convert to int for model
	if speed, err := strconv.ParseFloat(fields[13], 64); err == nil {
		speedInt := int(speed)
		packet.Speed = &speedInt
		
		// Debug: Print speed conversion
		fmt.Printf("=== DEBUG: Speed conversion: %s -> %d ===\n", fields[13], speedInt)
	}
	
	// Heading/Direction - use exact same algorithm as original
	if heading, err := strconv.ParseFloat(fields[14], 64); err == nil {
		r := heading / 100.0 // Convert to proper format (e.g., 6800 -> 68.0)
		directionInt := int(r) // Convert to int for model
		packet.Direction = &directionInt
		
		// Debug: Print heading conversion
		fmt.Printf("=== DEBUG: Heading conversion: %s -> %.1f -> %d ===\n", fields[14], r, directionInt)
	}
	
	// Position status
	posStatus := "A" // Assume valid fix
	packet.PositioningStatus = &posStatus
	
	// Handle any additional fields if present
	if len(fields) > 15 {
		//ioStatus := fields[15]
		
		// Create IoPortsStatus with default values
		ioPortStatus := &models.IoPortsStatus{
			Port1: 0, Port2: 0, Port3: 0, Port4: 0,
			Port5: 0, Port6: 0, Port7: 0, Port8: 0,
		}
		packet.IoPortStatus = ioPortStatus
		
		// Add analog inputs if present
		if len(fields) > 16 {
			analogInputs := &models.AnalogInputs{}
			if len(fields) > 16 {
				ad1 := fields[16]
				analogInputs.AD1 = &ad1
			}
			if len(fields) > 17 {
				ad2 := fields[17]
				analogInputs.AD2 = &ad2
			}
			packet.AnalogInputs = analogInputs
		}
	}
	
	// Add packet to the model
	parsedModel.ListPackets["0"] = packet
	
	// Convert to JSON using the model's method
	result, err := parsedModel.ToJSON()
	if err != nil {
		return "", fmt.Errorf("error marshaling DVR data to Jono model: %v", err)
	}
	
	return result, nil
}

// parseDateTime attempts to parse the datetime from Huabao format
func parseDateTime(dateStr string) (time.Time, error) {
	// Try a few common formats
	formats := []string{
		"20060102150405",    // YYYYMMDDhhmmss
		"060102150405",      // YYMMDDhhmmss
		"2006-01-02 15:04:05", // YYYY-MM-DD hh:mm:ss
		time.RFC3339,        // ISO 8601
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("could not parse datetime: %s", dateStr)
}