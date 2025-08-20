package helpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type EventCode struct {
	Code int    `json:"Code"`
	Name string `json:"Name"`
}

type CodeModel struct {
	EventCode EventCode `json:"EventCode"`
}

func Spliter(dat []byte) []string {
	byt := 0

	var s []string
	for _, v := range dat {
		byt += 1
		s = append(s, fmt.Sprintf("%02x", v))
	}

	return s
}

func Hexi(s []string, start, bytes int) string {
	var value string
	final := start + bytes
	for i := start; i < final; i++ {
		value += s[i]
	}
	return value
}

// Mapa para almacenar los nombres de los parámetros de un byte
var oneByteParameterMap = map[int64]string{
	2:   "Din1",
	3:   "Din2",
	4:   "Din3",
	5:   "Ignition(Din4)",
	6:   "ModemTemperature",
	7:   "TrackByTimeInterval",
	8:   "TrackByDistance",
	10:  "CanRequestsSupported",
	11:  "CanDiagnosticsSupported",
	14:  "CanTachoDriver1TimeRelatedStatus",
	15:  "CanTachoDriver1Card",
	16:  "CanTachoDriver2TimeRelatedStatus",
	17:  "TachoCardReaderState",
	24:  "TachoCardReaderState",
	25:  "SecurityInfo",
	27:  "GsmUmtsSignalLevel",
	31:  "CanEngineBraking",
	32:  "PcbTemperature",
	35:  "CanClutchSwitch",
	36:  "CanBrakeSwitch",
	37:  "CanCruiseControlActive",
	38:  "CanPtoState",
	39:  "CanEnginePercentLoadAtCurrentSpeed",
	40:  "CanTireLocation",
	42:  "CanSystemEvent",
	43:  "CanTachoHandlingInformation",
	44:  "CanTachoDirectionIndicator",
	49:  "AccelerometerX",
	50:  "AccelerometerY",
	51:  "AccelerometerZ",
	56:  "DigitalFuelSensorC1Temperature",
	57:  "DigitalFuelSensorC2Temperature",
	58:  "DigitalFuelSensorC3Temperature",
	59:  "DigitalFuelSensorC4Temperature",
	60:  "DigitalFuelSensorC5Temperature",
	61:  "DigitalFuelSensorC6Temperature",
	62:  "DigitalFuelSensorC7Temperature",
	63:  "DigitalFuelSensorC8Temperature",
	64:  "DigitalFuelSensorC9Temperature",
	75:  "DigitalFuelSensorC10Temperature",
	76:  "DigitalFuelSensorB1Temperature",
	88:  "GsmUmtsJamming",
	95:  "ObdVehicleSpeed",
	130: "EcoMaxSpeed",
	176: "Speed",
	190: "FridgeHighSpeedStatus",
	218: "CanTachoVehicleOverspeed",
	423: "MeZeroSpeed",
	651: "Geofence",
	718: "FatigueSensorSnapshotOnSdCard",
	719: "DoutActivationByRollover",
	303: "TellTaleBatteryChargingCondition",
	923: "VehicleFuelLevelPercent",
	177: "FridgeFuelLevel",
	207: "CanFuelLevel1",
	209: "CanEngineFuelLevelSecondary",
	301: "TellTaleFuelLevel",
	481: "CanFuelLevel2",
}

// Mapa para almacenar descripciones específicas de valores
var valueDescriptions = map[int64]map[string]string{
	2: {
		"0": "0:LowLevelOnInput",
		"1": "1:HighLevelOnInput",
	},
	3: {
		"0": "0:LowLevelOnInput",
		"1": "1:HighLevelOnInput",
	},
	4: {
		"0": "0:LowLevelOnInput",
		"1": "1:HighLevelOnInput",
	},
	5: {
		"0": "0:LowLevelOnInput",
		"1": "1:HighLevelOnInput",
	},
	10: {
		"0": "0:RequestIsNotSupported",
		"1": "1:RequestIsSupported",
		"3": "3:Don'tCare",
	},
	11: {
		"0": "0:DiagnosticsIsNotSupported",
		"1": "1:DiagnosticsIsSupported",
		"3": "3:Don'tCare",
	},
	14: {
		"0": "0:CardNotPresent",
		"1": "1:CardPresent",
		"2": "2:Error",
		"3": "3:NotAvailable",
	},
	16: {
		"0":  "0:Normal",
		"1":  "1:15MinBefore4AndHalfH",
		"2":  "2:4AndHalfHReached",
		"3":  "3:15MinBefore9H",
		"4":  "4:9HReached",
		"5":  "5:15MinBefore16H",
		"6":  "6:16HReached",
		"14": "14:Error",
		"15": "15:NotAvailable",
	},
	17: {
		"0": "0:CardNotPresent",
		"1": "1:CardPresent",
	},
	24: {
		"0": "0:NotAvailable",
		"1": "1:Available",
	},
	25: {
		"1": "1:ConfigurationChangeAttemptWhenPlockEnabled",
		"2": "2:PlockCodeSendingAttemptFailed",
		"4": "4:PlockCodeSendingAttemptFailed",
	},
	27: {
		"0":  "24:GpsSignalLost",
		"1":  "1:VeryWeak",
		"31": "Excellent",
	},
	35: {
		"0": "PedalReleased",
		"1": "PedalPressed",
		"2": "2:Error",
		"3": "3:NotAvailable",
	},
	36: {
		"0": "PedalReleased",
		"1": "PedalPressed",
		"2": "2:Error",
		"3": "3:NotAvailable",
	},
	37: {
		"0": "SwitchedOff",
		"1": "SwitchedOn",
		"2": "2:Error",
		"3": "3:NotAvailable",
	},
	38: {
		"0":  "OffDisabled",
		"5":  "Set",
		"31": "NotAvailable",
	},
	651: {
		"0": "21",
		"1": "20",
	},
	718: {
		"0": "135",
	},
	719: {
		"2": "72",
	},
	303: {
		"1": "18",
	},
}

func BodyExtendedRecords(s []string) []map[string]string {
	// Minimum required length for basic processing (IMEI + header)
	if len(s) < 20 {
		return []map[string]string{}
	}

	mapasBridge := []map[string]string{}
	imei := Hexi(s, 2, 8)
	imeiDec, _ := strconv.ParseInt(imei, 16, 64)
	imeiStr := strings.Trim(fmt.Sprint(imeiDec), " ")

	mapaBridge := map[string]string{
		"IMEI": imeiStr,
	}

	mapaBridge = processHeader(mapaBridge, s)
	mapasBridge = append(mapasBridge, mapaBridge)

	return mapasBridge
}

func processHeader(mapaBridge map[string]string, s []string) map[string]string {
	// Validate input length before processing
	if len(s) < 20 {
		return mapaBridge
	}

	// Procesa parámetros principales del encabezado
	timeStamp2 := Hexi(s, 0, 4)
	timeStamp2Int, _ := strconv.ParseInt(timeStamp2, 16, 64)
	timeStampUTC2 := currentUTC(timeStamp2Int)
	fillMap(mapaBridge, "Datetime", strings.Trim(timeStampUTC2, " "))

	long := Hexi(s, 7, 4)
	longFloat := hexToFloat(long, 10000000)
	fillMap(mapaBridge, "Longitude", longFloat)

	lat := Hexi(s, 11, 4)
	latFloat := hexToFloat(lat, 10000000)
	fillMap(mapaBridge, "Latitude", latFloat)

	alt := Hexi(s, 15, 2)
	altFloat := hexToFloat(alt, 10)
	fillMap(mapaBridge, "Altitude", altFloat)

	// Angle
	angle := Hexi(s, 17, 2)
	angleDec, _ := strconv.ParseInt(angle, 16, 64)
	angleDec2 := float64(angleDec) / 100.0
	angleStr := strings.Trim(fmt.Sprint(angleDec2), " ")
	fillMap(mapaBridge, "Direction", angleStr)

	// Number of Satellites
	num := Hexi(s, 19, 1)
	numDec, _ := strconv.ParseInt(num, 16, 64)
	numStr := strings.Trim(fmt.Sprint(numDec), " ")
	fillMap(mapaBridge, "NumberOfSatellites", numStr)

	// Speed km/h
	spd := Hexi(s, 20, 2)
	spdDec, _ := strconv.ParseInt(spd, 16, 64)
	spdStr := strings.Trim(fmt.Sprint(spdDec), " ")
	fillMap(mapaBridge, "Speed", spdStr)

	// HDOP
	hdop := Hexi(s, 22, 1)
	hdopDec, _ := strconv.ParseInt(hdop, 16, 64)
	hdopDec2 := float64(hdopDec) / 10.0
	hdopStr := strings.Trim(fmt.Sprint(hdopDec2), " ")
	fillMap(mapaBridge, "Hdop", hdopStr)

	// Event ID
	eventID := Hexi(s, 23, 2)
	eventIdDec, _ := strconv.ParseInt(eventID, 16, 64)
	eventIdStr := fmt.Sprintf("%d", eventIdDec)
	fillMap(mapaBridge, "EventCode", eventIdStr)

	return mapaBridge
}

func fillMap(mapa map[string]string, name, value string) map[string]string {
	if value != "" {
		mapa[name] = value
	}
	return mapa
}

func currentUTC(timestamp int64) string {
	t := time.Unix(timestamp, 0).UTC()
	out := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return out
}

func hexToFloat(hexStr string, divisor float64) string {
	dec, _ := strconv.ParseInt(hexStr, 16, 64)
	floatValue := float64(int32(dec)) / divisor
	return strings.Trim(fmt.Sprint(floatValue), " ")
}
