package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"pinoprotocol/features/jono"
	"pinoprotocol/features/pino_protocol/models"
	"pinoprotocol/features/pino_protocol/usecases"
	"sync"
	"time"

	"github.com/MaddSystems/jonobridge/common/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func init() {
	// Initialize the flag using BoolVar instead of Bool

}

var imeiStore sync.Map
var deviceDataCache sync.Map // Cache to store the latest information for each device

// DeviceData represents the latest known good data for a device
type DeviceData struct {
	VoltageValue      float64                     // Battery voltage from heartbeat
	GSMSignalStrength int                         // GSM signal strength from heartbeat
	LastUpdated       time.Time                   // When this data was last updated
	LocationData      *models.LocationPacketModel // Last known good location data
}

type TrackerData struct {
	Payload    string `json:"payload"`
	RemoteAddr string `json:"remoteaddr"`
}

type TrackerAssign struct {
	Imei       string `json:"imei"`
	Protocol   string `json:"protocol"`
	RemoteAddr string `json:"remoteaddr"`
}

// MQTTClient wraps the MQTT client with additional functionality
type MQTTClient struct {
	client    mqtt.Client
	brokerURL string
	clientID  string
}

// NewMQTTClient creates a new MQTT client with the given configuration
func NewMQTTClient(brokerHost string, clientID string) (*MQTTClient, error) {
	if brokerHost == "" {
		return nil, fmt.Errorf("broker host cannot be empty")
	}

	brokerURL := fmt.Sprintf("tcp://%s:1883", brokerHost)
	return &MQTTClient{
		brokerURL: brokerURL,
		clientID:  clientID,
	}, nil
}

// Connect establishes a connection to the MQTT broker
func (m *MQTTClient) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.brokerURL)

	// Generate a unique client ID
	subscribe_topic := "pino" // You might want to make this configurable
	clientID := fmt.Sprintf("pinoprotocol_%s_%s_%d",
		subscribe_topic,
		os.Getenv("HOSTNAME"),
		time.Now().UnixNano()%100000)
	opts.SetClientID(clientID)

	// Configure settings for multiple listeners
	opts.SetCleanSession(false) // Maintain persistent session
	opts.SetAutoReconnect(true) // Auto reconnect on connection loss
	opts.SetKeepAlive(60 * time.Second)
	opts.SetOrderMatters(true) // Maintain message order
	opts.SetResumeSubs(true)   // Resume stored subscriptions
	opts.SetDefaultPublishHandler(m.messageHandler)

	m.client = mqtt.NewClient(opts)
	utils.VPrint("MQTT client created with ID: %s", clientID)
	// Try to connect with retries
	for {
		if token := m.client.Connect(); token.Wait() && token.Error() != nil {
			utils.VPrint("Error connecting to MQTT broker at %s: %v. Retrying in 5 seconds...", m.brokerURL, token.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		utils.VPrint("Successfully connected to the MQTT broker")
		break
	}

	return nil
}

// Subscribe subscribes to the specified topic
func (m *MQTTClient) Subscribe(topic string, qos byte) error {
	if token := m.client.Subscribe(topic, qos, nil); token.Wait() && token.Error() != nil {
		return fmt.Errorf("error subscribing to topic %s: %v", topic, token.Error())
	}
	utils.VPrint("Subscribed to topic: %s", topic)
	return nil
}

// Publish publishes a message to the specified topic
func (m *MQTTClient) Publish(topic string, payload interface{}) error {
	token := m.client.Publish(topic, 0, false, payload)
	token.Wait()
	if token.Error() != nil {
		utils.VPrint("Failed to publish to MQTT topic %s: %v", topic, token.Error())
		return token.Error()
	}

	utils.VPrint("Successfully published to MQTT topic: %s", topic)
	return nil
}

// messageHandler handles incoming MQTT messages
func (m *MQTTClient) messageHandler(client mqtt.Client, msg mqtt.Message) {
	// Setup logging

	// Don't print full message payload
	utils.VPrint("Received message on topic %s", msg.Topic())
	var json_data TrackerData
	if err := json.Unmarshal(msg.Payload(), &json_data); err != nil {
		utils.VPrint("Error unmarshaling JSON: %v", err)
		return
	}

	// Convert hex string to bytes
	rawBytes, err := hex.DecodeString(json_data.Payload)
	if err != nil {
		utils.VPrint("Error decoding hex string:%v", err)
		return
	}

	trackerPayload := string(msg.Payload())
	utils.VPrint("json:%v", trackerPayload)
	clientAddr := json_data.RemoteAddr
	if len(rawBytes) < 1 || (rawBytes[0] != 0x7E && rawBytes[0] != 0x78) {
		utils.VPrint("invalid frame: first byte is not 0x7E or 0x78")
	}

	if rawBytes[0] == 0x7E {
		utils.VPrint("BSJ TRACKER")
		rawBytes = rawBytes[1 : len(rawBytes)-1]

		data := rawBytes[:len(rawBytes)-1]
		checksum := rawBytes[len(rawBytes)-1]

		calculatedChecksum := usecases.CalculateChecksum(data)
		if calculatedChecksum != checksum {
			log.Printf("Checksum inválido: calculado 0x%X, recibido 0x%X", calculatedChecksum, checksum)
		}

		messageID := rawBytes[:2]                                          // ID del mensaje
		phoneNumber := usecases.DecodeTerminalMobileNumber(rawBytes[4:10]) // Teléfono (BCD)
		serialNumber := rawBytes[10:12]                                    // Número de serie
		utils.VPrint("Phone number BSJ: %s", phoneNumber)
		utils.VPrint("Numero de serie BSJ: %s", serialNumber)

		// Use the phone number as a device identifier initially
		imei := phoneNumber
		utils.VPrint("Phone number used as initial device ID: %s", imei)

		// For location data, check if we can extract the actual IMEI from extended data
		if bytes.Equal(messageID, []byte{0x02, 0x00}) { // Location packet
			locationData := rawBytes[12:] // Body of the location packet
			extendedData := usecases.ParseExtendedDataForIMEI(locationData)

			// If we found an IMEI in the extended data (ID 0x00D5), use it
			if extendedImei, ok := extendedData["IMEI"].(string); ok && extendedImei != "" {
				imei = extendedImei
				utils.VPrint("Using IMEI from extended data: %s", imei)
			}
		}

		utils.VPrint("Final IMEI (BSJ protocol): %s", imei)

		imeiStore.Store(clientAddr, imei)
		if bytes.Equal(rawBytes[:2], []byte{0x01, 0x00}) { // Registro
			log.Printf("Registro recibido. Teléfono: %s, Serial: %X\n", phoneNumber, serialNumber)

			response := usecases.GenerateRegistrationResponse(serialNumber, phoneNumber)
			tracker_data_json := TrackerData{
				Payload:    response,
				RemoteAddr: clientAddr,
			}
			response_json, err := json.Marshal(tracker_data_json)
			if err != nil {
				utils.VPrint("Error creating JSON:%v", err)
				return
			}

			client.Publish("tracker/send", 0, false, response_json)
		} else if bytes.Equal(messageID, []byte{0x01, 0x02}) { // Autenticación
			log.Printf("Autenticación recibida. Teléfono: %s, Serial: %X\n", phoneNumber, serialNumber)

			authCode := rawBytes[12 : len(rawBytes)-1]
			log.Printf("Auth Code recibido BSJ: %s", string(authCode))

			response := usecases.GenerateAuthenticationResponse(serialNumber, phoneNumber)
			tracker_data_json := TrackerData{
				Payload:    response,
				RemoteAddr: clientAddr,
			}
			response_json, err := json.Marshal(tracker_data_json)
			if err != nil {
				utils.VPrint("Error creating JSON:%v", err)
				return
			}

			client.Publish("tracker/send", 0, false, response_json)
		} else if bytes.Equal(messageID, []byte{0x00, 0x02}) { // Terminal heartbeat
			log.Printf("Heartbeat recibido. Teléfono: %s, Trama num. Serie: %X\n", phoneNumber, serialNumber)

			// Parse heartbeat message according to BSJ-EG01 protocol
			if len(rawBytes) < 15 { // Ensure we have enough data (12 header + 3 heartbeat data + checksum)
				utils.VPrint("Invalid heartbeat message: insufficient data")
				return
			}

			// Extract heartbeat data
			batteryPower := int(rawBytes[12]) // Battery percentage (0-100)
			csqValue := int(rawBytes[13])     // Signal strength
			deviceStatus := int(rawBytes[14]) // 0: operating mode, 1: standby, 2: turn off

			// Log the heartbeat data
			utils.VPrint("BSJ Heartbeat - Battery: %d%%, Signal: %d, Status: %d",
				batteryPower, csqValue, deviceStatus)

			// Store the device data in cache for future use
			// Convert battery percentage to voltage estimate (rough approximation)
			estimatedVoltage := float64(batteryPower) / 100.0 * 13.0 // Maximum voltage around 13V
			cacheDeviceData(imei, estimatedVoltage, csqValue)

			// Create and send the response
			response := usecases.GenerateHeartbeatResponse(serialNumber, phoneNumber)
			tracker_data_json := TrackerData{
				Payload:    response,
				RemoteAddr: clientAddr,
			}
			response_json, err := json.Marshal(tracker_data_json)
			if err != nil {
				utils.VPrint("Error creating JSON:%v", err)
				return
			}

			client.Publish("tracker/send", 0, false, response_json)
		} else if bytes.Equal(messageID, []byte{0x02, 0x00}) { // Localización BSJ
			log.Printf("Trama de localización recibida. Teléfono: %s, Serial: %X\n", phoneNumber, serialNumber)
			// Parsear datos de localización
			locationData := rawBytes[12:] // Cuerpo de la trama de localización
			imeiValue, ok := imeiStore.Load(clientAddr)
			if !ok {
				utils.VPrint("Error: IMEI not found for client %s", clientAddr)
				return
			}
			imei, ok := imeiValue.(string)
			if !ok {
				utils.VPrint("Error: Invalid IMEI type for client %s", clientAddr)
				return
			}

			// First parse location data
			parsedLocationData := usecases.ParseLocationData(locationData, imei, rawBytes)

			// Then convert to Jono format
			jonoNormalize, err := jono.Initialize(parsedLocationData)
			if err != nil {
				utils.VPrint("Error transforming to jono format: %v", err)
				return
			}

			// Verify and ensure IMEI appears in final output
			var jonoJson map[string]interface{}
			if err := json.Unmarshal([]byte(jonoNormalize), &jonoJson); err == nil {
				if jonoImei, exists := jonoJson["IMEI"]; exists {
					utils.VPrint("DEBUG BSJ: Final Jono output contains IMEI: %v", jonoImei)
				} else {
					// Force include the IMEI in the final output
					jonoJson["IMEI"] = imei
					if updatedData, err := json.Marshal(jonoJson); err == nil {
						jonoNormalize = string(updatedData)
						utils.VPrint("DEBUG BSJ: Added missing IMEI to final output: %s", imei)
					}
				}
			}
			compactJSON := new(bytes.Buffer)
			if err := json.Compact(compactJSON, []byte(jonoNormalize)); err != nil {
				utils.VPrint("Error compacting JSON: %v", err)
			} else {
				utils.VPrint("Jono Protocol: %s", compactJSON.String())
			}
			// Publish to jonoprotocol topic
			if err := m.Publish("tracker/jonoprotocol", jonoNormalize); err != nil {
				utils.VPrint("Error publishing to jonoprotocol: %v", err)
			}

			// IMPORTANT: Keep this code to publish IMEI assignment
			tracker_data_json := TrackerAssign{
				Imei:       imei,
				Protocol:   "pino",
				RemoteAddr: clientAddr,
			}
			assignImeiJson, err := json.Marshal(tracker_data_json)
			if err != nil {
				fmt.Println("Error marshaling assign-imei data:", err)
				return
			}
			// Publish to assign-imei2remoteaddr topic
			if err := m.Publish("tracker/assign-imei2remoteaddr", assignImeiJson); err != nil {
				utils.VPrint("Error publishing to assign-imei2remoteaddr: %v", err)
			}
			utils.VPrint("Publishing to tracker/assign-imei2remoteaddr: %s", assignImeiJson)
		} else {
			// Handle unknown message ID
			log.Printf("Message ID no implementado: %X", messageID)
			/*
				// Still publish the IMEI to assign-imei2remoteaddr if we have it
				tracker_data_json := TrackerAssign{
					Imei:       imei,
					Protocol:   "pino",
					RemoteAddr: clientAddr,
				}
				assignImeiJson, err := json.Marshal(tracker_data_json)
				if err != nil {
					fmt.Println("Error marshaling assign-imei data:", err)
					return
				}
				// Publish to assign-imei2remoteaddr topic

				if err := m.Publish("tracker/assign-imei2remoteaddr", assignImeiJson); err != nil {
					utils.VPrint("Error publishing to assign-imei2remoteaddr: %v", err)
				}
				utils.VPrint("Publishing to tracker/assign-imei2remoteaddr: %s", assignImeiJson)
			*/
		}
	} else if rawBytes[0] == 0x78 { // GT06 protocol
		utils.VPrint("GT06 TRACKER")
		switch {
		case usecases.IsLoginPacket(rawBytes):
			imei, err := usecases.ExtractIMEI(rawBytes)
			if err != nil {
				utils.VPrint("Error extracting IMEI: %v", err)
			}
			// Store the IMEI in the map
			imeiStore.Store(clientAddr, imei)
			response := hex.EncodeToString(usecases.BuildLoginResponse(rawBytes))
			tracker_data_json := TrackerData{
				Payload:    response,
				RemoteAddr: clientAddr,
			}
			response_json, err := json.Marshal(tracker_data_json)
			if err != nil {
				utils.VPrint("Error creating JSON:%v", err)
				return
			}
			client.Publish("tracker/send", 0, false, response_json)

		case usecases.IsStandardLocationPacket(rawBytes): // GT06 location packet
			utils.VPrint("Processing GT06 location packet")
			// Retrieve IMEI from the map
			imeiValue, ok := imeiStore.Load(clientAddr)
			if !ok {
				utils.VPrint("Error: IMEI not found for client %s", clientAddr)
				return
			}
			imei, ok := imeiValue.(string)
			if !ok {
				utils.VPrint("Error: Invalid IMEI type for client %s", clientAddr)
				return
			}
			data, err := usecases.DecodeStandardLocationData(rawBytes, imei, false)
			if err != nil {
				utils.VPrint("error decoding location data gt06, %s", err)
				return
			}

			// Debug the battery level before conversion to JSON
			utils.VPrint("Battery level detected: %d", data.BatteryLevel)

			// Enhance location data with cached device information
			enhanceLocationDataWithCache(imei, data)

			// Also explicitly cache this location data for future use
			cacheLocationData(imei, data)

			// Add the voltage value directly to the location map for Jono to pick it up
			var jonoNormalize string // Declare jonoNormalize here for use throughout the function

			if jsonBytes, err := json.Marshal(data); err == nil {
				var locationMap map[string]interface{}
				if err := json.Unmarshal(jsonBytes, &locationMap); err == nil {
					// Ensure VoltageValue is included in the JSON
					if data.VoltageValue > 0 {
						locationMap["VoltageValue"] = data.VoltageValue

						// Re-encode with the added field
						if enhancedJSON, err := json.Marshal(locationMap); err == nil {
							// Pass the enhanced decoded data to jono for normalization
							enhancedStr := string(enhancedJSON)
							var normalizeErr error
							jonoNormalize, normalizeErr = jono.Initialize(enhancedStr)
							if normalizeErr != nil {
								utils.VPrint("error decoding gt06 to jono, %s", normalizeErr)
								return
							}
						}
					} else {
						// If no voltage, use the original JSON
						jonoNormalize, err = jono.Initialize(string(jsonBytes))
						if err != nil {
							utils.VPrint("error decoding gt06 to jono, %s", err)
							return
						}
					}
				} else {
					// Fallback to original data
					var err error
					jonoNormalize, err = jono.Initialize(string(jsonBytes))
					if err != nil {
						utils.VPrint("error decoding gt06 to jono, %s", err)
						return
					}
				}
			} else {
				utils.VPrint("error marshalling location data, %s", err)
				return
			}

			utils.VPrint("Jono Protocol processed successfully (payload omitted for brevity)")

			// Publish to jonoprotocol topic
			if err := m.Publish("tracker/jonoprotocol", jonoNormalize); err != nil {
				utils.VPrint("Error publishing to jonoprotocol: %v", err)
			}

			tracker_data_json := TrackerAssign{
				Imei:       imei,
				Protocol:   "pino",
				RemoteAddr: clientAddr,
			}
			// Marshal the data to JSON before publishing
			assignImeiJson, err := json.Marshal(tracker_data_json)
			if err != nil {
				fmt.Println("Error marshaling assign-imei data:", err)
				return
			}
			// Publish to assign-imei2remoteaddr topic
			if err := m.Publish("tracker/assign-imei2remoteaddr", assignImeiJson); err != nil {
				utils.VPrint("Error publishing to assign-imei2remoteaddr: %v", err)
			}
			utils.VPrint("Publishing to tracker/assign-imei2remoteaddr: %s", assignImeiJson)

		case usecases.IsStandardAlarmPacket(rawBytes):
			imeiValue, ok := imeiStore.Load(clientAddr)
			if !ok {
				utils.VPrint("Error: IMEI not found for client %s", clientAddr)
				return
			}
			imei, ok := imeiValue.(string)
			if !ok {
				utils.VPrint("Error: Invalid IMEI type for client %s", clientAddr)
				return
			}

			// First decode the alarm frame
			data, err := usecases.DecodeAlarmFrame(rawBytes, imei)
			if err != nil {
				utils.VPrint("error decoding alarm data, %s", err)
				return
			}

			// Extract terminal information byte (byte index 4)
			terminalInfoByte := rawBytes[4]
			oilDisconnected, gpsTrackingOn, eventCode, chargeOn, accHigh, activated := usecases.DecodeTerminalInformationBits(terminalInfoByte)

			// Calculate voltage from voltageLevelByte (byte index 5)
			voltageLevelByte := rawBytes[5]
			utils.VPrint("Voltage Level Byte: %v", voltageLevelByte)
			var voltageValue float64
			switch voltageLevelByte {
			case 0:
				voltageValue = 0.0 // No Power
			case 1:
				voltageCalc := (3.0 * 1024.0) / 6.0
				voltageValue = float64(int(math.Round(voltageCalc)))
			case 2:
				voltageCalc := (6.0 * 1024.0) / 6.0
				voltageValue = float64(int(math.Round(voltageCalc)))
			case 3:
				voltageCalc := (9.0 * 1024.0) / 6.0
				voltageValue = float64(int(math.Round(voltageCalc)))
			case 4:
				voltageCalc := (12.0 * 1024.0) / 6.0
				voltageValue = float64(int(math.Round(voltageCalc)))
			case 5:
				voltageCalc := (12.5 * 1024.0) / 6.0
				voltageValue = float64(int(math.Round(voltageCalc)))
			case 6:
				voltageCalc := (13.0 * 1024.0) / 6.0
				voltageValue = float64(int(math.Round(voltageCalc)))
			default:
				voltageValue = 9.0 // Default to something reasonable
			}

			// Log the alarm data
			voltageLevel := data.VoltageLevel
			gsmSignal := data.GSMSignalStrength

			utils.VPrint("ALARM DETECTED - EventCode: %d", eventCode)
			utils.VPrint("Terminal Info: Oil=%v, GPS=%v, Charge=%v, ACC=%v, Active=%v",
				map[bool]string{true: "Disconnected", false: "Connected"}[oilDisconnected],
				map[bool]string{true: "On", false: "Off"}[gpsTrackingOn],
				chargeOn,
				map[bool]string{true: "High", false: "Low"}[accHigh],
				map[bool]string{true: "Yes", false: "No"}[activated])
			utils.VPrint("Voltage Level: %s (%.1f V)", voltageLevel, voltageValue)
			utils.VPrint("GSM Signal Strength: %s", gsmSignal)

			// Log location data
			locModel := data.LocationPacketModel
			if locModel != nil {
				utils.VPrint("Location - Lat: %f, Lon: %f, DateTime: %s",
					locModel.Latitude, locModel.Longitude, locModel.DateTime)
			}

			// Convert to JSON for Jono processing
			jsonData, err := data.ToJSON()
			if err != nil {
				utils.VPrint("error converting alarm data to JSON: %s", err)
				return
			}

			// Parse the JSON and add alarm specific fields including voltage
			var alarmMap map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &alarmMap); err == nil {
				// Add voltage to locationPacketModel
				if locModel, ok := alarmMap["locationPacketModel"].(map[string]interface{}); ok {
					locModel["VoltageValue"] = voltageValue
				}

				// Add other alarm fields
				alarmMap["Message"] = "Alarm event"
				alarmMap["EventCode"] = fmt.Sprintf("%d", eventCode)

				// Add alarm type based on event code
				alarmType := ""
				switch eventCode {
				case 1:
					alarmType = "SOS"
				case 23:
					alarmType = "Power Cut Alarm"
				case 50:
					alarmType = "Alarm"
				case 79:
					alarmType = "Shock Alarm"
				case 35:
					alarmType = "Normal"
				}

				if alarmType != "" {
					alarmMap["AlarmType"] = alarmType
					alarmMap["Message"] = fmt.Sprintf("Alarm event: %s", alarmType)
				}

				if enhancedJSON, err := json.Marshal(alarmMap); err == nil {
					jsonData = string(enhancedJSON)
				}
			}

			// Transform to Jono format
			jonoNormalize, err := jono.Initialize(jsonData)
			if err != nil {
				utils.VPrint("error transforming alarm to jono format: %s", err)
				return
			}

			utils.VPrint("Alarm Jono Protocol GT06 processed successfully")

			// Publish to jonoprotocol topic
			if err := m.Publish("tracker/jonoprotocol", jonoNormalize); err != nil {
				utils.VPrint("Error publishing to jonoprotocol: %v", err)
			}

			// Handle tracker assignment
			tracker_data_json := TrackerAssign{
				Imei:       imei,
				Protocol:   "pino",
				RemoteAddr: clientAddr,
			}
			assignImeiJson, err := json.Marshal(tracker_data_json)
			if err != nil {
				fmt.Println("Error marshaling assign-imei data:", err)
				return
			}
			if err := m.Publish("tracker/assign-imei2remoteaddr", assignImeiJson); err != nil {
				utils.VPrint("Error publishing to assign-imei2remoteaddr: %v", err)
			}
			utils.VPrint("Publishing to tracker/assign-imei2remoteaddr: %s", assignImeiJson)

		case usecases.IsHeartbeatPacket(rawBytes):
			// Retrieve IMEI from the map
			imeiValue, ok := imeiStore.Load(clientAddr)
			if !ok {
				utils.VPrint("Error: IMEI not found for client %s", clientAddr)
				return
			}
			imei, ok := imeiValue.(string)
			if !ok {
				utils.VPrint("Error: Invalid IMEI type for client %s", clientAddr)
				return
			}

			// Decode heartbeat packet and print debug info
			statusData, err := usecases.DecodeHeartbeatPacket(rawBytes, imei)
			if err != nil {
				utils.VPrint("error decoding heartbeat data: %s", err)
				return
			}

			// Convert to JSON for further processing (for debug purposes only)
			jsonData, err := statusData.ToJSON()
			if err != nil {
				utils.VPrint("error converting heartbeat data to JSON: %s", err)
				return
			}

			// Cache the heartbeat data for this device
			cacheDeviceData(imei, statusData.VoltageValue, int(statusData.GSMSignalStrengthByte))

			// Print detailed debug info
			utils.VPrint("Heartbeat packet from %s (IMEI: %s)", clientAddr, imei)
			utils.VPrint("Terminal info: %s", statusData.TerminalInformationString)
			utils.VPrint("Voltage level: %s (%.1f volts)", statusData.VoltageLevelString, statusData.VoltageValue)
			utils.VPrint("GSM signal: %s", statusData.GSMSignalStrengthString)
			utils.VPrint("Complete status data: %s", jsonData)
			utils.VPrint("Device data cached for IMEI: %s", imei)
			utils.VPrint("NOTE: Heartbeat data cached but not published to jonoprotocol (waiting for location data)")

			// REMOVED: Do not publish heartbeat data directly to jonoprotocol
			// Only cache it for enhancing location packets

			// Respond to the heartbeat packet
			response := hex.EncodeToString(usecases.BuildHeartbeatResponse())
			tracker_data_json := TrackerData{
				Payload:    response,
				RemoteAddr: clientAddr,
			}
			response_json, err := json.Marshal(tracker_data_json)
			if err != nil {
				utils.VPrint("Error creating JSON:%v", err)
				return
			}
			client.Publish("tracker/send", 0, false, response_json)

		case usecases.IsStringInformationPacket(rawBytes):
			// Retrieve IMEI from the map
			imeiValue, ok := imeiStore.Load(clientAddr)
			if !ok {
				utils.VPrint("Error: IMEI not found for client %s", clientAddr)
				return
			}
			imei, ok := imeiValue.(string)
			if !ok {
				utils.VPrint("Error: Invalid IMEI type for client %s", clientAddr)
				return
			}

			utils.VPrint("Processing String Information packet (0x15) from IMEI: %s", imei)

			// Decode the string information packet
			data, err := usecases.DecodeStringInformationPacket(rawBytes, imei)
			if err != nil {
				utils.VPrint("Error decoding string information data: %s", err)
				return
			}

			// Log the extracted location information
			utils.VPrint("Extracted location - Lat: %f, Lon: %f, DateTime: %s",
				data.Latitude, data.Longitude, data.DateTime)

			// Enhance with any cached device information
			enhanceLocationDataWithCache(imei, data)

			// Also explicitly cache this location data for future use
			cacheLocationData(imei, data)

			// Convert to JSON for Jono normalization
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				utils.VPrint("Error marshalling string information data: %s", err)
				return
			}

			// Transform to Jono format
			jonoNormalize, err := jono.Initialize(string(jsonBytes))
			if err != nil {
				utils.VPrint("Error transforming string information to jono format: %s", err)
				return
			}

			// Verify and ensure IMEI appears in final output
			var jonoJson map[string]interface{}
			if err := json.Unmarshal([]byte(jonoNormalize), &jonoJson); err == nil {
				if _, exists := jonoJson["IMEI"]; !exists {
					// Force include the IMEI in the final output if missing
					jonoJson["IMEI"] = imei
					if updatedData, err := json.Marshal(jonoJson); err == nil {
						jonoNormalize = string(updatedData)
						utils.VPrint("Added missing IMEI to final output: %s", imei)
					}
				}
			}

			// Publish to jonoprotocol topic
			if err := m.Publish("tracker/jonoprotocol", jonoNormalize); err != nil {
				utils.VPrint("Error publishing to jonoprotocol: %v", err)
			}

			// Publish IMEI assignment
			tracker_data_json := TrackerAssign{
				Imei:       imei,
				Protocol:   "pino",
				RemoteAddr: clientAddr,
			}
			assignImeiJson, err := json.Marshal(tracker_data_json)
			if err != nil {
				fmt.Println("Error marshaling assign-imei data:", err)
				return
			}
			if err := m.Publish("tracker/assign-imei2remoteaddr", assignImeiJson); err != nil {
				utils.VPrint("Error publishing to assign-imei2remoteaddr: %v", err)
			}
			utils.VPrint("Published IMEI assignment for string information packet")

		default:
			utils.VPrint("packet unknown")
		}
	}
}

// Store device data in cache for later use - enhanced to store location data
func cacheDeviceData(imei string, voltage float64, gsmSignal int) {
	// Check if we have existing data to preserve location info
	var locationData *models.LocationPacketModel = nil

	// If there's already cached data, preserve the location info
	if existingData, found := deviceDataCache.Load(imei); found {
		if deviceData, ok := existingData.(DeviceData); ok {
			locationData = deviceData.LocationData
		}
	}

	deviceData := DeviceData{
		VoltageValue:      voltage,
		GSMSignalStrength: gsmSignal,
		LastUpdated:       time.Now(),
		LocationData:      locationData, // Preserve existing location data
	}
	deviceDataCache.Store(imei, deviceData)
}

// Update the cacheLocationData function to be more thorough:
func cacheLocationData(imei string, data *models.LocationPacketModel) {
	// First get existing device data if available
	var deviceData DeviceData

	if existingData, found := deviceDataCache.Load(imei); found {
		if existing, ok := existingData.(DeviceData); ok {
			deviceData = existing
		}
	}

	// Update with the new location data
	deviceData.LocationData = data
	deviceData.LastUpdated = time.Now()

	// If we have voltage and GSM signal info in this location packet, cache those too
	if data.VoltageValue > 0 {
		deviceData.VoltageValue = data.VoltageValue
	}
	if data.GSMSignalStrength > 0 {
		deviceData.GSMSignalStrength = data.GSMSignalStrength
	}

	// Store back to the cache
	deviceDataCache.Store(imei, deviceData)

	utils.VPrint("Cached location data for IMEI %s: lat=%f, lon=%f, voltage=%f, GSM=%d",
		imei, data.Latitude, data.Longitude, deviceData.VoltageValue, deviceData.GSMSignalStrength)
}

// Enhance location data with cached heartbeat information
func enhanceLocationDataWithCache(imei string, locationData *models.LocationPacketModel) {
	cachedData, found := deviceDataCache.Load(imei)
	if !found {
		// No cached data available
		return
	}

	deviceData, ok := cachedData.(DeviceData)
	if !ok {
		// Invalid cached data type
		return
	}

	// Use cached voltage for both BatteryLevel and VoltageValue
	// Map the voltage value back to battery level (0-6 scale) for compatibility
	batteryLevel := mapVoltageToBatteryLevel(deviceData.VoltageValue)
	locationData.BatteryLevel = batteryLevel * 100
	locationData.VoltageValue = deviceData.VoltageValue

	// Use cached GSM signal strength if location data doesn't have it
	if locationData.GSMSignalStrength == 0 {
		locationData.GSMSignalStrength = deviceData.GSMSignalStrength
	}

	// Store this location data for future alarms
	cacheLocationData(imei, locationData)
}

// New function that safely enhances alarm data with cached location
// without overriding alarm-specific data
// Map voltage value back to battery level for compatibility
func mapVoltageToBatteryLevel(voltage float64) int {
	switch {
	case voltage <= 0.1:
		return 0 // No Power
	case voltage <= 3.5:
		return 1 // Extremely Low
	case voltage <= 6.5:
		return 2 // Very Low
	case voltage <= 9:
		return 3 // Low
	case voltage <= 9:
		return 4 // Medium
	case voltage <= 9:
		return 5 // High
	default:
		return 6 // Very High
	}
}

// Add a helper function to compare output formats
func compareProtocolOutput(jonoNormalize string, packetType string) {
	// Extract the basic structure of the Jono protocol output
	var parsedOutput map[string]interface{}
	if err := json.Unmarshal([]byte(jonoNormalize), &parsedOutput); err == nil {
		// Check for critical fields that should be present
		utils.VPrint("%s packet Jono output structure:", packetType)
		utils.VPrint("- Has IMEI: %v", parsedOutput["IMEI"] != nil)
		// Don't print the message field contents, just presence
		utils.VPrint("- Has Message field: %v", parsedOutput["Message"] != nil)
		utils.VPrint("- DataPackets: %v", parsedOutput["DataPackets"])

		// Check if ListPackets exists and extract EventCode
		if listPackets, ok := parsedOutput["ListPackets"].(map[string]interface{}); ok {
			utils.VPrint("- ListPackets count: %d", len(listPackets))

			// Look at the first packet in ListPackets
			for key, packet := range listPackets {
				if packetObj, ok := packet.(map[string]interface{}); ok {
					if eventCode, ok := packetObj["EventCode"].(map[string]interface{}); ok {
						utils.VPrint("- Packet %s EventCode: Code=%v, Name=%v",
							key, eventCode["Code"], eventCode["Name"])
					} else {
						utils.VPrint("- Packet %s is missing EventCode structure", key)
					}
				}
				// Just check the first packet
				break
			}
		} else {
			utils.VPrint("- Missing ListPackets structure")
		}
	} else {
		utils.VPrint("Error parsing Jono output: %v", err)
	}
}

func main() {
	utils.VPrint("Starting Pino Tracker Protocol")
	// Get MQTT broker host from environment
	mqttBrokerHost := os.Getenv("MQTT_BROKER_HOST")
	if mqttBrokerHost == "" {
		log.Fatal("MQTT_BROKER_HOST environment variable not set")
	}

	// Create and configure MQTT client
	mqttClient, err := NewMQTTClient(mqttBrokerHost, "go_mqtt_client")
	if err != nil {
		log.Fatal("Failed to create MQTT client:", err)
	}

	// Connect to the broker
	if err := mqttClient.Connect(); err != nil {
		log.Fatal("Failed to connect to MQTT broker:", err)
	}

	// Subscribe to the raw topic
	if err := mqttClient.Subscribe("tracker/from-tcp", 1); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	// Keep the application running
	select {}
}
