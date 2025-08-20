package suntech_protocol

import (
	"encoding/json"
	"fmt"
	"suntechprotocol/features/suntech_protocol/models"
	"suntechprotocol/features/suntech_protocol/usecases"
	"time"
)

func Initialize(data string) (string, error) {
	model := models.IdentifyModel(data)
	if model != "" {
		fmt.Printf("Identified model: %s\n", model)
		if model == "ST300" {
			// Handle ST300 model
			dataParsed, err := usecases.ParseST300Fields(data)
			fmt.Printf("Parsed ST300 Data:\n")
			fmt.Printf("  Header: %s\n", dataParsed.Header)
			fmt.Printf("  Device ID: %s\n", dataParsed.IMEI)
			fmt.Printf("  Message Type: %s\n", dataParsed.MessageType)
			fmt.Printf("  Timestamp: %s\n", dataParsed.Timestamp.Format(time.RFC3339))
			fmt.Printf("  Latitude: %.6f\n", dataParsed.Latitude)
			fmt.Printf("  Longitude: %.6f\n", dataParsed.Longitude)
			fmt.Printf("  Speed: %.2f km/h\n", dataParsed.Speed)
			fmt.Printf("  Heading: %.2f degrees\n", dataParsed.Heading)
			fmt.Printf("  Satellites: %d\n", dataParsed.Satellites)
			fmt.Printf("  Ignition: %v\n", dataParsed.Ignition)
			fmt.Printf("  Battery Level: %.2f V\n", dataParsed.BatteryLevel)
			if err != nil {
				return "", fmt.Errorf("error: st300 - %v - data %s", err, data)
			}
			dataJSON, err := json.Marshal(dataParsed)
			if err != nil {
				return "", fmt.Errorf("error marshaling ST300 data: %v", err)
			}
			return string(dataJSON), nil
		}
		if model == "ST4300" {
			// Handle ST4300 model
			dataParsed, err := usecases.ParseST4300Fields(data)
			fmt.Printf("Parsed ST4300 Data:\n")
			fmt.Printf("  Header: %s\n", dataParsed.Header)
			fmt.Printf("  Device ID: %s\n", dataParsed.IMEI)
			fmt.Printf("  Message Type: %s\n", dataParsed.MessageType)
			fmt.Printf("  Timestamp: %s\n", dataParsed.Timestamp.Format(time.RFC3339))
			fmt.Printf("  Latitude: %.6f\n", dataParsed.Latitude)
			fmt.Printf("  Longitude: %.6f\n", dataParsed.Longitude)
			fmt.Printf("  Speed: %.2f km/h\n", dataParsed.Speed)
			fmt.Printf("  Heading: %.2f degrees\n", dataParsed.Heading)
			fmt.Printf("  Satellites: %d\n", dataParsed.Satellites)
			fmt.Printf("  HDOP: %.2f\n", dataParsed.HDOP)
			fmt.Printf("  Altitude: %.2f m\n", dataParsed.Altitude)
			fmt.Printf("  Ignition: %v\n", dataParsed.Ignition)
			fmt.Printf("  Battery Level: %.2f V\n", dataParsed.BatteryLevel)
			fmt.Printf("  Odometer: %.2f km\n", dataParsed.Odometer)
			fmt.Printf("  Input Status: %d\n", dataParsed.InputStatus)
			fmt.Printf("  Output Status: %d\n", dataParsed.OutputStatus)
			if err != nil {
				return "", fmt.Errorf("error: st4300 - %v - data %s", err, data)
			}
			dataJSON, err := json.Marshal(dataParsed)
			if err != nil {
				return "", fmt.Errorf("error marshaling ST4300 data: %v", err)
			}
			return string(dataJSON), nil
		}
	} else {
		fmt.Println("Could not identify model from data")
	}
	return "", fmt.Errorf("no valid model identified from data: %s", data)
}
