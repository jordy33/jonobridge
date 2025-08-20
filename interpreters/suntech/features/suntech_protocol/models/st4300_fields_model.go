package models

import "time"

// ST4300Model represents the parsed data from an ST4300 device
type ST4300Model struct {
	Header       string    // e.g., "ST4300STT"
	IMEI         string    // Unique device identifier
	MessageType  string    // Parsed from MessageType_suntech (e.g., "STTReport")
	Timestamp    time.Time // Date and time of the report
	Latitude     float64   // GPS latitude
	Longitude    float64   // GPS longitude
	Speed        float64   // Speed in km/h
	Heading      float64   // Direction in degrees
	Satellites   int       // Number of GPS satellites
	BatteryLevel float64   // Battery voltage (if present)
	Ignition     bool      // Ignition status (1 = on, 0 = off)
	HDOP         float64   // Horizontal Dilution of Precision
	Altitude     float64   // Altitude in meters
	Odometer     float64   // Vehicle odometer reading (if present)
	InputStatus  uint32    // Input status flags
	OutputStatus uint32    // Output status flags
	RawData      string    // Original unparsed data for reference
}
