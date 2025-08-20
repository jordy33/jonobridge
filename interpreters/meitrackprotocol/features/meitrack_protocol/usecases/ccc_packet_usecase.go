package usecases

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"meitrackprotocol/features/meitrack_protocol/config"
	"meitrackprotocol/features/meitrack_protocol/helpers"
	"meitrackprotocol/features/meitrack_protocol/models"
)

func ParseCCCFields(cccFields *models.CCCModel) (string, error) {
	hexValue := hex.EncodeToString([]byte(cccFields.Rest))
	parser := NewDataParser(hexValue)
	protocolVersionHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	protocolVersion := helpers.HexToLittleEndian(protocolVersionHex)
	lenghtPacketHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	lenghtPacket := helpers.HexToLittleEndianDecimal(lenghtPacketHex)
	numberOfRemainingCachesHex, err := parser.GetPart(8)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	numberOfRemainingCaches := helpers.HexToLittleEndianDecimal(numberOfRemainingCachesHex)
	eventCodeHex, err := parser.GetPart(2)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	eventCode := func(eventCodeHex string) interface{} {
		return helpers.FetchInEventCodes(eventCodeHex, config.EventCodes)
	}
	latitudeHex, err := parser.GetPart(8)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	latitude := helpers.LatLngValue(latitudeHex)
	longitudeHex, err := parser.GetPart(8)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	longitude := helpers.LatLngValue(longitudeHex)
	dateTimeHex, err := parser.GetPart(8)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	dateTime := helpers.DateAndTime(dateTimeHex)

	positioningStatusHex, err := parser.GetPart(2)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	positioningStatus, err := helpers.HexToUTF8(positioningStatusHex)
	if err != nil {
		return "", fmt.Errorf("error hex to utf-8: %v", err)
	}
	numberOfSatellitesHex, err := parser.GetPart(2)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	numberOfSatellites := helpers.HexToInt(numberOfSatellitesHex)
	speedHex, err := parser.GetPart(2)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	spped := helpers.HexToInt(speedHex)
	directionHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	direction := helpers.HexToLittleEndianDecimal(directionHex)
	horizontalPositioningHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	horizontalPositioning := helpers.HexToLittleEndianDecimal(horizontalPositioningHex)
	altitudeHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	altitude := helpers.HexToLittleEndianDecimal(altitudeHex)
	mileageHex, err := parser.GetPart(8)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	mileage := helpers.HexToLittleEndianDecimal(mileageHex)
	runtimeHex, err := parser.GetPart(8)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	runtime := helpers.HexToLittleEndianDecimal(runtimeHex)
	mccHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	mcc := helpers.HexToLittleEndianDecimal(mccHex)
	mncHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	mnc := helpers.HexToLittleEndianDecimal(mncHex)
	lacHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	lac := helpers.HexToLittleEndianDecimal(lacHex)
	ciHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	ci := helpers.HexToLittleEndianDecimal(ciHex)
	baseStationInfo := map[string]interface{}{
		"mcc": mcc,
		"mnc": mnc,
		"lac": lac,
		"ci":  ci,
	}

	ioPortStatusHex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	ioPortStatus := helpers.HexLittleEndianToBinary(ioPortStatusHex)

	ad1Hex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	ad1 := helpers.HexToLittleEndianDecimal(ad1Hex)

	ad4Hex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	ad4 := helpers.HexToLittleEndianDecimal(ad4Hex)
	ad5Hex, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	ad5 := helpers.HexToLittleEndianDecimal(ad5Hex)

	analogsInput := map[string]interface{}{
		"ad1": ad1,
		"ad4": ad4,
		"ad5": ad5,
	}

	geoFenceNumberHex, err := parser.GetPart(8)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	geoFenceNumber := helpers.HexToLittleEndianDecimal(geoFenceNumberHex)

	latitudeValue, ok := latitude.(float64)
	if !ok {
		return "", fmt.Errorf("latitude is not a valid float64")
	}

	longitudeValue, ok := longitude.(float64)
	if !ok {
		return "", fmt.Errorf("longitude is not a valid float64")
	}

	horizontalPositioningValue, ok := horizontalPositioning.(int)
	if !ok {
		return "", fmt.Errorf("horizontalPositioning is not a valid int")
	}

	altitudeValue, ok := altitude.(int)
	if !ok {
		return "", fmt.Errorf("altitude is not a valid int")
	}
	lenghtPacketInt, ok := lenghtPacket.(int)
	if !ok {
		return "", fmt.Errorf("lenghtPacket is not a valid int")
	}

	numberOfRemainingCachesInt, ok := numberOfRemainingCaches.(int)
	if !ok {
		return "", fmt.Errorf("numberOfRemainingCaches is not a valid int")
	}

	sppedInt, ok := spped.(int)
	if !ok {
		return "", fmt.Errorf("spped is not a valid int")
	}

	directionInt, ok := direction.(int)
	if !ok {
		return "", fmt.Errorf("direction is not a valid int")
	}

	mileageInt, ok := mileage.(int)
	if !ok {
		return "", fmt.Errorf("mileage is not a valid int")
	}

	runtimeInt, ok := runtime.(int)
	if !ok {
		return "", fmt.Errorf("runtime is not a valid int")
	}

	geoFenceNumberInt, ok := geoFenceNumber.(int)
	if !ok {
		return "", fmt.Errorf("geoFenceNumber is not a valid int")
	}

	numberOfSatellitesInt, ok := numberOfSatellites.(int)
	if !ok {
		return "", fmt.Errorf("numberOfSatellites is not a valid int")
	}

	cccModel := models.CCCModel{
		GeneralModel:            cccFields.GeneralModel,
		ProtocolVersion:         fmt.Sprintf("%v", protocolVersion),
		PacketLength:            lenghtPacketInt,
		NumberOrRemainingCaches: numberOfRemainingCachesInt,
		EventCode:               eventCode(eventCodeHex),
		Latitude:                latitudeValue,
		Longitude:               longitudeValue,
		Datetime:                fmt.Sprintf("%v", dateTime),
		PositioningStatus:       positioningStatus,
		NumberOfSatellites:      numberOfSatellitesInt,
		Speed:                   sppedInt,
		Direction:               directionInt,
		HorizontalPositioning:   horizontalPositioningValue,
		Altitude:                altitudeValue,
		Mileage:                 mileageInt,
		RunTime:                 runtimeInt,
		BaseStationInfo:         baseStationInfo,
		IoPortStatus:            fmt.Sprintf("%v", ioPortStatus),
		AnalogsInput:            analogsInput,
		GeoFenceNumber:          geoFenceNumberInt,
	}

	jsonData, err := json.Marshal(cccModel)
	if err != nil {
		return "", fmt.Errorf("error converting to JSON: %v", err)
	}

	return string(jsonData), nil
}
