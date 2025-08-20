package usecases

import (
	"ruptelaprotocol/features/ruptela_protocol/helpers"
	"strconv"
)

func Conversion(passline []byte) ([]map[string]string, []byte, error) {
	var ack []byte

	dataSplit := helpers.Spliter(passline)

	cr16_fromData := helpers.Hexi(dataSplit, len(dataSplit)-2, 2)
	cr16_fromDataInt, _ := strconv.ParseInt(cr16_fromData, 16, 64)
	cr16_fromUs := helpers.Crc16Funtion(passline[2 : len(passline)-2])

	if uint16(cr16_fromDataInt) == cr16_fromUs {
		ack = helpers.Acknowledge()
	} // else {
	// 	fmt.Println("Corrupt Data")
	// }

	mapasBridge := []map[string]string{}
	mapasBridge = helpers.BodyExtendedRecords(dataSplit)

	return mapasBridge, ack, nil
}
