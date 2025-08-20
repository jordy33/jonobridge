package models

import "time"

var MessageType_suntech = map[string]string{
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

// ST300Data represents the parsed data from an ST300 device
type ST300Model struct {
	Header       string    // e.g., "ST300STT"
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
	RawData      string    // Original unparsed data for reference
}
