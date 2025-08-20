package models

import "time"

// Queclink350Model represents the parsed data from a Queclink 350 device
type Queclink350Model struct {
	// Basic information
	DeviceType    string // Device type (350)
	DeviceVersion string // Version of the device
	IMEI          string // Unique device identifier
	DeviceName    string // Name assigned to the device
	MessageType   string // Type of message (e.g., +RESP, +BUFF)
	ReportType    string // Type of report
	ReportID      string // ID of the report

	// Event information
	EventCode string // Code representing the event

	// Location information
	Timestamp  time.Time // Date and time of the report
	Latitude   float64   // GPS latitude
	Longitude  float64   // GPS longitude
	Speed      float64   // Speed in km/h
	Heading    float64   // Direction in degrees
	Altitude   float64   // Altitude in meters
	Satellites int       // Number of GPS satellites

	// Network information
	LAC    string // Location Area Code
	CellID string // Cell ID

	// Status information
	Ignition      bool    // Ignition status (true = on, false = off)
	BatteryLevel  float64 // Internal battery voltage
	ExternalPower float64 // External power supply voltage
	InputStatus   int     // Digital input status
	OutputStatus  int     // Digital output status

	// Additional information
	Odometer float64 // Odometer reading in km

	// Original data
	RawData string // Original unparsed data for reference
}
