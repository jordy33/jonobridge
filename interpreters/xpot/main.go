package main

import (
	"database/sql"
	"encoding/hex"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"xpot/utils"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/go-sql-driver/mysql"
)

// Codigo de la vieja escuela
var server1Address *string = flag.String("g", getEnvWithDefault("XPOT_FORWARD_HOST", "server1.gpscontrol.com.mx:8500"), "Gate address")

func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getMySQLDSN() string {
	host := getEnvWithDefault("MYSQL_HOST", "127.0.0.1")
	port := getEnvWithDefault("MYSQL_PORT", "3306")
	user := getEnvWithDefault("MYSQL_USER", "gpscontrol")
	pass := getEnvWithDefault("MYSQL_PASS", "qazwsxedc")
	dbname := getEnvWithDefault("MYSQL_DB", "bridge")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, dbname)
}

func crc(source string) string {
	sum := 0
	for i := 0; i < len(source); i++ {
		newString := source[i : i+1]
		nv := []rune(newString)[0]
		sum = sum + int(nv)
	}

	module := sum % 256

	hv := fmt.Sprintf("%x", module)

	res1 := strings.ToUpper(hv)
	return res1
}

// FIN del codigo de la vieja escuela

// Definición de las estructuras para deserializar el XML
type Response struct {
	XMLName             xml.Name            `xml:"response"`
	FeedMessageResponse FeedMessageResponse `xml:"feedMessageResponse"`
}

type FeedMessageResponse struct {
	Count         int       `xml:"count"`
	Feed          Feed      `xml:"feed"`
	TotalCount    int       `xml:"totalCount"`
	ActivityCount int       `xml:"activityCount"`
	Messages      []Message `xml:"messages>message"`
}

type Feed struct {
	ID                   string `xml:"id"`
	Name                 string `xml:"name"`
	Description          string `xml:"description"`
	Status               string `xml:"status"`
	Usage                int    `xml:"usage"`
	DaysRange            int    `xml:"daysRange"`
	DetailedMessageShown bool   `xml:"detailedMessageShown"`
	Type                 string `xml:"type"`
}

type Message struct {
	ID             int     `xml:"id"`
	MessengerID    string  `xml:"messengerId"`
	MessengerName  string  `xml:"messengerName"`
	UnixTime       int64   `xml:"unixTime"`
	MessageType    string  `xml:"messageType"`
	Latitude       float64 `xml:"latitude"`
	Longitude      float64 `xml:"longitude"`
	ModelID        string  `xml:"modelId"`
	ShowCustomMsg  string  `xml:"showCustomMsg"`
	DateTime       string  `xml:"dateTime"`
	BatteryState   string  `xml:"batteryState"`
	Hidden         int     `xml:"hidden"`
	Altitude       int     `xml:"altitude"`
	MessageContent string  `xml:"messageContent,omitempty"`
}

type Device struct {
	ID     int
	Imei   string
	Plates string
	VIN    string
}

func processSpotXData(db *sql.DB) error {
	// Set up MQTT client options
	opts := mqtt.NewClientOptions()
	mqttBrokerHost := os.Getenv("MQTT_BROKER_HOST")
	if mqttBrokerHost == "" {
		log.Fatal("MQTT_BROKER_HOST environment variable not set")
	}
	brokerURL := fmt.Sprintf("tcp://%s:1883", mqttBrokerHost)
	opts.AddBroker(brokerURL)
	subscribe_topic := "http/get"
	clientID := fmt.Sprintf("xpot_%s_%s_%d",
		subscribe_topic,
		os.Getenv("HOSTNAME"),
		time.Now().UnixNano()%100000)
	opts.SetClientID(clientID)

	// Configure MQTT client settings
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetOrderMatters(true)
	opts.SetResumeSubs(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})
	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		log.Printf("Attempting to reconnect to MQTT broker")
	})

	// Create and connect MQTT client
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("error connecting to MQTT broker: %v", token.Error())
	}
	log.Printf("Connected to MQTT broker at %s", brokerURL)

	// Subscribe to the topic
	if token := mqttClient.Subscribe(subscribe_topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		trackerPayload := string(msg.Payload())
		var trackerData string

		// Try to decode as hex, if it fails, use the original message
		bytes, err := hex.DecodeString(trackerPayload)
		if err != nil {
			trackerData = trackerPayload
			utils.VPrint("Hex decode error: %s", err)
		} else {
			trackerData = string(bytes)
		}

		var response Response
		if err := xml.Unmarshal([]byte(trackerData), &response); err != nil {
			utils.VPrint("Error deserializing XML: %v", err)
			return
		}

		utils.VPrint("Processing %d messages from SpotX", len(response.FeedMessageResponse.Messages))
		processedIDs := make(map[string]bool)

		for _, message := range response.FeedMessageResponse.Messages {
			// Skip if we've already processed this messenger in this batch
			if processedIDs[message.MessengerID] {
				continue
			}
			processedIDs[message.MessengerID] = true

			utils.VPrint("Processing message for Messenger ID: %s", message.MessengerID)

			var exists bool
			err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM spotx_message WHERE id = ? and messengerId = ?)",
				message.ID, message.MessengerID).Scan(&exists)
			if err != nil {
				log.Printf("Error checking message existence: %v", err)
				continue
			}

			if !exists {
				if err := processMessage(db, message); err != nil {
					log.Printf("Error processing message: %v", err)
					continue
				}
			} else {
				utils.VPrint("Message already exists in database, skipping")
			}
		}
	}); token.Wait() && token.Error() != nil {
		return fmt.Errorf("error subscribing to topic: %v", token.Error())
	}

	log.Printf("Subscribed to topic: %s", subscribe_topic)

	// Keep the connection alive
	select {}
}

func processMessage(db *sql.DB, message Message) error {
	utils.VPrint("New message found, processing...")
	// --------- CODIGO PARA ACTUALIZAR EL HORARIO EN LA VISTA DEL BRIDGE
	input := message.DateTime
	parsedTime, err := time.Parse("2006-01-02T15:04:05-0700", input)
	if err != nil {
		return fmt.Errorf("error parsing date: %v", err)
	}

	// Convertir a UTC
	utcTime := parsedTime.UTC()
	output := utcTime.Format("2006/01/02 15:04:05")
	utils.VPrint("Message Values:")
	utils.VPrint("  MessengerID: %s", message.MessengerID)
	utils.VPrint("  DateTime: %s", message.DateTime)
	utils.VPrint("  MessageType: %s", message.MessageType)
	utils.VPrint("  Latitude: %f", message.Latitude)
	utils.VPrint("  Longitude: %f", message.Longitude)
	utils.VPrint("  Altitude: %d", message.Altitude)

	// Send to Elasticsearch
	client_id := os.Getenv("CLIENT_ID")

	logData := utils.ElasticLogData{
		Client:      client_id,
		MessengerId: message.MessengerID,
		DateTime:    message.DateTime,
		Type:        message.MessageType,
		Lat:         message.Latitude,
		Lon:         message.Longitude,
		Alt:         message.Altitude,
	}

	if err := utils.SendToElastic(logData); err != nil {
		utils.VPrint("Error sending to elastic: %v", err)
		// Don't return the error as we don't want to fail the main operation
	}

	payloaddata := "*" + output + " IMEI:" + message.MessengerID +
		" fecha:" + message.DateTime +
		" EC:" + message.MessageType +
		" lat:" + strconv.FormatFloat(message.Latitude, 'f', 6, 64) +
		" lon:" + strconv.FormatFloat(message.Longitude, 'f', 6, 64) +
		" alt:" + strconv.Itoa(message.Altitude) +
		" vel:0 az:0"
	cmd := "UPDATE devices SET log='" + payloaddata + "' WHERE protocol=100 and password='" + message.MessengerID + "' and ff0= " + strconv.Itoa(message.ID)
	utils.VPrint("Executing update query: %s", cmd)
	_, err = db.Exec(cmd)
	if err != nil {
		return fmt.Errorf("error updating device log: %v", err)
	}

	// --------- ENVIO AL SERVER 1
	utils.VPrint("Preparing to send data to Server1")
	input2 := message.DateTime
	originalTime, err := time.Parse("2006-01-02T15:04:05-0700", input2)
	if err != nil {
		return fmt.Errorf("error parsing date for Server1: %v", err)
	}
	utcTime = originalTime.In(time.UTC)
	ano := utcTime.Year() % 100
	mes := utcTime.Month()
	dia := utcTime.Day()
	hora := utcTime.Hour()
	minuto := utcTime.Minute()
	segundo := utcTime.Second()

	dateserver := fmt.Sprintf("%02d%02d%02d%02d%02d%02d", ano, mes, dia, hora, minuto, segundo)
	idlimpio := strings.Replace(message.MessengerID, "-", "", -1)
	imei := "2024000" + idlimpio
	trama := "," + imei + ",AAA,35," + strconv.FormatFloat(message.Latitude, 'f', 6, 64) + "," + strconv.FormatFloat(message.Longitude, 'f', 6, 64) + "," + dateserver + ",A,1,14,0,100,1.0,2264,0,0,334|3|2349|A37D,0000,0002|0000|0000|0A27|0000,00000001,*F0"
	totalchar2 := len(trama) + 5
	header2 := "$$A" + strconv.Itoa(totalchar2)
	preoutput2 := header2 + trama
	payload2 := preoutput2 + crc(preoutput2) + "\r\n"
	data2 := []byte(payload2)

	utils.VPrint("Connecting to Server1: %s", *server1Address)
	nDDr2, err := net.ResolveTCPAddr("tcp", *server1Address)
	if err != nil {
		return fmt.Errorf("error resolving Server1 address: %v", err)
	}

	server1TimeConn, err := net.DialTCP("tcp", nil, nDDr2)
	if err != nil {
		return fmt.Errorf("error connecting to Server1: %v", err)
	}
	defer server1TimeConn.Close()

	utils.VPrint("Sending data to Server1: %s", payload2)
	_, err = server1TimeConn.Write(data2)
	if err != nil {
		return fmt.Errorf("error sending data to Server1: %v", err)
	}

	// ---------  SAVE THE MESSAGE TO DATABASE
	cmd = "INSERT INTO spotx_message (id,messengerId,messengerName,unixTime,messageType,latitude,longitude,modelId,showCustomMsg,dateTime,batteryState,hidden,altitude) " +
		"value (" + strconv.Itoa(message.ID) + ",'" + message.MessengerID + "','" + message.MessengerName + "','" + strconv.FormatInt(message.UnixTime, 10) + "','" +
		message.MessageType + "','" + strconv.FormatFloat(message.Latitude, 'f', 6, 64) + "','" + strconv.FormatFloat(message.Longitude, 'f', 6, 64) +
		"','" + message.ModelID + "','" + message.ShowCustomMsg + "','" + message.DateTime + "','" + message.BatteryState + "'," + strconv.Itoa(message.Hidden) + "," + strconv.Itoa(message.Altitude) + ");"
	utils.VPrint("Saving message to database: %s", cmd)
	_, err = db.Exec(cmd)
	if err != nil {
		return fmt.Errorf("error saving message to database: %v", err)
	}

	return nil
}

func main() {
	// Add debug flag
	debugFlag := flag.Bool("v", false, "Enable verbose output")
	flag.Parse()

	utils.SetVerbose(*debugFlag)

	// Paso 1. Ir por todos los equipos registrados en la base de datos:
	var db *sql.DB
	var err error
	utils.VPrint("Connecting to database...")
	dsn := getMySQLDSN()
	utils.VPrint("Using MySQL DSN: %s", dsn)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Error al cargar el driver mysql: %v", err)
		return
	}
	defer db.Close()

	// Create devices table if it doesn't exist
	utils.VPrint("Creating devices table if not exists...")
	createDevicesSQL := `CREATE TABLE IF NOT EXISTS devices (
		imei varchar(255) DEFAULT NULL,
		plates varchar(255) DEFAULT NULL,
		vin varchar(255) DEFAULT NULL,
		protocol int(11) DEFAULT NULL,
		password varchar(255) DEFAULT NULL,
		log text DEFAULT NULL,
		ff0 int(11) DEFAULT NULL,
		INDEX idx_protocol (protocol),
		INDEX idx_password (password),
		INDEX idx_ff0 (ff0)
	) ENGINE=InnoDB DEFAULT CHARSET=latin1`

	_, err = db.Exec(createDevicesSQL)
	if err != nil {
		log.Printf("Error creating devices table: %v", err)
		return
	}

	// Create spotx_message table if it doesn't exist
	utils.VPrint("Creating spotx_message table if not exists...")
	createTableSQL := `CREATE TABLE IF NOT EXISTS spotx_message (
		id varchar(255) DEFAULT NULL,
		messengerId varchar(255) DEFAULT NULL,
		messengerName varchar(255) DEFAULT NULL,
		unixTime varchar(255) DEFAULT NULL,
		messageType varchar(255) DEFAULT NULL,
		latitude varchar(255) DEFAULT NULL,
		longitude varchar(255) DEFAULT NULL,
		modelId varchar(255) DEFAULT NULL,
		showCustomMsg varchar(255) DEFAULT NULL,
		dateTime varchar(255) DEFAULT NULL,
		batteryState varchar(255) DEFAULT NULL,
		hidden int(11) DEFAULT NULL,
		altitude int(11) DEFAULT NULL,
		INDEX idx_id_messenger (id, messengerId)
	) ENGINE=InnoDB DEFAULT CHARSET=latin1`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Printf("Error creating spotx_message table: %v", err)
		return
	}

	log.Println("Starting MQTT listener for SpotX data...")
	if err := processSpotXData(db); err != nil {
		log.Fatal(err)
	}
}
