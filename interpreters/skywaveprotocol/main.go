package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"skywaveprotocol/features/jono"
	"skywaveprotocol/features/skywave_protocol"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	verbose = flag.Bool("v", false, "Enable verbose logging")
)

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

type MQTTClient struct {
	client    mqtt.Client
	brokerURL string
	clientID  string
	verbose   bool
}

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

func (m *MQTTClient) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.brokerURL)

	subscribe_topic := "skywave"
	clientID := fmt.Sprintf("skywaveprotocol_%s_%s_%d",
		subscribe_topic,
		os.Getenv("HOSTNAME"),
		time.Now().UnixNano()%100000)
	opts.SetClientID(clientID)

	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetOrderMatters(true)
	opts.SetResumeSubs(true)
	opts.SetDefaultPublishHandler(m.messageHandler)

	m.client = mqtt.NewClient(opts)

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

func (m *MQTTClient) Subscribe(topic string, qos byte) error {
	if token := m.client.Subscribe(topic, qos, nil); token.Wait() && token.Error() != nil {
		return fmt.Errorf("error subscribing to topic %s: %v", topic, token.Error())
	}
	if m.verbose {
		vPrint("Subscribed to topic: %s", topic)
	}
	return nil
}

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

func (m *MQTTClient) messageHandler(client mqtt.Client, msg mqtt.Message) {
	if msg.Topic() == "tracker/from-udp" {
		if m.verbose {
			tracker_bytes := []byte(msg.Payload())
			vPrint("Received message on topic :\n%v", hex.Dump(tracker_bytes[:min(32, len(tracker_bytes))]))
		}
		trackerPayload := string(msg.Payload())

		var trackerData string
		bytes, err := hex.DecodeString(trackerPayload)
		if err != nil {
			trackerData = trackerPayload
			if m.verbose {
				vPrint("error: %s", err)
			}
		} else {
			trackerData = string(bytes)
		}

		dataskywave, err := skywave_protocol.Initialize(trackerData)
		if err != nil {
			fmt.Println(err)
			return
		}

		jonoNormalize, err := jono.Initialize(dataskywave)
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := m.Publish("tracker/jonoprotocol", jonoNormalize); err != nil {
			fmt.Println("Error publishing to jonoprotocol:", err)
			return
		}
		if m.verbose {
			var jsonObj map[string]interface{}
			if err := json.Unmarshal([]byte(jonoNormalize), &jsonObj); err == nil {
				if compactJSON, err := json.Marshal(jsonObj); err == nil {
					vPrint("Jono Protocol: %s", string(compactJSON))
				} else {
					vPrint("Jono Protocol: %s", jonoNormalize)
				}
			} else {
				vPrint("Jono Protocol: %s", jonoNormalize)
			}
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

	trackerPayload := payload

	var trackerData string
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

	dataskywave, err := skywave_protocol.Initialize(trackerData)
	if err != nil {
		fmt.Println(err)
		return
	}

	imei := ""
	jonoNormalize, err := jono.Initialize(dataskywave)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := m.Publish("tracker/jonoprotocol", jonoNormalize); err != nil {
		fmt.Println("Error publishing to jonoprotocol:", err)
		return
	}
	if m.verbose {
		var jsonObj map[string]interface{}
		if err := json.Unmarshal([]byte(jonoNormalize), &jsonObj); err == nil {
			if compactJSON, err := json.Marshal(jsonObj); err == nil {
				vPrint("Jono Protocol: %s", string(compactJSON))
			} else {
				vPrint("Jono Protocol: %s", jonoNormalize)
			}
		} else {
			vPrint("Jono Protocol: %s", jonoNormalize)
		}
	}
	tracker_data_json := TrackerAssign{
		Imei:       imei,
		Protocol:   "skywave",
		RemoteAddr: remote_addr,
	}

	assignImeiJson, err := json.Marshal(tracker_data_json)
	if err != nil {
		fmt.Println("Error marshaling assign-imei data:", err)
		return
	}

	if err := m.Publish("tracker/assign-imei2remoteaddr", assignImeiJson); err != nil {
		fmt.Println("Error publishing to assign-imei2remoteaddr:", err)
	}
	if m.verbose {
		fmt.Printf("Publishing to tracker/assign-imei2remoteaddr: %s", assignImeiJson)
	}
}

func main() {
	flag.Parse()

	mqttBrokerHost := os.Getenv("MQTT_BROKER_HOST")
	if mqttBrokerHost == "" {
		log.Fatal("MQTT_BROKER_HOST environment variable not set")
	}

	mqttClient, err := NewMQTTClient(mqttBrokerHost, "go_mqtt_client", *verbose)
	if err != nil {
		log.Fatal("Failed to create MQTT client:", err)
	}

	if err := mqttClient.Connect(); err != nil {
		log.Fatal("Failed to connect to MQTT broker:", err)
	}

	if err := mqttClient.Subscribe("tracker/from-tcp", 1); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	if err := mqttClient.Subscribe("tracker/from-udp", 1); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	select {}
}
