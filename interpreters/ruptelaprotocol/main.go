package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"ruptelaprotocol/features/jono"
	"ruptelaprotocol/features/ruptela_protocol"
	"ruptelaprotocol/utils"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	verbose = flag.Bool("v", false, "Enable verbose logging") // Verbose flag
)

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
	subscribe_topic := "meitrack" // You might want to make this configurable
	clientID := fmt.Sprintf("ruptelaprotocol_%s_%s_%d",
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
				utils.VPrint("Error connecting to MQTT broker at %s: %v. Retrying in 5 seconds...", m.brokerURL, token.Error())
			}
			time.Sleep(5 * time.Second)
			continue
		}
		if m.verbose {
			utils.VPrint("Successfully connected to the MQTT broker")
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
		utils.VPrint("Subscribed to topic: %s", topic)
	}
	return nil
}

// Publish publishes a message to the specified topic
func (m *MQTTClient) Publish(topic string, payload interface{}) error {
	token := m.client.Publish(topic, 0, false, payload)
	token.Wait()
	if token.Error() != nil {
		if m.verbose {
			utils.VPrint("Failed to publish to MQTT topic %s: %v", topic, token.Error())
		}
		return token.Error()
	}
	if m.verbose {
		utils.VPrint("Successfully published to MQTT topic: %s", topic)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// messageHandler handles incoming MQTT messages
func (m *MQTTClient) messageHandler(client mqtt.Client, msg mqtt.Message) {

	if msg.Topic() == "tracker/from-udp" {
		if m.verbose {
			utils.VPrint("Received message on topic %s: %s", msg.Topic(), msg.Payload())
		}

		trackerPayload := string(msg.Payload())

		dataRuptela, err := ruptela_protocol.Initialize(trackerPayload)
		if err != nil {
			fmt.Println(err)
			return
		}

		jonoNormalize, err := jono.Initialize(dataRuptela)
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
			utils.VPrint("Jono Protocol: %s", jonoNormalize)
		}
		return
	}

	var json_data TrackerData
	if err := json.Unmarshal(msg.Payload(), &json_data); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}
	payload := json_data.Payload
	remote_addr := json_data.RemoteAddr

	fmt.Println("remote_addr:", remote_addr)

	//utils.VPrint("Data fetched:\n%v", )
	bytes, _ := hex.DecodeString(payload)
	n := len(bytes)

	if m.verbose {
		//utils.VPrint("Received message on topic %s:\n%s", msg.Topic(),hex.Dump(bytes[:min(32, n)]))
		utils.VPrint("Received message on topic %s:\n%s", msg.Topic(), hex.Dump(bytes[:min(32, n)]))
	}

	dataRuptela, err := ruptela_protocol.Initialize(payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	jonoNormalize, err := jono.Initialize(dataRuptela)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Parse the data to get just the fields we want
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jonoNormalize), &data); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Extract the packet_1 data
	listPackets := data["ListPackets"].(map[string]interface{})
	packet1 := listPackets["packet_1"].(map[string]interface{})
	extras := packet1["Extras"].(map[string]interface{})

	// Create the final JSON structure
	finalJSON := map[string]string{
		"Altitude":           fmt.Sprint(packet1["Altitude"]),
		"Datetime":           fmt.Sprint(packet1["Datetime"]),
		"Direction":          fmt.Sprint(extras["Direction"]),
		"EventCode":          fmt.Sprint(packet1["EventCode"]),
		"Hdop":               fmt.Sprint(extras["Hdop"]),
		"IMEI":               fmt.Sprint(data["IMEI"]),
		"Latitude":           fmt.Sprint(packet1["Latitude"]),
		"Longitude":          fmt.Sprint(packet1["Longitude"]),
		"NumberOfSatellites": fmt.Sprint(extras["NumberOfSatellites"]),
		"Speed":              fmt.Sprint(packet1["Speed"]),
	}

	// Convert to JSON and print
	jsonResult, err := json.Marshal(finalJSON)
	if err != nil {
		fmt.Println("Error creating JSON:", err)
		return
	}

	if m.verbose {
		utils.VPrint("Jono Protocol:\n%v", string(jsonResult))
	}
	// Publish to jonoprotocol topic
	if err := m.Publish("tracker/jonoprotocol", string(jsonResult)); err != nil {
		fmt.Println("Error publishing to jonoprotocol:", err)
		return
	}
	tracker_data_json := TrackerAssign{
		Imei:       data["IMEI"].(string),
		Protocol:   "ruptela",
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
}

func main() {
	// Parse command-line flags
	flag.Parse()

	// Set verbose flag in utils package
	utils.SetVerbose(verbose)
	utils.VPrint("Rupetela Protocol")
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

	// Subscribe to the raw topic
	if err := mqttClient.Subscribe("tracker/from-tcp", 1); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	// Subscribe to tracker/from-udp topic
	if err := mqttClient.Subscribe("tracker/from-udp", 1); err != nil {
		log.Fatal("Failed to subscribe to tracker/from-udp:", err)
	}

	// Keep the application running
	select {}
}
