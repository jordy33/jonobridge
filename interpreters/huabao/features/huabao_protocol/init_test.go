package huabao_protocol

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	// Trama de prueba
	testData := "$$f167,864507035846483,AAA,1,18.950273,-97.922888,241205120405,V,0,13,0,69,0.0,2217,358868041,192062311,334|3|7663|00AA7FAB,0000,0001|0000|0000|01A5|0514,,,3,,,108,106*C6"

	// Resultado esperado (estructura legible)
	expectedData := `{
		"StartSignal": "$$",
		"Identifier": "f",
		"DataLength": "167",
		"IMEI": "864507035846483",
		"CommandType": "AAA",
		"Rest": "1,18.950273,-97.922888,241205120405,V,0,13,0,69,0.0,2217,358868041,192062311,334|3|7663|00AA7FAB,0000,0001|0000|0000|01A5|0514,,,3,,,108,106*C6",
		"Message": "$$f167,864507035846483,AAA,1,18.950273,-97.922888,241205120405,V,0,13,0,69,0.0,2217,358868041,192062311,334|3|7663|00AA7FAB,0000,0001|0000|0000|01A5|0514,,,3,,,108,106*C6",
		"EventCode": {"Code": 1, "Name": "Input 1 Active"},
		"Latitude": 18.950273513793945,
		"Longitude": -97.92288970947266,
		"Datetime": "2024-12-05T12:04:05Z",
		"PositioningStatus": "V",
		"NumberOfSatellites": 0,
		"GsmSignalStrength": 13,
		"Speed": 0,
		"Direction": 69,
		"Hdop": 0,
		"Altitude": 2217,
		"Mileage": 358868041,
		"RunTime": 192062311,
		"BaseStationInfo": {"cellId": "00AA7FAB", "lac": "7663", "mmc": "334", "mnc": "3", "rxLevel": "-1"},
		"IoPortStatus": "0000",
		"AnalogInputs": {"AD1": "0001", "AD2": "0000", "AD3": "0000", "AD4": "01A5", "AD5": "0514"},
		"AssistedEventInfo": "",
		"CustomizedData": "",
		"ProtocolVersion": 3,
		"FuelPercentage": "",
		"TemperatureSensor": "",
		"MaxAcceleration": 108,
		"MaxDesceleration": 106,
		"Checksum": "C6"
	}`

	result, err := Initialize(testData)

	assert.NoError(t, err, "La función devolvió un error inesperado")

	var expectedNormalized, resultNormalized map[string]interface{}

	err = json.Unmarshal([]byte(expectedData), &expectedNormalized)
	assert.NoError(t, err, "Error al parsear el JSON esperado")

	err = json.Unmarshal([]byte(result), &resultNormalized)
	assert.NoError(t, err, "Error al parsear el JSON actual")

	assert.Equal(t, expectedNormalized, resultNormalized, "El resultado no coincide con el esperado")
}

func TestInitializeWithBinaryFile(t *testing.T) {

	filePath := "data.bin"

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("No se pudo leer el archivo %s: %v", filePath, err)
	}

	dataString := string(data)

	result, err := Initialize(dataString)
	assert.NoError(t, err, "La función devolvió un error inesperado")

	expectedData := map[string]interface{}{
		"StartSignal":           "$$",
		"Identifier":            "[",
		"DataLength":            "139",
		"IMEI":                  "866811062546604",
		"CommandType":           "CCE",
		"Rest":                  "\u0000\u0000\u0000\u0000\u0001\u0000i\u0000\u0017\u0000\u0005\u0005\u0000\u0006\u0000\u0007\n\u0014\u0000\u0015\u0002\t\b\u0000\u0000\t\u0000\u0000\n\u0000\u0000\u000b\u0000\u0000\u0016\t\u0000\u0017\u0005\u0000\u0019\u0001\u0000\u001a\ufffd\u0004@#\u0000\u0006\u0002\ufffd\ufffd)\u0001\u0003\ufffd\u0026\u0016\ufffd\u0004jv.\f\u0000\u0000\u0000\u0000\rH.\u0001\u0000\u001c\u0000 \u0000\u0000\u0003\u000e\fN\u0001\u0014\u0000Z\u0002\u0019~K\u0002\u0000\u0000I\t\u0004\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000K\u0005\u0001\u0001\u00024G*A5\r\n",
		"Message":               "$$[139,866811062546604,CCE,\u0000\u0000\u0000\u0000\u0001\u0000i\u0000\u0017\u0000\u0005\u0005\u0000\u0006\u0000\u0007\n\u0014\u0000\u0015\u0002\t\b\u0000\u0000\t\u0000\u0000\n\u0000\u0000\u000b\u0000\u0000\u0016\t\u0000\u0017\u0005\u0000\u0019\u0001\u0000\u001a\ufffd\u0004@#\u0000\u0006\u0002\ufffd\ufffd)\u0001\u0003\ufffd\u0026\u0016\ufffd\u0004jv.\f\u0000\u0000\u0000\u0000\rH.\u0001\u0000\u001c\u0000 \u0000\u0000\u0003\u000e\fN\u0001\u0014\u0000Z\u0002\u0019~K\u0002\u0000\u0000I\t\u0004\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000K\u0005\u0001\u0001\u00024G*A5\r\n",
		"RemainingCacheRecords": 0,
		"DataPackets":           1,
		"ListPackets": map[string]interface{}{
			"packet_1": map[string]interface{}{
				"AD1":      0,
				"AD2":      0,
				"AD4":      0,
				"AD5":      12,
				"Altitude": 0,
				"BaseStationInfo": map[string]interface{}{
					"cellId":  38501913,
					"lac":     602,
					"mmc":     334,
					"mnc":     20,
					"rxLevel": 0,
				},
				"CameraStatus": map[string]interface{}{
					"camerasNumber": 4,
					"status":        "0",
				},
				"CurrentNetworkInfo": map[string]interface{}{
					"decriptorLen": 2,
					"descriptor":   "4G",
					"type":         "01",
					"version":      "01",
				},
				"Datetime":           "2024-09-19T23:55:22Z",
				"Direction":          0,
				"EventCode":          map[string]interface{}{"Code": 35, "Name": "Track By Time Interval"},
				"GsmSignalStrength":  10,
				"Hdop":               0,
				"IoPortStatus":       "10",
				"Latitude":           19.52101,
				"Longitude":          -99.211608,
				"Mileage":            0,
				"NumberOfSatellites": 0,
				"OutputPortStatus":   "0",
				"PositioningStatus":  false,
				"RunTime":            77384,
				"Speed":              0,
				"SystemFlag":         false,
			},
		},
	}
	assert.NoError(t, err, "La función devolvió un error inesperado")

	var resultMap map[string]interface{}
	err = json.Unmarshal([]byte(result), &resultMap)
	assert.NoError(t, err, "Error al parsear el JSON del resultado")

	assert.Equal(t, expectedData, resultMap, "El resultado no coincide con el esperado")
}
