package usecases

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"meitrackprotocol/features/meitrack_protocol/helpers"
	"meitrackprotocol/features/meitrack_protocol/models"
)

type DataParser struct {
	hexValue string
	index    int
}

func NewDataParser(data string) *DataParser {
	return &DataParser{hexValue: data}
}

func (p *DataParser) GetPart(length int) (string, error) {
	if p.index+length > len(p.hexValue) {
		return "", fmt.Errorf("out of range data %d", length)
	}
	part := p.hexValue[p.index : p.index+length]
	p.index += length
	return part, nil
}

func ParseCCEFields(cceFields *models.CCEModel) (string, error) {
	hexValue := hex.EncodeToString([]byte(cceFields.Rest))

	parser := NewDataParser(hexValue)

	part8, err := parser.GetPart(8)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	remainingCacheRecords := helpers.HexToLittleEndianDecimal(part8)
	cceFields.RemainingCacheRecords = remainingCacheRecords.(int)

	part4, err := parser.GetPart(4)
	if err != nil {
		return "", fmt.Errorf("error data too short: %v", err)
	}
	dataPackets := helpers.HexToLittleEndianDecimal(part4)
	cceFields.DataPackets = dataPackets.(int)

	cceFields.ListPackets = make(map[string]any)

	for i := 0; i < dataPackets.(int); i++ {
		packet, err := parsePacket(parser)
		if err != nil {
			if i == 0 {
				return "", fmt.Errorf("error packet %d: %v, number of packets %s, data: ", i, err, fmt.Sprintf("%v", dataPackets))
			} else {
				break
			}
		}
		cceFields.ListPackets[fmt.Sprintf("packet_%d", i+1)] = packet
	}
	jsonData, err := json.Marshal(cceFields)
	if err != nil {
		return "", fmt.Errorf("error json conversion")
	}
	return string(jsonData), nil
}

func parsePacket(parser *DataParser) (map[string]any, error) {
	packet := make(map[string]any)
	// data packet lenght
	parser.GetPart(4)
	// id bytes lenght
	numberIds, err := parser.GetPart(4)
	if err != nil {
		return nil, fmt.Errorf("error started parse packer %v", err)
	}

	parseIDs1, err := parseIDs(parser, models.IDOneByte, 4)
	if err != nil {
		return nil, fmt.Errorf("IDs1: %v, number of ids: %s", err, numberIds)
	}
	for k, v := range parseIDs1 {
		packet[k] = v
	}

	parseIDs2, err := parseIDs(parser, models.IDTwoBytes, 6)
	if err != nil {
		return nil, fmt.Errorf("IDs2: %v, number of ids: %s", err, numberIds)
	}
	for k, v := range parseIDs2 {
		packet[k] = v
	}

	parseIDs4, err := parseIDs(parser, models.IDFourBytes, 10)
	if err != nil {
		return nil, fmt.Errorf("IDs4: %v, number of ids: %s", err, numberIds)
	}
	for k, v := range parseIDs4 {
		packet[k] = v
	}

	undefinedIDs, err := parseUndefinedIDs(parser)
	if err != nil {
		return nil, fmt.Errorf("IDs Undefined: %v - %s", err, parser.hexValue)
	}
	for k, v := range undefinedIDs {
		packet[k] = v
	}

	return packet, nil
}

func parseIDs(parser *DataParser, mapId map[string]models.IDModel, byteLength int) (map[string]any, error) {
	ids := make(map[string]any)
	part2, err := parser.GetPart(2)
	if err != nil {
		return nil, fmt.Errorf("ids quantity - %v", err)
	}
	count := helpers.HexToInt(part2)
	if count.(int) == 0 {
		part2, err = parser.GetPart(2)
		count = helpers.HexToInt(part2)
	}
	idStr, err := parser.GetPart(count.(int) * byteLength)
	if err != nil {
		return nil, fmt.Errorf("length %d count * %d bytes not possible - %v", count.(int), byteLength, err)
	}

	index := 0
	for i := 0; i < count.(int); i++ {
		id := idStr[index : index+2]

		if index+(byteLength-2)+2 > len(idStr) {
			return nil, fmt.Errorf("id %s bytes %d - bytes not possible - %v", idStr, index+(byteLength-2)+2, err)
		}
		value := idStr[index+2 : index+(byteLength-2)+2]
		if id == "fe" {
			id = idStr[index : index+4]
			value, err = parser.GetPart(2)
			if err != nil {
				return nil, fmt.Errorf("length %d count * %d bytes not possible - %v", count.(int), byteLength, err)
			}
		}
		if model, exists := mapId[id]; exists {
			ids[model.Name] = model.Conversion(value)
		} else {
			ids[id] = value
		}
		index += 2 + (byteLength - 2)
	}
	return ids, nil
}

func parseUndefinedIDs(parser *DataParser) (map[string]any, error) {
	idsUndefined := make(map[string]any)
	part2, err := parser.GetPart(2)
	if err != nil {
		return nil, fmt.Errorf("first part %v", err)
	}
	count := helpers.HexToInt(part2)

	for i := 0; i < count.(int); i++ {
		id, err := parser.GetPart(2)
		if err != nil {
			return nil, fmt.Errorf("second part %v", err)
		}

		if id == "fe" {
			nextID, err := parser.GetPart(2)
			if err != nil {
				return nil, fmt.Errorf("error id started with fe %v", err)
			}
			id += nextID
		}

		part2Length, err := parser.GetPart(2)
		if err != nil {
			return nil, fmt.Errorf("third part %v", err)
		}
		length := helpers.HexToInt(part2Length)

		value, err := parser.GetPart(length.(int) * 2)
		if err != nil {
			return nil, fmt.Errorf("fourth part %v", err)
		}

		if model, exists := models.IDUndefinedBytes[id]; exists && model.Conversion != nil {
			idsUndefined[model.Name] = model.Conversion(value)
		} else {
			idsUndefined[id] = value
		}
	}
	return idsUndefined, nil
}
