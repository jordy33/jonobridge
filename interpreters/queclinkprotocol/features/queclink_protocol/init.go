package queclink_protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"queclinkprotocol/features/queclink_protocol/models"
	"queclinkprotocol/features/queclink_protocol/usecases"

	"github.com/MaddSystems/jonobridge/common/utils"
)

func Initialize(data string) (string, error) {
	model := models.IdentifyModel(data)
	if model != "" {
		utils.VPrint("Identified model: %s\n", model)

		switch model {
		case "300":
			// Handle Queclink 300 model
			dataParsed, err := usecases.ParseQueclink300Fields(data)
			if err != nil {
				return "", fmt.Errorf("error: model 300 - %v - data %s", err, data)
			}

			utils.VPrint("Parsed Queclink 300 Data:\n")
			utils.VPrint("  Device Type: %s\n", dataParsed.DeviceType)
			utils.VPrint("  Device Version: %s\n", dataParsed.DeviceVersion)
			utils.VPrint("  IMEI: %s\n", dataParsed.IMEI)
			utils.VPrint("  Message Type: %s\n", dataParsed.MessageType)
			utils.VPrint("  Event Code: %s\n", dataParsed.EventCode)
			utils.VPrint("  Timestamp: %s\n", dataParsed.Timestamp.Format(time.RFC3339))
			utils.VPrint("  Latitude: %.6f\n", dataParsed.Latitude)
			utils.VPrint("  Longitude: %.6f\n", dataParsed.Longitude)
			utils.VPrint("  Speed: %.2f km/h\n", dataParsed.Speed)
			utils.VPrint("  Heading: %.2f degrees\n", dataParsed.Heading)
			utils.VPrint("  Satellites: %d\n", dataParsed.Satellites)
			utils.VPrint("  Ignition: %v\n", dataParsed.Ignition)
			utils.VPrint("  Battery Level: %.2f V\n", dataParsed.BatteryLevel)
			utils.VPrint("  External Power: %.2f V\n", dataParsed.ExternalPower)

			dataJSON, err := json.Marshal(dataParsed)
			if err != nil {
				return "", fmt.Errorf("error marshaling Queclink 300 data: %v", err)
			}
			return string(dataJSON), nil

		case "320":
			// Handle Queclink 320 model
			dataParsed, err := usecases.ParseQueclink320Fields(data)
			if err != nil {
				return "", fmt.Errorf("error: model 320 - %v - data %s", err, data)
			}

			utils.VPrint("Parsed Queclink 320 Data:\n")
			utils.VPrint("  Device Type: %s\n", dataParsed.DeviceType)
			utils.VPrint("  Device Version: %s\n", dataParsed.DeviceVersion)
			utils.VPrint("  IMEI: %s\n", dataParsed.IMEI)
			utils.VPrint("  Message Type: %s\n", dataParsed.MessageType)
			utils.VPrint("  Event Code: %s\n", dataParsed.EventCode)
			utils.VPrint("  Timestamp: %s\n", dataParsed.Timestamp.Format(time.RFC3339))
			utils.VPrint("  Latitude: %.6f\n", dataParsed.Latitude)
			utils.VPrint("  Longitude: %.6f\n", dataParsed.Longitude)
			utils.VPrint("  Speed: %.2f km/h\n", dataParsed.Speed)
			utils.VPrint("  Heading: %.2f degrees\n", dataParsed.Heading)
			utils.VPrint("  Satellites: %d\n", dataParsed.Satellites)
			utils.VPrint("  HDOP: %.2f\n", dataParsed.HDOP)
			utils.VPrint("  Altitude: %.2f m\n", dataParsed.Altitude)
			utils.VPrint("  Ignition: %v\n", dataParsed.Ignition)
			utils.VPrint("  Battery Level: %.2f V\n", dataParsed.BatteryLevel)
			utils.VPrint("  External Power: %.2f V\n", dataParsed.ExternalPower)
			utils.VPrint("  Input Status: %d\n", dataParsed.InputStatus)
			utils.VPrint("  Output Status: %d\n", dataParsed.OutputStatus)

			dataJSON, err := json.Marshal(dataParsed)
			if err != nil {
				return "", fmt.Errorf("error marshaling Queclink 320 data: %v", err)
			}
			return string(dataJSON), nil

		case "350":
			// Handle Queclink 350 model
			dataParsed, err := usecases.ParseQueclink350Fields(data)
			if err != nil {
				return "", fmt.Errorf("error: model 350 - %v - data %s", err, data)
			}

			utils.VPrint("Parsed Queclink 350 Data:\n")
			utils.VPrint("  Device Type: %s\n", dataParsed.DeviceType)
			utils.VPrint("  Device Version: %s\n", dataParsed.DeviceVersion)
			utils.VPrint("  IMEI: %s\n", dataParsed.IMEI)
			utils.VPrint("  Message Type: %s\n", dataParsed.MessageType)
			utils.VPrint("  Event Code: %s\n", dataParsed.EventCode)
			utils.VPrint("  Timestamp: %s\n", dataParsed.Timestamp.Format(time.RFC3339))
			utils.VPrint("  Latitude: %.6f\n", dataParsed.Latitude)
			utils.VPrint("  Longitude: %.6f\n", dataParsed.Longitude)
			utils.VPrint("  Speed: %.2f km/h\n", dataParsed.Speed)
			utils.VPrint("  Heading: %.2f degrees\n", dataParsed.Heading)
			utils.VPrint("  Satellites: %d\n", dataParsed.Satellites)
			utils.VPrint("  Ignition: %v\n", dataParsed.Ignition)
			utils.VPrint("  Battery Level: %.2f V\n", dataParsed.BatteryLevel)
			utils.VPrint("  External Power: %.2f V\n", dataParsed.ExternalPower)

			dataJSON, err := json.Marshal(dataParsed)
			if err != nil {
				return "", fmt.Errorf("error marshaling Queclink 350 data: %v", err)
			}
			return string(dataJSON), nil
		}
	} else {
		fmt.Println("Could not identify model from data")
	}
	return "", fmt.Errorf("no valid model identified from data: %s", data)
}
