package usecases

import (
	"fmt"
	"strings"
	"suntechprotocol/features/suntech_protocol/models"
	"time"
)

func ParseST300Fields(data string) (models.ST300Model, error) {
	parts := strings.Split(data, ";")
	if len(parts) < 10 { // Minimum expected fields for a basic ST300 packet
		return models.ST300Model{}, fmt.Errorf("invalid ST300 data: too few fields (%d)", len(parts))
	}

	// Initialize the struct
	st300 := models.ST300Model{RawData: data}

	// Parse fields (adjust indices based on actual ST300 protocol)
	st300.Header = parts[0] // e.g., "ST300STT"
	st300.IMEI = parts[1]   // e.g., "123456789"
	msgCode := parts[2]     // e.g., "18"
	if msgType, exists := models.MessageType_suntech[msgCode]; exists {
		st300.MessageType = msgType // e.g., "STTReport"
	} else {
		st300.MessageType = "Unknown (" + msgCode + ")"
	}

	// Parse timestamp (format: YYYYMMDDHHMMSS, e.g., "20250407120000")
	timestamp, err := time.Parse("20060102150405", parts[3])
	if err != nil {
		return models.ST300Model{}, fmt.Errorf("invalid timestamp: %v", err)
	}
	st300.Timestamp = timestamp

	// Parse latitude (e.g., "+37.123456")
	if len(parts[4]) > 0 {
		lat, err := parseCoordinate(parts[4])
		if err != nil {
			return models.ST300Model{}, fmt.Errorf("invalid latitude: %v", err)
		}
		st300.Latitude = lat
	}

	// Parse longitude (e.g., "-122.123456")
	if len(parts[5]) > 0 {
		lon, err := parseCoordinate(parts[5])
		if err != nil {
			return models.ST300Model{}, fmt.Errorf("invalid longitude: %v", err)
		}
		st300.Longitude = lon
	}

	// Parse speed (e.g., "60")
	if speed, err := parseFloat(parts[6]); err == nil {
		st300.Speed = speed
	}

	// Parse heading (e.g., "180")
	if heading, err := parseFloat(parts[7]); err == nil {
		st300.Heading = heading
	}

	// Parse satellites (e.g., "10")
	if sats, err := parseInt(parts[8]); err == nil {
		st300.Satellites = sats
	}

	// Parse ignition (e.g., "1" or "0")
	if ign, err := parseInt(parts[9]); err == nil {
		st300.Ignition = ign == 1
	}

	// Optional: Battery level (if present in additional fields, e.g., parts[10])
	if len(parts) > 10 {
		if battery, err := parseFloat(parts[10]); err == nil {
			st300.BatteryLevel = battery
		}
	}

	return st300, nil
}

// Helper function to parse coordinates (e.g., "+37.123456" or "-122.123456")
func parseCoordinate(coord string) (float64, error) {
	if len(coord) < 2 {
		return 0, fmt.Errorf("coordinate too short")
	}
	value, err := parseFloat(coord)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// Helper function to parse float values
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// Helper function to parse int values
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
