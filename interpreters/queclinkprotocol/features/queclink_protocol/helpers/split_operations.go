package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

func BaseStationCommandTypeAAA(baseStationInfo string) any {
	parts := strings.Split(baseStationInfo, "|")

	return map[string]any{
		"mmc":     parts[0],
		"mnc":     parts[1],
		"lac":     parts[2],
		"cellId":  parts[3],
		"rxLevel": "-1",
	}
}

func AnalogsInputCommandTypeAAA(anlogsInput string) any {
	parts := strings.Split(anlogsInput, "|")
	result := make(map[string]string)

	for i, part := range parts {
		key := fmt.Sprintf("AD%d", i+1)
		result[key] = part
	}

	return result
}

func BaseStationInfo(hexString string) any {
	// Parse hexString into proper values based on CCE protocol documentation
	// MCC: 16-bit unsigned; little-endian (2 bytes)
	// MNC: 16-bit unsigned; little-endian (2 bytes)
	// LAC: 16-bit unsigned; little-endian (2 bytes)
	// CELL_ID: 32-bit unsigned; little-endian (4 bytes)
	// RX_LEVEL: 16-bit signed; little-endian (2 bytes)

	if len(hexString) < 24 { // Minimum length for all fields (12 bytes = 24 hex chars)
		return map[string]any{
			"mcc":     nil,
			"mnc":     nil,
			"lac":     nil,
			"cellId":  nil,
			"rxLevel": nil,
		}
	}

	// Extract fields from hex string
	mcc := HexToLittleEndianDecimal(hexString[:4])
	mnc := HexToLittleEndianDecimal(hexString[4:8])
	lac := HexToLittleEndianDecimal(hexString[8:12])
	cellId := HexToLittleEndianDecimal(hexString[12:20])
	rxLevel := HexToLittleEndianDecimal(hexString[20:24])

	// Convert to proper string format for consistency
	mccStr := fmt.Sprintf("%v", mcc)
	mncStr := fmt.Sprintf("%v", mnc)
	lacStr := fmt.Sprintf("%v", lac)
	cellIdStr := fmt.Sprintf("%v", cellId)
	rxLevelStr := fmt.Sprintf("%v", rxLevel)

	return map[string]any{
		"mcc":     mccStr,
		"mnc":     mncStr,
		"lac":     lacStr,
		"cellId":  cellIdStr,
		"rxLevel": rxLevelStr,
	}
}

func NetworkInformation(hexString string) any {
	descriptorLen, ok := HexToLittleEndianDecimal(hexString[4:6]).(int)
	if !ok {
		descriptorLen = 6
	}
	descriptorHex := hexString[6 : 6+(descriptorLen*2)]
	descriptor, err := HexToUTF8(descriptorHex)
	//LTE
	if err != nil {
		descriptor = descriptorHex
	}
	return map[string]any{
		"version":      hexString[:2],
		"type":         hexString[2:4],
		"decriptorLen": descriptorLen,
		"descriptor":   descriptor,
	}
}

func CameraStatus(hexString string) any {
	status := HexLittleEndianToBinary(hexString[2:])
	return map[string]any{
		"camerasNumber": HexToLittleEndianDecimal(hexString[0:2]),
		"status":        status,
	}

}

func AdditionalAlertInfo(hexString string, alarmTypesFirstProtocol map[string]any, alarmTypesSecondProtocol map[string]any) any {
	alarmProtocol := hexString[:2]
	alarmTypeHex := hexString[2:4]
	alertInfo := make(map[string]any)
	alertInfo["alarmProtocol"] = alarmProtocol
	if alarmProtocol == "02" {
		alertInfo["alarmType"] = FetchInAlarmTypes(alarmTypeHex, alarmTypesSecondProtocol)
	}
	if alarmProtocol == "01" {
		alertInfo["alarmType"] = FetchInAlarmTypes(alarmTypeHex, alarmTypesFirstProtocol)
	}
	photoName := hexString[4:]
	if photoName == "00" {
		photoName = "Photo doesn't exist"
	} else {
		photoName, _ = HexToUTF8(photoName)
	}
	alertInfo["photoName"] = photoName
	return alertInfo
}

func FatigueDrivingInfo(hexString string, alarmTypes map[string]any) any {
	alarmProtocol := hexString[:2]
	alarmTypeHex := hexString[2:4]
	alertInfo := make(map[string]any)
	alertInfo["alarmProtocol"] = alarmProtocol
	alertInfo["alarmType"] = FetchInAlarmTypes(alarmTypeHex, alarmTypes)

	photoName := hexString[4:]
	if photoName == "00" {
		photoName = "Photo doesn't exist"
	} else {
		photoName, _ = HexToUTF8(photoName)
	}
	alertInfo["photoName"] = photoName
	return alertInfo
}

func PictureName(hexString string) any {
	time := HexToLittleEndianDecimal(hexString[:8]).(int)
	timeStr := strconv.Itoa(time)
	last_part := hexString[8:]
	name := timeStr + last_part
	return name
}

func TemperatureSensor(hexString string) any {
	sensorNumber := hexString[:2]
	intValue := HexToLittleEndianDecimal(hexString).(int)
	bitSize := len(hexString) * 4
	signedIntValue := TwosComplement(intValue, bitSize)
	decimalValue := signedIntValue / 100
	return map[string]any{
		"sensorNumber": sensorNumber,
		"value":        decimalValue,
	}
}

func AdditionalInfoBluetooth(hexString string, alarmTypes map[string]any) any {
	version := hexString[:2]
	alertType := FetchInAlarmTypes(hexString[2:4], alarmTypes)
	data := hexString[4:]
	return map[string]any{
		"version":   version,
		"alertType": alertType,
		"data":      data,
	}
}

func BluetoothBeacon(hexString string) any {
	version := hexString[:2]
	lengthDeviceName := HexToInt(hexString[2:4]).(int)
	deviceName, _ := HexToUTF8(hexString[4 : 4+lengthDeviceName])
	mac, _ := HexToUTF8(hexString[4+lengthDeviceName : lengthDeviceName+16])
	batteryPower := hexString[16+lengthDeviceName : lengthDeviceName+18]
	signalStrength := hexString[18+lengthDeviceName : lengthDeviceName+20]
	return map[string]any{
		"version":          version,
		"lengthDeviceName": lengthDeviceName,
		"deviceName":       deviceName,
		"mac":              mac,
		"batteryPower":     batteryPower,
		"signalStrength":   signalStrength,
	}
}

func TemperatureAndHumidity(hexString string) any {
	lengthDeviceName := HexToInt(hexString[2:4]).(int)
	data := BluetoothBeacon(hexString).(map[string]any)
	data["temperature"] = HexToInt(hexString[18+lengthDeviceName : lengthDeviceName+22])
	data["humidity"] = HexToInt(hexString[22+lengthDeviceName : lengthDeviceName+26])
	data["alertHighTemperature"] = HexToInt(hexString[22+lengthDeviceName : lengthDeviceName+26])
	data["alertLowTemperature"] = HexToInt(hexString[26+lengthDeviceName : lengthDeviceName+30])
	data["alertHighHumidity"] = HexToInt(hexString[30+lengthDeviceName : lengthDeviceName+34])
	data["alertLowHumidity"] = HexToInt(hexString[34+lengthDeviceName : lengthDeviceName+38])
	return data
}
