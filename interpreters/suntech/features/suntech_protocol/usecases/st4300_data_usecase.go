package usecases

import (
	"fmt"
	"strconv"
	"strings"
	"suntechprotocol/features/suntech_protocol/models"
	"time"
)

func ParseST4300Fields(data string) (models.ST4300Model, error) {
	parts := strings.Split(data, ";")
	if len(parts) < 12 { // Minimum expected fields for a basic ST4300 packet
		return models.ST4300Model{}, fmt.Errorf("invalid ST4300 data: too few fields (%d)", len(parts))
	}

	// Initialize the struct
	st4300 := models.ST4300Model{RawData: data}

	// Parse header and basic fields
	st4300.Header = parts[0] // e.g., "ST4300STT"
	st4300.IMEI = parts[1]   // Device ID / IMEI

	// Parse message type
	msgCode := parts[2] // e.g., "18"
	if msgType, exists := models.MessageType_suntech[msgCode]; exists {
		st4300.MessageType = msgType // e.g., "STTReport"
	} else {
		st4300.MessageType = "Unknown (" + msgCode + ")"
	}

	// Parse timestamp (format: YYYYMMDDHHMMSS, e.g., "20250407120000")
	timestamp, err := time.Parse("20060102150405", parts[3])
	if err != nil {
		return models.ST4300Model{}, fmt.Errorf("invalid timestamp: %v", err)
	}
	st4300.Timestamp = timestamp

	// Parse latitude (e.g., "+37.123456")
	if len(parts[4]) > 0 {
		lat, err := parseCoordinate(parts[4])
		if err != nil {
			return models.ST4300Model{}, fmt.Errorf("invalid latitude: %v", err)
		}
		st4300.Latitude = lat
	}

	// Parse longitude (e.g., "-122.123456")
	if len(parts[5]) > 0 {
		lon, err := parseCoordinate(parts[5])
		if err != nil {
			return models.ST4300Model{}, fmt.Errorf("invalid longitude: %v", err)
		}
		st4300.Longitude = lon
	}

	// Parse speed (e.g., "60")
	if speed, err := parseFloat(parts[6]); err == nil {
		st4300.Speed = speed
	}

	// Parse heading (e.g., "180")
	if heading, err := parseFloat(parts[7]); err == nil {
		st4300.Heading = heading
	}

	// Parse satellites (e.g., "10")
	if sats, err := parseInt(parts[8]); err == nil {
		st4300.Satellites = sats
	}

	// Parse HDOP (horizontal dilution of precision)
	if len(parts) > 9 && parts[9] != "" {
		if hdop, err := parseFloat(parts[9]); err == nil {
			st4300.HDOP = hdop
		}
	}

	// Parse altitude
	if len(parts) > 10 && parts[10] != "" {
		if altitude, err := parseFloat(parts[10]); err == nil {
			st4300.Altitude = altitude
		}
	}

	// Parse ignition (e.g., "1" or "0")
	if len(parts) > 11 {
		if ign, err := parseInt(parts[11]); err == nil {
			st4300.Ignition = ign == 1
		}
	}

	// Parse battery level (if present)
	if len(parts) > 12 && parts[12] != "" {
		if battery, err := parseFloat(parts[12]); err == nil {
			st4300.BatteryLevel = battery
		}
	}

	// Parse odometer (if present)
	if len(parts) > 13 && parts[13] != "" {
		if odometer, err := parseFloat(parts[13]); err == nil {
			st4300.Odometer = odometer
		}
	}

	// Parse input status (if present)
	if len(parts) > 14 && parts[14] != "" {
		if inputStatus, err := parseUint32(parts[14]); err == nil {
			st4300.InputStatus = inputStatus
		}
	}

	// Parse output status (if present)
	if len(parts) > 15 && parts[15] != "" {
		if outputStatus, err := parseUint32(parts[15]); err == nil {
			st4300.OutputStatus = outputStatus
		}
	}

	return st4300, nil
}

// Helper function to parse uint32 values
func parseUint32(s string) (uint32, error) {
	i, err := strconv.ParseUint(s, 10, 32)
	return uint32(i), err
}
