package usecases

import (
	"encoding/hex"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeStandardLocationData(t *testing.T) {
	// Datos de prueba
	hexString := "78781f121901100e3523cf021b2c940aa469df05dcde014e322602000000013a503b0d0a"
	data, err := hex.DecodeString(hexString)
	assert.NoError(t, err, "Error decoding hex string")

	imei := "123456789012345" // IMEI ficticio
	isAlarm := false

	// Llamar a la función que se está probando
	result, err := DecodeStandardLocationData(data, imei, isAlarm)

	// Validar que Extra esté vacío porque el campo isAlarm es false
	assert.Equal(t, "", result.Extra, "Extra mismatch")
}
func TestDecodeStandardLocationData2(t *testing.T) {
	// Datos de prueba
	hexString := "78788b15830000000143757272656e7420706f736974696f6e3a4c61743a4e31392e3532313038352c4c6f6e3a5739392e3231313636392c4461746554696d653a323032352d30322d31332031373a34323a35372c687474703a2f2f6d6170732e676f6f676c652e636f6d2f6d6170733f713d4e31392e3532313038352c5739392e3231313636390002001d1fc00d0a"
	data, err := hex.DecodeString(hexString)
	assert.NoError(t, err, "Error decoding hex string")

	imei := "123456789012345" // IMEI ficticio
	isAlarm := false

	// Llamar a la función que se está probando
	result, err := DecodeStandardLocationData(data, imei, isAlarm)

	// Validar que Extra esté vacío porque el campo isAlarm es false
	assert.Equal(t, "", result.Extra, "Extra mismatch")
}
