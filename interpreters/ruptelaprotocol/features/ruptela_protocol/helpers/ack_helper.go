package helpers

import (
	"encoding/hex"
	"log"
)

func Acknowledge() []byte {
	data, err := hex.DecodeString("6401")
	if err != nil {
		log.Println("Error en ACK:", err)
	}
	crc := Crc16Funtion(data)
	return append(data, byte(crc>>8), byte(crc))
}
