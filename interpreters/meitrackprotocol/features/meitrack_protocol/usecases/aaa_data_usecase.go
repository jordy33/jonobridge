package usecases

import (
	"encoding/json"
	"fmt"
	"meitrackprotocol/features/meitrack_protocol/config"
	"meitrackprotocol/features/meitrack_protocol/helpers"
	"meitrackprotocol/features/meitrack_protocol/models"
	"strconv"
	"strings"
)

func ParseAAAFields(aaaFields *models.AAAModel) (string, error) {
	parts := strings.Split(aaaFields.Rest, ",")
	datetime, err := helpers.ParseDatetime(parts[3])
	if err != nil {
		return "", fmt.Errorf("error parse datetime")
	}

	if len(parts) >= 18 {

		aaaFields.EventCode = helpers.FetchInEventCodes(parts[0], config.EventCodesAAA)
		aaaFields.Latitude, _ = strconv.ParseFloat(parts[1], 32)
		aaaFields.Longitude, _ = strconv.ParseFloat(parts[2], 32)
		aaaFields.Datetime = datetime
		aaaFields.PositioningStatus = parts[4]
		aaaFields.NumberOfSatellites, _ = strconv.Atoi(parts[5])
		aaaFields.GsmSignalStrength, _ = strconv.Atoi(parts[6])
		aaaFields.Speed, _ = strconv.Atoi(parts[7])
		aaaFields.Direction, _ = strconv.Atoi(parts[8])
		aaaFields.HDOP, _ = strconv.ParseFloat(parts[9], 32) // Changed from Hdop to HDOP
		aaaFields.Altitude, _ = strconv.ParseFloat(parts[10], 32)
		aaaFields.Mileage, _ = strconv.Atoi(parts[14]) // Correct the Mileage position from 11 to 14
		aaaFields.RunTime, _ = strconv.Atoi(parts[12])
		aaaFields.BaseStationInfo = helpers.BaseStationCommandTypeAAA(parts[13])
		aaaFields.IoPortStatus = parts[14]
		aaaFields.AnalogInputs = helpers.AnalogsInputCommandTypeAAA(parts[15])
		aaaFields.AssistedEventInfo = parts[16]
	}
	if len(parts) == 23 {
		aaaFields.CustomizedData = parts[17]
		aaaFields.ProtocolVersion, _ = strconv.Atoi(parts[18])
		aaaFields.FuelPercentage = parts[19]
		aaaFields.TemperatureSensor = parts[20]
		aaaFields.MaxAcceleration, _ = strconv.Atoi(parts[21])
		values := strings.Split(parts[22], "*")
		if len(values) == 2 {
			aaaFields.MaxDesceleration, _ = strconv.Atoi(values[0])
			aaaFields.Checksum = values[1]
		} else {
			return "", fmt.Errorf("undefined checksum")
		}
	}
	jsonData, err := json.Marshal(aaaFields)
	if err != nil {
		return "", fmt.Errorf("error json conversion")
	}
	return string(jsonData), nil
}
