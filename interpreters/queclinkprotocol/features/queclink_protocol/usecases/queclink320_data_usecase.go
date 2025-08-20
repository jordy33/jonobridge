package usecases

import (
	"fmt"
	"queclinkprotocol/features/queclink_protocol/models"
	"strconv"
	"strings"
	"time"
)

// ParseQueclink320Fields parses data from a Queclink 320 device
func ParseQueclink320Fields(data string) (models.Queclink320Model, error) {
	// Initialize an empty model
	q320 := models.Queclink320Model{RawData: data}

	// Split the data by commas
	parts := strings.Split(data, ",")
	if len(parts) < 4 {
		return q320, fmt.Errorf("invalid Queclink 320 data: too few fields (%d)", len(parts))
	}

	// Get message type (e.g., +RESP, +BUFF)
	messageType := parts[0][:5]
	q320.MessageType = messageType

	// Extract device info
	if len(parts[1]) >= 2 {
		q320.DeviceType = parts[1][:2]
	}
	if len(parts[1]) >= 4 {
		version, err := strconv.ParseInt(parts[1][2:4], 16, 64)
		if err == nil {
			subVersion, err := strconv.ParseInt(parts[1][4:], 16, 64)
			if err == nil {
				if subVersion <= 9 {
					q320.DeviceVersion = fmt.Sprintf("%v.0%v", version, subVersion)
				} else {
					q320.DeviceVersion = fmt.Sprintf("%v.%v", version, subVersion)
				}
			}
		}
	}

	// IMEI and device name
	if len(parts) > 2 {
		q320.IMEI = parts[2]
	}
	if len(parts) > 3 {
		q320.DeviceName = parts[3]
	}

	// Determine message format and parse accordingly
	msgCommand := ""
	if len(parts[0]) > 6 {
		msgCommand = parts[0][6:] // Get command type (e.g., GTFRI)
	}

	// Different parsing based on message command
	switch msgCommand {
	case "GTFRI": // Position report
		if len(parts) > 10 {
			// Parse timestamp
			if len(parts[4]) >= 14 {
				timestamp, err := time.Parse("20060102150405", parts[4])
				if err == nil {
					q320.Timestamp = timestamp
				}
			}

			// Parse coordinates
			if len(parts) > 6 {
				lat, err := strconv.ParseFloat(parts[5], 64)
				if err == nil {
					q320.Latitude = lat
				}

				lon, err := strconv.ParseFloat(parts[6], 64)
				if err == nil {
					q320.Longitude = lon
				}
			}

			// Parse speed, heading, etc.
			if len(parts) > 7 {
				speed, err := strconv.ParseFloat(parts[7], 64)
				if err == nil {
					q320.Speed = speed
				}
			}

			if len(parts) > 8 {
				heading, err := strconv.ParseFloat(parts[8], 64)
				if err == nil {
					q320.Heading = heading
				}
			}

			if len(parts) > 9 {
				altitude, err := strconv.ParseFloat(parts[9], 64)
				if err == nil {
					q320.Altitude = altitude
				}
			}

			// Parse satellites and HDOP
			if len(parts) > 10 {
				satellites, err := strconv.Atoi(parts[10])
				if err == nil {
					q320.Satellites = satellites
				}
			}

			if len(parts) > 11 {
				hdop, err := strconv.ParseFloat(parts[11], 64)
				if err == nil {
					q320.HDOP = hdop
				}
			}

			// Parse ignition status (typically field 19 for 320)
			if len(parts) > 19 {
				ign, err := strconv.Atoi(parts[19])
				if err == nil {
					q320.Ignition = ign == 1
				}
			}

			// Parse external power and battery (typically fields 20, 21)
			if len(parts) > 20 {
				externalPower, err := strconv.ParseInt(parts[20], 16, 64)
				if err == nil {
					q320.ExternalPower = float64(externalPower) / 1000.0
				}
			}

			if len(parts) > 21 {
				battery, err := strconv.ParseFloat(parts[21], 64)
				if err == nil {
					q320.BatteryLevel = battery
				}
			}

			// Set default event code
			q320.EventCode = getEventCode320(msgCommand, "0")

			// Parse report ID and type if available
			if len(parts) > 13 && len(parts[13]) >= 2 {
				q320.ReportID = parts[13][:1]
				q320.ReportType = parts[13][1:]
				q320.EventCode = getEventCode320(msgCommand, q320.ReportType)
			}
		}

	case "GTGEO", "GTSPD", "GTSOS", "GTRTL", "GTPNL", "GTNMR", "GTDIS", "GTDOG", "GTIGL", "GTLOC":
		// Parsing common alarm/report messages
		if len(parts) > 10 {
			// Parse timestamp
			if len(parts[4]) >= 14 {
				timestamp, err := time.Parse("20060102150405", parts[4])
				if err == nil {
					q320.Timestamp = timestamp
				}
			}

			// Parse coordinates
			if len(parts) > 6 {
				lat, err := strconv.ParseFloat(parts[5], 64)
				if err == nil {
					q320.Latitude = lat
				}

				lon, err := strconv.ParseFloat(parts[6], 64)
				if err == nil {
					q320.Longitude = lon
				}
			}

			// Parse reportID/reportType
			if len(parts) > 12 && (msgCommand == "GTGEO" || msgCommand == "GTSPD" || msgCommand == "GTIGL" || msgCommand == "GTRTL") {
				q320.ReportID = parts[12]
				q320.EventCode = getEventCode320(msgCommand, q320.ReportID)
			}
		}
	}

	return q320, nil
}

// getEventCode320 returns the event code for Queclink 320 based on message type and report type
func getEventCode320(msgType, reportType string) string {
	// This is a simplified version - in a real implementation, you would have a more
	// comprehensive mapping based on the Queclink protocol specification
	switch msgType {
	case "GTFRI":
		return "0" // Regular position report
	case "GTGEO":
		if reportType == "0" {
			return "10" // Enter geo-fence
		}
		return "11" // Exit geo-fence
	case "GTSPD":
		if reportType == "0" {
			return "12" // Speed limit exceeded
		}
		return "13" // Speed back to normal
	case "GTIGL":
		if reportType == "0" {
			return "1" // Ignition on
		}
		return "2" // Ignition off
	case "GTTEM":
		if reportType == "0" {
			return "14" // Temperature alarm
		}
		return "15" // Temperature back to normal
	default:
		return "99" // Unknown event
	}
}
