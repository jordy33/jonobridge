package usecases

import (
	"encoding/json"
	"fmt"
	"time"
)

// GetDataJono converts Huabao protocol data to Jono protocol format
func GetDataJono(data string) (string, error) {
	// Parse the input Huabao data
	var huabaoData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &huabaoData); err != nil {
		return "", fmt.Errorf("error parsing Huabao data: %v", err)
	}

	// Create the Jono protocol structure
	jonoData := make(map[string]interface{})
	
	// Copy required fields
	jonoData["IMEI"] = getStringValue(huabaoData, "IMEI", "")
	jonoData["Message"] = getStringValue(huabaoData, "Message", "")
	jonoData["DataPackets"] = 1
	
	// Create packet data
	packets := make(map[string]interface{})
	
	// Check if we have ListPackets in the input data (from parsed Huabao)
	if listPackets, ok := huabaoData["ListPackets"].(map[string]interface{}); ok {
		// Use the existing packet structure from parsed Huabao data
		packetCount := 1
		for _, packetData := range listPackets {
			if packetMap, ok := packetData.(map[string]interface{}); ok {
				packet := make(map[string]interface{})
				
				// Copy location data with proper handling - these should already be calculated
				if lat, exists := packetMap["Latitude"]; exists && lat != nil {
					packet["Latitude"] = lat
				} else {
					packet["Latitude"] = nil
				}
				
				if lon, exists := packetMap["Longitude"]; exists && lon != nil {
					packet["Longitude"] = lon
				} else {
					packet["Longitude"] = nil
				}
				
				// Copy other existing fields or set defaults
				packet["Altitude"] = getValueOrDefault(packetMap, "Altitude", 0)
				
				// Process datetime
				if dt, exists := packetMap["Datetime"]; exists && dt != nil {
					packet["Datetime"] = dt
				} else {
					packet["Datetime"] = time.Now().UTC().Format(time.RFC3339)
				}
				
				// Process event code
				if ec, exists := packetMap["EventCode"]; exists {
					packet["EventCode"] = ec
				} else {
					packet["EventCode"] = map[string]interface{}{
						"Code": 35,
						"Name": "Track By Time Interval",
					}
				}
				
				// Copy speed
				if speed, exists := packetMap["Speed"]; exists && speed != nil {
					packet["Speed"] = speed
				} else {
					packet["Speed"] = 0
				}
				
				// Copy direction
				if dir, exists := packetMap["Direction"]; exists && dir != nil {
					packet["Direction"] = dir
				} else {
					packet["Direction"] = 0
				}
				
				// Set positioning status
				if ps, exists := packetMap["PositioningStatus"]; exists && ps != nil {
					if psStr, ok := ps.(string); ok {
						if psStr == "A" {
							packet["PositioningStatus"] = "true"
						} else {
							packet["PositioningStatus"] = psStr
						}
					} else {
						packet["PositioningStatus"] = ps
					}
				} else {
					packet["PositioningStatus"] = "true"
				}
				
				// Copy or set other fields with defaults
				packet["NumberOfSatellites"] = getValueOrDefault(packetMap, "NumberOfSatellites", 0)
				packet["GSMSignalStrength"] = getValueOrDefault(packetMap, "GSMSignalStrength", 0)
				packet["HDOP"] = getValueOrDefault(packetMap, "HDOP", 0)
				packet["Mileage"] = getValueOrDefault(packetMap, "Mileage", 0)
				packet["RunTime"] = getValueOrDefault(packetMap, "RunTime", 0)
				
				// Handle IoPortStatus
				if ioStatus, exists := packetMap["IoPortStatus"]; exists && ioStatus != nil {
					packet["IoPortStatus"] = ioStatus
				} else {
					packet["IoPortStatus"] = convertIoPortStatus("")
				}
				
				// Handle AnalogInputs
				if ai, exists := packetMap["AnalogInputs"]; exists && ai != nil {
					packet["AnalogInputs"] = ai
				} else {
					packet["AnalogInputs"] = map[string]interface{}{
						"AD1": nil, "AD2": nil, "AD3": nil, "AD4": nil, "AD5": nil,
						"AD6": nil, "AD7": nil, "AD8": nil, "AD9": nil, "AD10": nil,
					}
				}
				
				// Set BaseStationInfo
				if bsi, exists := packetMap["BaseStationInfo"]; exists {
					packet["BaseStationInfo"] = bsi
				} else {
					packet["BaseStationInfo"] = nil
				}
				
				// Add all additional fields with default null values
				packet["AdditionalAlertInfoADASDMS"] = getValueOrDefault(packetMap, "AdditionalAlertInfoADASDMS", map[string]interface{}{
					"AlarmProtocol": nil, "AlarmType": nil, "PhotoName": nil,
				})
				packet["BluetoothBeaconA"] = getValueOrDefault(packetMap, "BluetoothBeaconA", map[string]interface{}{
					"Version": nil, "DeviceName": nil, "MAC": nil, "BatteryPower": nil, "SignalStrength": nil,
				})
				packet["BluetoothBeaconB"] = getValueOrDefault(packetMap, "BluetoothBeaconB", map[string]interface{}{
					"Version": nil, "DeviceName": nil, "MAC": nil, "BatteryPower": nil, "SignalStrength": nil,
				})
				packet["CameraStatus"] = getValueOrDefault(packetMap, "CameraStatus", map[string]interface{}{
					"CameraNumber": nil, "Status": nil,
				})
				packet["CurrentNetworkInfo"] = getValueOrDefault(packetMap, "CurrentNetworkInfo", map[string]interface{}{
					"Version": nil, "Type": nil, "Descriptor": nil,
				})
				packet["FatigueDrivingInformation"] = getValueOrDefault(packetMap, "FatigueDrivingInformation", map[string]interface{}{
					"Version": nil, "Type": nil, "Descriptor": nil,
				})
				packet["InputPortStatus"] = getValueOrDefault(packetMap, "InputPortStatus", map[string]interface{}{
					"Input1": nil, "Input2": nil, "Input3": nil, "Input4": nil,
					"Input5": nil, "Input6": nil, "Input7": nil, "Input8": nil,
				})
				packet["OutputPortStatus"] = getValueOrDefault(packetMap, "OutputPortStatus", map[string]interface{}{
					"Output1": nil, "Output2": nil, "Output3": nil, "Output4": nil,
					"Output5": nil, "Output6": nil, "Output7": nil, "Output8": nil,
				})
				packet["SystemFlag"] = getValueOrDefault(packetMap, "SystemFlag", map[string]interface{}{
					"EEP2": nil, "ACC": nil, "AntiTheft": nil, "VibrationFlag": nil,
					"MovingFlag": nil, "ExternalPowerSupply": nil, "Charging": nil, 
					"SleepMode": nil, "FMS": nil, "FMSFunction": nil, "SystemFlagExtras": nil,
				})
				packet["TemperatureSensor"] = getValueOrDefault(packetMap, "TemperatureSensor", map[string]interface{}{
					"SensorNumber": nil, "Value": nil,
				})
				packet["TemperatureAndHumiditySensor"] = getValueOrDefault(packetMap, "TemperatureAndHumiditySensor", map[string]interface{}{
					"DeviceName": nil, "MAC": nil, "BatteryPower": nil, "Temperature": nil, 
					"Humidity": nil, "AlertHighTemperature": nil, "AlertLowTemperature": nil,
					"AlertHighHumidity": nil, "AlertLowHumidity": nil,
				})
				
				// Use packet_1, packet_2, etc. format
				packetKey := fmt.Sprintf("packet_%d", packetCount)
				packets[packetKey] = packet
				packetCount++
			}
		}
	} else {
		// Fallback to original logic if no ListPackets found
		packet := make(map[string]interface{})
		
		// Copy location data
		packet["Latitude"] = getFloatValue(huabaoData, "Latitude", 0)
		packet["Longitude"] = getFloatValue(huabaoData, "Longitude", 0)
		
		// Process and format datetime
		datetimeStr := getStringValue(huabaoData, "Datetime", "")
		if datetimeStr == "" {
			if datetime, ok := huabaoData["Datetime"].(time.Time); ok {
				packet["Datetime"] = datetime.Format(time.RFC3339)
			} else {
				packet["Datetime"] = time.Now().UTC().Format(time.RFC3339)
			}
		} else {
			packet["Datetime"] = datetimeStr
		}
		
		// Copy event code as a structured object
		eventCodeVal, ok := huabaoData["EventCode"]
		if ok {
			eventCode := 0
			eventCodeName := ""
			
			switch v := eventCodeVal.(type) {
			case float64:
				eventCode = int(v)
			case int:
				eventCode = v
			case map[string]interface{}:
				if codeVal, codeOk := v["Code"]; codeOk {
					if codeFloat, isFloat := codeVal.(float64); isFloat {
						eventCode = int(codeFloat)
					} else if codeInt, isInt := codeVal.(int); isInt {
						eventCode = codeInt
					}
				}
				if nameVal, nameOk := v["Name"]; nameOk {
					if nameStr, isStr := nameVal.(string); isStr {
						eventCodeName = nameStr
					}
				}
			}
			
			if eventCodeName == "" {
				switch eventCode {
				case 35:
					eventCodeName = "Track By Time Interval"
				case 101:
					eventCodeName = "Data Report"
				case 142:
					eventCodeName = "Status Report"
				default:
					eventCodeName = fmt.Sprintf("Event %d", eventCode)
				}
			}
			
			packet["EventCode"] = map[string]interface{}{
				"Code": eventCode,
				"Name": eventCodeName,
			}
		} else {
			packet["EventCode"] = map[string]interface{}{
				"Code": 35,
				"Name": "Track By Time Interval",
			}
		}
		
		// Copy speed
		speed := getFloatValue(huabaoData, "Speed", 0)
		if speed == 0 {
			speed = getFloatValue(huabaoData, "Heading", 0)
		}
		packet["Speed"] = int(speed)
		
		// Set Direction/Azimuth field
		direction := getFloatValue(huabaoData, "Direction", 0)
		if direction == 0 {
			direction = getFloatValue(huabaoData, "Heading", 0)
		}
		packet["Direction"] = int(direction)
		
		// Add altitude
		packet["Altitude"] = int(getFloatValue(huabaoData, "Altitude", 0))
		
		// Set PositioningStatus
		posStatus := getStringValue(huabaoData, "PositioningStatus", "A")
		if posStatus == "A" {
			packet["PositioningStatus"] = "true"
		} else {
			packet["PositioningStatus"] = posStatus
		}
		
		// Set IoPortStatus as a structured object
		ioPortStatus := getStringValue(huabaoData, "IoPortStatus", "")
		packet["IoPortStatus"] = convertIoPortStatus(ioPortStatus)
		
		// Copy analog inputs if present
		if analogInputs, ok := huabaoData["AnalogInputs"].(map[string]interface{}); ok {
			packet["AnalogInputs"] = analogInputs
		} else {
			packet["AnalogInputs"] = map[string]interface{}{
				"AD1": nil, "AD2": nil, "AD3": nil, "AD4": nil, "AD5": nil,
				"AD6": nil, "AD7": nil, "AD8": nil, "AD9": nil, "AD10": nil,
			}
		}
		
		// Add missing fields with null values to match complete structure
		packet["NumberOfSatellites"] = getIntValue(huabaoData, "NumberOfSatellites", 0)
		packet["GSMSignalStrength"] = getIntValue(huabaoData, "GsmSignalStrength", 0)
		packet["HDOP"] = getFloatValue(huabaoData, "Hdop", 0)
		packet["Mileage"] = getIntValue(huabaoData, "Mileage", 0)
		packet["RunTime"] = getIntValue(huabaoData, "RunTime", 0)
		
		// Set up base station info
		packet["BaseStationInfo"] = nil
		if bsInfo, ok := huabaoData["BaseStationInfo"].(map[string]interface{}); ok {
			packet["BaseStationInfo"] = bsInfo
		}
		
		// Add all additional fields with null values
		packet["AdditionalAlertInfoADASDMS"] = map[string]interface{}{
			"AlarmProtocol": nil, "AlarmType": nil, "PhotoName": nil,
		}
		packet["BluetoothBeaconA"] = map[string]interface{}{
			"Version": nil, "DeviceName": nil, "MAC": nil, "BatteryPower": nil, "SignalStrength": nil,
		}
		packet["BluetoothBeaconB"] = map[string]interface{}{
			"Version": nil, "DeviceName": nil, "MAC": nil, "BatteryPower": nil, "SignalStrength": nil,
		}
		packet["CameraStatus"] = map[string]interface{}{
			"CameraNumber": nil, "Status": nil,
		}
		packet["CurrentNetworkInfo"] = map[string]interface{}{
			"Version": nil, "Type": nil, "Descriptor": nil,
		}
		packet["FatigueDrivingInformation"] = map[string]interface{}{
			"Version": nil, "Type": nil, "Descriptor": nil,
		}
		packet["InputPortStatus"] = map[string]interface{}{
			"Input1": nil, "Input2": nil, "Input3": nil, "Input4": nil,
			"Input5": nil, "Input6": nil, "Input7": nil, "Input8": nil,
		}
		packet["OutputPortStatus"] = map[string]interface{}{
			"Output1": nil, "Output2": nil, "Output3": nil, "Output4": nil,
			"Output5": nil, "Output6": nil, "Output7": nil, "Output8": nil,
		}
		packet["SystemFlag"] = map[string]interface{}{
			"EEP2": nil, "ACC": nil, "AntiTheft": nil, "VibrationFlag": nil,
			"MovingFlag": nil, "ExternalPowerSupply": nil, "Charging": nil, 
			"SleepMode": nil, "FMS": nil, "FMSFunction": nil, "SystemFlagExtras": nil,
		}
		packet["TemperatureSensor"] = map[string]interface{}{
			"SensorNumber": nil, "Value": nil,
		}
		packet["TemperatureAndHumiditySensor"] = map[string]interface{}{
			"DeviceName": nil, "MAC": nil, "BatteryPower": nil, "Temperature": nil, 
			"Humidity": nil, "AlertHighTemperature": nil, "AlertLowTemperature": nil,
			"AlertHighHumidity": nil, "AlertLowHumidity": nil,
		}
		
		packets["packet_1"] = packet
	}
	
	// Add packets to Jono data
	jonoData["ListPackets"] = packets
	
	// Convert to JSON
	result, err := json.Marshal(jonoData)
	if err != nil {
		return "", fmt.Errorf("error marshaling Jono data: %v", err)
	}
	
	return string(result), nil
}

// Helper function to convert IoPortStatus string to structured format
func convertIoPortStatus(ioStatus string) map[string]interface{} {
	// Always return the proper structure even if we can't parse the input
	result := map[string]interface{}{
		"Port1": 0, "Port2": 0, "Port3": 0, "Port4": 0,
		"Port5": 0, "Port6": 0, "Port7": 0, "Port8": 0,
	}
	
	// In a real implementation, we would parse the bits from ioStatus
	// For now, just ensure we return the expected structure
	return result
}

// Helper functions to safely extract values from the map
func getStringValue(data map[string]interface{}, key, defaultVal string) string {
	if val, ok := data[key]; ok {
		if strVal, isStr := val.(string); isStr {
			return strVal
		}
	}
	return defaultVal
}

func getFloatValue(data map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := parseFloat(v); err == nil {
				return f
			}
		}
	}
	return defaultVal
}

func getIntValue(data map[string]interface{}, key string, defaultVal int) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := parseInt(v); err == nil {
				return i
			}
		}
	}
	return defaultVal
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// Helper function to get value or default
func getValueOrDefault(data map[string]interface{}, key string, defaultVal interface{}) interface{} {
	if val, exists := data[key]; exists && val != nil {
		return val
	}
	return defaultVal
}
