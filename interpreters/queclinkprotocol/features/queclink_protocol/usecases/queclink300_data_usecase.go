package usecases

import (
	"fmt"
	"queclinkprotocol/features/queclink_protocol/models"
	"strconv"
	"strings"
	"time"
)

// ParseQueclink300Fields parses data from a Queclink 300 device
func ParseQueclink300Fields(data string) (models.Queclink300Model, error) {
	// Initialize an empty model
	q300 := models.Queclink300Model{RawData: data}

	// Split the data by commas
	parts := strings.Split(data, ",")
	if len(parts) < 4 {
		return q300, fmt.Errorf("invalid Queclink 300 data: too few fields (%d)", len(parts))
	}

	// Get message type (e.g., +RESP, +BUFF)
	messageType := parts[0][:5]
	q300.MessageType = messageType

	// Extract device info
	if len(parts[1]) >= 2 {
		q300.DeviceType = parts[1][:2]
	}
	if len(parts[1]) >= 4 {
		version, err := strconv.ParseInt(parts[1][2:4], 16, 64)
		if err == nil {
			subVersion, err := strconv.ParseInt(parts[1][4:], 16, 64)
			if err == nil {
				if subVersion <= 9 {
					q300.DeviceVersion = fmt.Sprintf("%v.0%v", version, subVersion)
				} else {
					q300.DeviceVersion = fmt.Sprintf("%v.%v", version, subVersion)
				}
			}
		}
	}

	// IMEI and device name
	if len(parts) > 2 {
		q300.IMEI = parts[2]
	}
	if len(parts) > 3 {
		q300.DeviceName = parts[3]
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
					q300.Timestamp = timestamp
				}
			}

			// Parse coordinates
			if len(parts) > 6 {
				lat, err := strconv.ParseFloat(parts[5], 64)
				if err == nil {
					q300.Latitude = lat
				}

				lon, err := strconv.ParseFloat(parts[6], 64)
				if err == nil {
					q300.Longitude = lon
				}
			}

			// Parse speed, heading, etc.
			if len(parts) > 7 {
				speed, err := strconv.ParseFloat(parts[7], 64)
				if err == nil {
					q300.Speed = speed
				}
			}

			if len(parts) > 8 {
				heading, err := strconv.ParseFloat(parts[8], 64)
				if err == nil {
					q300.Heading = heading
				}
			}

			if len(parts) > 9 {
				altitude, err := strconv.ParseFloat(parts[9], 64)
				if err == nil {
					q300.Altitude = altitude
				}
			}

			// Parse satellites
			if len(parts) > 10 {
				satellites, err := strconv.Atoi(parts[10])
				if err == nil {
					q300.Satellites = satellites
				}
			}

			// Set event code
			q300.EventCode = "0" // Default event code
			if len(parts) > 13 { // Check device status position
				q300.DeviceStatus = parts[13]
				q300.EventCode = getEventCode300(msgCommand, q300.DeviceStatus)
			}

			// Parse ignition status
			if len(parts) > 15 { // Ignition typically in field 15 of GTFRI
				ign, err := strconv.Atoi(parts[15])
				if err == nil {
					q300.Ignition = ign == 1
				}
			}

			// Parse battery level
			if len(parts) > 16 {
				externalBattery, err := strconv.ParseInt(parts[16], 16, 64)
				if err == nil {
					q300.ExternalPower = float64(externalBattery) / 1000.0
				}
			}

			if len(parts) > 17 {
				batteryLevel, err := strconv.ParseFloat(parts[17], 64)
				if err == nil {
					q300.BatteryLevel = batteryLevel
				}
			}
		}

	case "GTTOW", "GTDIS", "GTIOB", "GTSPD", "GTSOS", "GTRTL", "GTDOG", "GTIGL", "GTHBM", "GTVGL":
		// Parsing common alarm/report messages
		if len(parts) > 10 {
			// Parse timestamp
			if len(parts[4]) >= 14 {
				timestamp, err := time.Parse("20060102150405", parts[4])
				if err == nil {
					q300.Timestamp = timestamp
				}
			}

			// Parse coordinates
			if len(parts) > 6 {
				lat, err := strconv.ParseFloat(parts[5], 64)
				if err == nil {
					q300.Latitude = lat
				}

				lon, err := strconv.ParseFloat(parts[6], 64)
				if err == nil {
					q300.Longitude = lon
				}
			}

			// Parse reportID/reportType
			if len(parts) > 10 && msgCommand == "GTSPD" || msgCommand == "GTIGL" || msgCommand == "GTHBM" {
				if len(parts[10]) >= 2 {
					reportIDType := parts[10]
					if len(reportIDType) >= 1 {
						reportID, err := strconv.ParseInt(reportIDType[:1], 16, 64)
						if err == nil {
							q300.ReportID = fmt.Sprintf("%d", reportID)
						}
					}
					if len(reportIDType) >= 2 {
						reportType, err := strconv.ParseInt(reportIDType[1:], 16, 64)
						if err == nil {
							q300.ReportType = fmt.Sprintf("%d", reportType)
							q300.EventCode = getEventCode300(msgCommand, q300.ReportType)
						}
					}
				}
			}
		}
	}

	return q300, nil
}

// getEventCode300 returns the event code for Queclink 300 based on message type and status
func getEventCode300(msgType, status string) string {
	// This is a simplified version - in a real implementation, you would have a more
	// comprehensive mapping based on the Queclink protocol specification
	switch msgType {
	case "GTFRI":
		return "0" // Regular position report
	case "GTIGL":
		if status == "1" {
			return "1" // Ignition on
		}
		return "2" // Ignition off
	case "GTSPD":
		if status == "1" {
			return "3" // Speeding
		}
		return "4" // Speed back to normal
	case "GTSOS":
		return "5" // SOS
	case "GTRTL":
		return "6" // Geo-fence
	case "GTHBM":
		return "7" // Harsh behavior
	default:
		return "99" // Unknown event
	}
}
