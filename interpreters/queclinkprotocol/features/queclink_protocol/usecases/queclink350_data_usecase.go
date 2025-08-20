package usecases

import (
	"fmt"
	"queclinkprotocol/features/queclink_protocol/models"
	"strconv"
	"strings"
	"time"
)

// ParseQueclink350Fields parses data from a Queclink 350 device
func ParseQueclink350Fields(data string) (models.Queclink350Model, error) {
	// Initialize an empty model
	q350 := models.Queclink350Model{RawData: data}

	// Split the data by commas
	parts := strings.Split(data, ",")
	if len(parts) < 4 {
		return q350, fmt.Errorf("invalid Queclink 350 data: too few fields (%d)", len(parts))
	}

	// Get message type (e.g., +RESP, +BUFF)
	messageType := parts[0][:5]
	q350.MessageType = messageType

	// Extract device info
	if len(parts[1]) >= 2 {
		q350.DeviceType = parts[1][:2]
	}
	if len(parts[1]) >= 4 {
		version, err := strconv.ParseInt(parts[1][2:4], 16, 64)
		if err == nil {
			subVersion, err := strconv.ParseInt(parts[1][4:], 16, 64)
			if err == nil {
				if subVersion <= 9 {
					q350.DeviceVersion = fmt.Sprintf("%v.0%v", version, subVersion)
				} else {
					q350.DeviceVersion = fmt.Sprintf("%v.%v", version, subVersion)
				}
			}
		}
	}

	// IMEI and device name
	if len(parts) > 2 {
		q350.IMEI = parts[2]
	}
	if len(parts) > 3 {
		q350.DeviceName = parts[3]
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
					q350.Timestamp = timestamp
				}
			}

			// Parse coordinates
			if len(parts) > 6 {
				lat, err := strconv.ParseFloat(parts[5], 64)
				if err == nil {
					q350.Latitude = lat
				}

				lon, err := strconv.ParseFloat(parts[6], 64)
				if err == nil {
					q350.Longitude = lon
				}
			}

			// Parse speed, heading, etc.
			if len(parts) > 7 {
				speed, err := strconv.ParseFloat(parts[7], 64)
				if err == nil {
					q350.Speed = speed
				}
			}

			if len(parts) > 8 {
				heading, err := strconv.ParseFloat(parts[8], 64)
				if err == nil {
					q350.Heading = heading
				}
			}

			if len(parts) > 9 {
				altitude, err := strconv.ParseFloat(parts[9], 64)
				if err == nil {
					q350.Altitude = altitude
				}
			}

			// Parse satellites
			if len(parts) > 10 {
				satellites, err := strconv.Atoi(parts[10])
				if err == nil {
					q350.Satellites = satellites
				}
			}

			// Parse ignition status
			if len(parts) > 16 {
				ign, err := strconv.Atoi(parts[16])
				if err == nil {
					q350.Ignition = ign == 1
				}
			}

			// Parse external power and battery
			if len(parts) > 17 {
				externalPower, err := strconv.ParseInt(parts[17], 16, 64)
				if err == nil {
					q350.ExternalPower = float64(externalPower) / 1000.0
				}
			}

			if len(parts) > 18 {
				battery, err := strconv.ParseFloat(parts[18], 64)
				if err == nil {
					q350.BatteryLevel = battery
				}
			}

			// Set event code
			q350.EventCode = getEventCode350(msgCommand, "0")

			// Parse report ID and type if available
			if len(parts) > 13 && len(parts[13]) >= 2 {
				reportIDType := parts[13]
				if len(reportIDType) >= 1 {
					q350.ReportID = reportIDType[:1]
				}
				if len(reportIDType) >= 2 {
					q350.ReportType = reportIDType[1:]
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
					q350.Timestamp = timestamp
				}
			}

			// Parse coordinates
			if len(parts) > 6 {
				lat, err := strconv.ParseFloat(parts[5], 64)
				if err == nil {
					q350.Latitude = lat
				}

				lon, err := strconv.ParseFloat(parts[6], 64)
				if err == nil {
					q350.Longitude = lon
				}
			}

			// Parse report ID and type for specific messages
			if len(parts) > 10 && (msgCommand == "GTSPD" || msgCommand == "GTIGL" || msgCommand == "GTVGL" || msgCommand == "GTHBM") {
				if len(parts[10]) >= 2 {
					reportIDType := parts[10]
					if len(reportIDType) >= 1 {
						reportID, err := strconv.ParseInt(reportIDType[:1], 16, 64)
						if err == nil {
							q350.ReportID = fmt.Sprintf("%d", reportID)
						}
					}
					if len(reportIDType) >= 2 {
						reportType, err := strconv.ParseInt(reportIDType[1:], 16, 64)
						if err == nil {
							q350.ReportType = fmt.Sprintf("%d", reportType)
							q350.EventCode = getEventCode350(msgCommand, q350.ReportType)
						}
					}
				}
			}
		}
	}

	return q350, nil
}

// getEventCode350 returns the event code for Queclink 350 based on message type and report type
func getEventCode350(msgType, reportType string) string {
	// This is a simplified version - in a real implementation, you would have a more
	// comprehensive mapping based on the Queclink protocol specification
	switch msgType {
	case "GTFRI":
		return "0" // Regular position report
	case "GTSPD":
		if reportType == "1" {
			return "3" // Speeding
		}
		return "4" // Speed back to normal
	case "GTIGL":
		if reportType == "1" {
			return "1" // Ignition on
		}
		return "2" // Ignition off
	case "GTVGL":
		return "20" // Vehicle GPS Location
	case "GTHBM":
		return "21" // Harsh Behavior
	default:
		return "99" // Unknown event
	}
}
