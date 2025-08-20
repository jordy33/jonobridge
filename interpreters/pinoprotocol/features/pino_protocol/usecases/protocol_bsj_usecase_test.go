package usecases

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestParseLocationData(t *testing.T) {
	// Datos de prueba (hexadecimal)
	rawDataHex := "7e0200007d020990744775950006000000000000000000000000000000000000000000002501150712320104000035e4300119310100eb54000c00b28952020924191082248f00060089ffffffff000600c5ffffffff0003010204000400ce018f000b00d8014e14025a024b7d02050004002d0f96000300a85a001100d5383630363939303734343737353935627e"

	// Decodificar el string hexadecimal a bytes
	rawData, err := hex.DecodeString(rawDataHex)
	if err != nil {
		t.Fatalf("Error decoding raw data: %v", err)
	}

	// Remover los delimitadores 0x7E y separar data y checksum
	if len(rawData) < 2 || rawData[0] != 0x7E || rawData[len(rawData)-1] != 0x7E {
		t.Fatalf("Invalid raw data: missing 0x7E delimiters")
	}
	data := rawData[1 : len(rawData)-1]

	// Extraer phoneNumber del mock
	phoneNumber := DecodeBCD(data[4:11]) // Teléfono (BCD)

	// Llamar a ParseLocationData
	jsonResult := ParseLocationData(data[12:], phoneNumber, rawData)

	// Validar que el resultado no sea vacío
	if jsonResult == "" {
		t.Fatalf("ParseLocationData returned an empty string")
	}

	// Validar que el JSON tenga las claves esperadas
	expectedKeys := []string{"IMEI", "AlarmSign", "Status", "Latitude", "Longitude", "Datetime", "Speed", "Direction", "Elevation"}
	for _, key := range expectedKeys {
		if !containsKey(jsonResult, key) {
			t.Errorf("Key %s is missing in the JSON result", key)
		}
	}

	// Explicitly print the status bytes to check hemisphere information
	if len(data) >= 16 {
		t.Logf("Status bytes: %X", data[12+4:12+8])
		statusInt := bytesToInt(data[12+4 : 12+8])
		t.Logf("Status integer: %d", statusInt)
		t.Logf("North Latitude: %v", (statusInt&(1<<2)) == 0)
		t.Logf("East Longitude: %v", (statusInt&(1<<3)) == 0)

		// Print raw coordinate bytes
		t.Logf("Raw latitude bytes: %X", data[12+8:12+12])
		t.Logf("Raw longitude bytes: %X", data[12+12:12+16])

		// Print raw integer values
		latInt := bytesToInt(data[12+8 : 12+12])
		longInt := bytesToInt(data[12+12 : 12+16])
		t.Logf("Raw latitude integer: %d", latInt)
		t.Logf("Raw longitude integer: %d", longInt)

		// Calculate expected coordinates per protocol
		latFloat := float64(latInt) / 1000000.0
		longFloat := float64(longInt) / 1000000.0
		if (statusInt & (1 << 2)) != 0 {
			latFloat = -latFloat
		}
		if (statusInt & (1 << 3)) != 0 {
			longFloat = -longFloat
		}
		t.Logf("Expected coordinates: Lat=%f, Long=%f", latFloat, longFloat)
	}

	// Imprimir el JSON para inspección
	t.Logf("Resulting JSON: %s", jsonResult)
}
func TestParseLocationData2(t *testing.T) {
	// Datos de prueba (hexadecimal)
	rawDataHex := "7e020000710990744775950009000000000000000b0129de2305e9d9d208f8000000002501152225170104000035e430011f31010deb47000c00b28952020924191082248f00060089ffffffff000600c5ffffbfff0003010204000400ce01890004002d0f5d000300a85a001100d5383630363939303734343737353935eb7e"

	// Decodificar el string hexadecimal a bytes
	rawData, err := hex.DecodeString(rawDataHex)
	if err != nil {
		t.Fatalf("Error decoding raw data: %v", err)
	}

	// Remover los delimitadores 0x7E y separar data y checksum
	if len(rawData) < 2 || rawData[0] != 0x7E || rawData[len(rawData)-1] != 0x7E {
		t.Fatalf("Invalid raw data: missing 0x7E delimiters")
	}
	data := rawData[1 : len(rawData)-1]

	// Extraer phoneNumber del mock
	phoneNumber := DecodeBCD(data[4:11]) // Teléfono (BCD)

	// Llamar a ParseLocationData
	jsonResult := ParseLocationData(data[12:], phoneNumber, rawData)

	// Validar que el resultado no sea vacío
	if jsonResult == "" {
		t.Fatalf("ParseLocationData returned an empty string")
	}

	// Validar que el JSON tenga las claves esperadas
	expectedKeys := []string{"IMEI", "AlarmSign", "Status", "Latitude", "Longitude", "Datetime", "Speed", "Direction", "Elevation"}
	for _, key := range expectedKeys {
		if !containsKey(jsonResult, key) {
			t.Errorf("Key %s is missing in the JSON result", key)
		}
	}

	// Imprimir el JSON para inspección
	t.Logf("Resulting JSON: %s", jsonResult)
}

// containsKey verifica si un JSON contiene una clave específica
func containsKey(jsonStr string, key string) bool {
	return strings.Contains(jsonStr, "\""+key+"\":")
}
