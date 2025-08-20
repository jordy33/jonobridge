package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"suntechprotocol/features/jono"
	"suntechprotocol/features/suntech_protocol"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	verbose = flag.Bool("v", false, "Enable verbose logging") // Verbose flag
)

// Helper function to print verbose logs if enabled
func vPrint(format string, v ...interface{}) {
	if *verbose {
		log.Printf(format, v...)
	}
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MQTTClient wraps the MQTT client with additional functionality
type MQTTClient struct {
	client    mqtt.Client
	brokerURL string
	clientID  string
	verbose   bool
}

// NewMQTTClient creates a new MQTT client with the given configuration
func NewMQTTClient(brokerHost string, clientID string, verbose bool) (*MQTTClient, error) {
	if brokerHost == "" {
		return nil, fmt.Errorf("broker host cannot be empty")
	}

	brokerURL := fmt.Sprintf("tcp://%s:1883", brokerHost)
	return &MQTTClient{
		brokerURL: brokerURL,
		clientID:  clientID,
		verbose:   verbose,
	}, nil
}

// Connect establishes a connection to the MQTT broker
func (m *MQTTClient) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.brokerURL)

	// Generate a unique client ID
	subscribe_topic := "suntech" // You might want to make this configurable
	clientID := fmt.Sprintf("suntechprotocol_%s_%s_%d",
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

	// Try to connect with retries
	for {
		if token := m.client.Connect(); token.Wait() && token.Error() != nil {
			if m.verbose {
				vPrint("Error connecting to MQTT broker at %s: %v. Retrying in 5 seconds...", m.brokerURL, token.Error())
			}
			time.Sleep(5 * time.Second)
			continue
		}
		if m.verbose {
			vPrint("Successfully connected to MQTT broker!")
		}
		break
	}

	return nil
}

// Subscribe subscribes to the specified topic
func (m *MQTTClient) Subscribe(topic string, qos byte) error {
	if token := m.client.Subscribe(topic, qos, nil); token.Wait() && token.Error() != nil {
		return fmt.Errorf("error subscribing to topic %s: %v", topic, token.Error())
	}
	if m.verbose {
		vPrint("Subscribed to topic: %s", topic)
	}
	return nil
}

// Publish publishes a message to the specified topic
func (m *MQTTClient) Publish(topic string, payload interface{}) error {
	token := m.client.Publish(topic, 0, false, payload)
	token.Wait()
	if token.Error() != nil {
		if m.verbose {
			vPrint("Failed to publish to MQTT topic %s: %v", topic, token.Error())
		}
		return token.Error()
	}
	if m.verbose {
		vPrint("Successfully published to MQTT topic: %s", topic)
	}
	return nil
}

// messageHandler handles incoming MQTT messages
func (m *MQTTClient) messageHandler(client mqtt.Client, msg mqtt.Message) {
	if msg.Topic() == "tracker/from-udp" {
		if m.verbose {
			tracker_bytes := []byte(msg.Payload())
			vPrint("Received message on topic :\n%v", hex.Dump(tracker_bytes[:min(32, len(tracker_bytes))]))
		}
		trackerPayload := string(msg.Payload())

		var trackerData string
		// Try to decode as hex, if it fails, use the original message
		bytes, err := hex.DecodeString(trackerPayload)
		if err != nil {
			trackerData = trackerPayload
			if m.verbose {
				vPrint("error: %s", err)
			}
		} else {
			trackerData = string(bytes)
		}

		//fields := strings.Split(trackerData, ",")
		//if len(fields) > 2 {
		dataSuntech, err := suntech_protocol.Initialize(trackerData)
		if err != nil {
			fmt.Println(err)
			return
		}

		jonoNormalize, err := jono.Initialize(dataSuntech)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Publish to jonoprotocol topic
		if err := m.Publish("tracker/jonoprotocol", jonoNormalize); err != nil {
			fmt.Println("Error publishing to jonoprotocol:", err)
			return
		}
		if m.verbose {
			// Create a compact version of the JSON for logging
			var jsonObj map[string]interface{}
			if err := json.Unmarshal([]byte(jonoNormalize), &jsonObj); err == nil {
				// Re-marshal without indentation
				if compactJSON, err := json.Marshal(jsonObj); err == nil {
					vPrint("Jono Protocol: %s", string(compactJSON))
				} else {
					vPrint("Jono Protocol: %s", jonoNormalize) // Fallback to pretty JSON if compact fails
				}
			} else {
				vPrint("Jono Protocol: %s", jonoNormalize) // Fallback to pretty JSON if unmarshaling fails
			}
		}
		//}
		return
	}

	// Handle JSON messages
	var json_data TrackerData
	if err := json.Unmarshal(msg.Payload(), &json_data); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}
	payload := json_data.Payload
	remote_addr := json_data.RemoteAddr
	//fmt.Printf("Received message on topic %s from %s: %s\n", msg.Topic(), remote_addr, payload)
	trackerPayload := payload

	var trackerData string
	// Try to decode as hex, if it fails, use the original message
	bytes, err := hex.DecodeString(trackerPayload)
	if err != nil {
		trackerData = trackerPayload
		if m.verbose {
			vPrint("error: %s", err)
		}
	} else {
		trackerData = string(bytes)
	}

	if m.verbose {
		vPrint("Received message on topic %s: %s", msg.Topic(), msg.Payload())
		tracker_bytes := []byte(trackerData)
		vPrint("Received message on topic :\n%v", hex.Dump(tracker_bytes[:min(32, len(tracker_bytes))]))
	}
	//fields := strings.Split(trackerData, ",")
	//if len(fields) > 2 {
	dataSuntech, err := suntech_protocol.Initialize(trackerData)
	if err != nil {
		fmt.Println(err)
		return
	}
	//imei := fields[1]
	imei := ""
	jonoNormalize, err := jono.Initialize(dataSuntech)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Publish to jonoprotocol topic
	if err := m.Publish("tracker/jonoprotocol", jonoNormalize); err != nil {
		fmt.Println("Error publishing to jonoprotocol:", err)
		return
	}
	if m.verbose {
		// Create a compact version of the JSON for logging
		var jsonObj map[string]interface{}
		if err := json.Unmarshal([]byte(jonoNormalize), &jsonObj); err == nil {
			// Re-marshal without indentation
			if compactJSON, err := json.Marshal(jsonObj); err == nil {
				vPrint("Jono Protocol: %s", string(compactJSON))
			} else {
				vPrint("Jono Protocol: %s", jonoNormalize) // Fallback to pretty JSON if compact fails
			}
		} else {
			vPrint("Jono Protocol: %s", jonoNormalize) // Fallback to pretty JSON if unmarshaling fails
		}
	}
	tracker_data_json := TrackerAssign{
		Imei:       imei,
		Protocol:   "suntech",
		RemoteAddr: remote_addr,
	}
	// Marshal the data to JSON before publishing
	assignImeiJson, err := json.Marshal(tracker_data_json)
	if err != nil {
		fmt.Println("Error marshaling assign-imei data:", err)
		return
	}
	// Publish to assign-imei2remoteaddr topic
	if err := m.Publish("tracker/assign-imei2remoteaddr", assignImeiJson); err != nil {
		fmt.Println("Error publishing to assign-imei2remoteaddr:", err)
	}
	if m.verbose {
		fmt.Printf("Publishing to tracker/assign-imei2remoteaddr: %s", assignImeiJson)
	}
	//}
}

func main() {
	// Parse command-line flags
	flag.Parse()

	// Get MQTT broker host from environment
	mqttBrokerHost := os.Getenv("MQTT_BROKER_HOST")
	if mqttBrokerHost == "" {
		log.Fatal("MQTT_BROKER_HOST environment variable not set")
	}

	// Create and configure MQTT client
	mqttClient, err := NewMQTTClient(mqttBrokerHost, "go_mqtt_client", *verbose)
	if err != nil {
		log.Fatal("Failed to create MQTT client:", err)
	}

	// Connect to the broker
	if err := mqttClient.Connect(); err != nil {
		log.Fatal("Failed to connect to MQTT broker:", err)
	}
	// Subscribe to the connection topic
	if err := mqttClient.Subscribe("tracker/from-tcp", 1); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}
	// Subscribe to the raw topic
	if err := mqttClient.Subscribe("tracker/from-udp", 1); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}
	// Keep the application running
	select {}
}
