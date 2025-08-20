package models

import "time"

// MessageType_Queclink maps message codes to their descriptive names
var MessageType_Queclink = map[string]string{
	"1":  "NTWCmd",
	"2":  "RPTCmd",
	"3":  "EVTCmd",
	"4":  "GSMCmd",
	"5":  "SVCCmd",
	"6":  "MBVCmd",
	"7":  "MSRCmd",
	"8":  "CGFCmd",
	"9":  "ADPCmd",
	"10": "NPTCmd",
	"11": "LTMCmd",
	"12": "PLGCmd",
	"13": "PLSCmd",
	"14": "PLCCmd",
	"15": "CTRCmd",
	"16": "STRCmd",
	"17": "GTRCmd",
	"18": "STTReport",
	"19": "EMGReport",
	"20": "EVTReport",
	"21": "ALTReport",
	"22": "ALVReport",
	"23": "UEXReport",
	"24": "DEXReport",
	"25": "CMD",
}

// Queclink300Model represents the parsed data from a Queclink 300 device
type Queclink300Model struct {
	// Basic information
	DeviceType    string // Device type (300)
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
	DeviceStatus  string  // Status of the device
	BatteryLevel  float64 // Internal battery voltage
	ExternalPower float64 // External power supply voltage

	// Original data
	RawData string // Original unparsed data for reference
}
