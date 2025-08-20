package helpers

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

func ParseCoordinate(data []byte, isLongitude bool) float64 {
	// Convertir los 4 bytes a un entero de 32 bits con signo (int32)
	value := int32(data[0])<<24 | int32(data[1])<<16 | int32(data[2])<<8 | int32(data[3])

	coordinate := float64(value) / (30000.0 * 60.0)

	// Validar rango de latitud y longitud
	if isLongitude && (coordinate < -180 || coordinate > 180) {
		fmt.Printf("Longitud inválida: %.6f\n", coordinate)
		return -1
	} else if !isLongitude && (coordinate < -90 || coordinate > 90) {
		fmt.Printf("Latitud inválida: %.6f\n", coordinate)
		return -1
	}
	if isLongitude && coordinate > 0 {
		coordinate = -coordinate
	}

	return coordinate
}

func CalculateCRC(data []byte) []byte {
	var crc uint16 = 0x0000
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return []byte{byte(crc >> 8), byte(crc & 0xFF)}
}

func BytesToHexAndDecimal(data []byte) (int, error) {
	// Convertir a hexadecimal
	hexValue := hex.EncodeToString(data)

	// Convertir hexadecimal a decimal
	decimalValue := new(big.Int)
	_, ok := decimalValue.SetString(hexValue, 16) // Base 16 (hexadecimal)
	if !ok {
		return 0, fmt.Errorf("error converting hex to decimal")
	}

	// Verificar si el valor cabe en un int
	if !decimalValue.IsInt64() {
		return 0, fmt.Errorf("value exceeds int64 range")
	}

	return int(decimalValue.Int64()), nil
}
