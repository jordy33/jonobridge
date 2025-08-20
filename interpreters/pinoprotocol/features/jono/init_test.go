package jono_test

import (
	"encoding/json"
	"pinoprotocol/features/jono"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testJSON = `{
	"StartSignal": "$$",
	"Identifier": "a",
	"DataLength": "139",
	"IMEI": "866811062546604",
	"CommandType": "CCE",
	"Rest": "\u0000\u0000\u0000\u0000\u0001\u0000i\u0000\u0017\u0000\u0005\u0005\u0000\u0006\u0000\u0007\b\u0014\u0000\u0015\u0002\t\b\u0000\u0000\t\u0000\u0000\n\u0000\u0000\u000b\u0000\u0000\u0016\b\u0000\u0017\u0006\u0000\u0019\u0001\u0000\u001a\ufffd\u0004@#\u0000\u0006\u0002\ufffd\ufffd)\u0001\u0003\ufffd\u0026\u0016\ufffd\u0004㿷.\f\u0000\u0000\u0000\u0000\r\b\ufffd\u0003\u0000\u001c\u0001 \u0000\u0000\u0003\u000e\fN\u0001\u0014\u0000Z\u0002\u0002~K\u0002\u0000\u0000I\t\u0004\u0003\u0000\u0000\u0000\u0000\u0000\u0000\u0000K\u0005\u0001\u0001\u00024G*F7\r\n",
	"Message": "$$a139,866811062546604,CCE,\u0000\u0000\u0000\u0000\u0001\u0000i\u0000\u0017\u0000\u0005\u0005\u0000\u0006\u0000\u0007\b\u0014\u0000\u0015\u0002\t\b\u0000\u0000\t\u0000\u0000\n\u0000\u0000\u000b\u0000\u0000\u0016\b\u0000\u0017\u0006\u0000\u0019\u0001\u0000\u001a\ufffd\u0004@#\u0000\u0006\u0002\ufffd\ufffd)\u0001\u0003\ufffd\u0026\u0016\ufffd\u0004㿷.\f\u0000\u0000\u0000\u0000\r\b\ufffd\u0003\u0000\u001c\u0001 \u0000\u0000\u0003\u000e\fN\u0001\u0014\u0000Z\u0002\u0002~K\u0002\u0000\u0000I\t\u0004\u0003\u0000\u0000\u0000\u0000\u0000\u0000\u0000K\u0005\u0001\u0001\u00024G*F7\r\n",
	"RemainingCacheRecords": 0,
	"DataPackets": 1,
	"ListPackets": {
		"packet_1": {
			"Altitude": 0,
			"Datetime": "2024-09-19T23:55:22Z",
			"EventCode": {"Code": 35, "Name": "Track By Time Interval"},
			"GsmSignalStrength": 10,
			"Latitude": 19.52101,
			"Longitude": -99.211608,
			"Speed": 0
		},
		"packet_2": {
			"Altitude": 0,
			"Datetime": "2024-09-19T23:55:22Z",
			"EventCode": {"Code": 35, "Name": "Track By Time Interval"},
			"GsmSignalStrength": 10,
			"Latitude": 19.52101,
			"Longitude": -99.211608,
			"Speed": 0
		}
	}
}`

// Test function for Initialize
func TestInitialize(t *testing.T) {
	// Setup: define the expected output after processing testJSON
	expectedOutput := `{
		"IMEI": "866811062546604",
		"DataPackets": 1,
		"ListPackets": {
			"packet_1": {
				"Altitude": 0,
				"Datetime": "2024-09-19T23:55:22Z",
				"EventCode": {"Code": 35, "Name": "Track By Time Interval"},
				"Latitude": 19.52101,
				"Longitude": -99.211608,
				"Speed": 0,
				"Extras": {"GsmSignalStrength": 10}
			},
			"packet_2": {
				"Altitude": 0,
				"Datetime": "2024-09-19T23:55:22Z",
				"EventCode": {"Code": 35, "Name": "Track By Time Interval"},
				"Latitude": 19.52101,
				"Longitude": -99.211608,
				"Speed": 0,
				"Extras": {"GsmSignalStrength": 10}
			}
		}
	}`

	// Call the Initialize function with testJSON data
	output, err := jono.Initialize(testJSON)

	// Deserialize expected and actual output JSON
	var expectedMap, outputMap map[string]interface{}
	json.Unmarshal([]byte(expectedOutput), &expectedMap)
	json.Unmarshal([]byte(output), &outputMap)

	// Remove the Message field to avoid comparison on it
	delete(expectedMap, "Message")
	delete(outputMap, "Message")

	// Assert no error occurred
	assert.NoError(t, err, "Expected no error from Initialize function")

	// Assert the maps match, excluding Message field
	assert.Equal(t, expectedMap, outputMap, "Expected output does not match actual output")
}

// JSON de ejemplo para el modelo AAA
const jsonAAA = `{
	"StartSignal": "$$",
	"Identifier": "f",
	"DataLength": "167",
	"IMEI": "864507035846483",
	"CommandType": "AAA",
	"Rest": "1,18.950273,-97.922888,12546542156,V,0,13,0,69,0.0,2217,358868041,192062311,334|3|7663|00AA7FAB,0000,0001|0000|0000|01A5|0514,,,3,,,108,106*C6",
	"Message": "$$f167,864507035846483,AAA,1,18.950273,-97.922888,12546542156,V,0,13,0,69,0.0,2217,358868041,192062311,334|3|7663|00AA7FAB,0000,0001|0000|0000|01A5|0514,,,3,,,108,106*C6",
	"EventCode": {"Code": 1, "Name": "Input 1 Active"},
	"Latitude": 18.950273513793945,
	"Longitude": -97.92288970947266,
	"Datetime": "12546542156",
	"PositioningStatus": "V",
	"NumberOfSatellites": 0,
	"GsmSignalStrength": 13,
	"Speed": 0,
	"Direction": 69,
	"Hdop": 0,
	"Altitude": 2217,
	"Mileage": 358868041,
	"RunTime": 192062311,
	"BaseStationInfo": {"cellId": "00AA7FAB", "lac": "7663", "mmc": "334", "mnc": "3", "rxLevel": -1},
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

func TestInitializeWithAAA(t *testing.T) {
	expectedOutput := `{
		"IMEI": "864507035846483",
		"Message": "$$f167,864507035846483,AAA,1,18.950273,-97.922888,12546542156,V,0,13,0,69,0.0,2217,358868041,192062311,334|3|7663|00AA7FAB,0000,0001|0000|0000|01A5|0514,,,3,,,108,106*C6",
		"DataPackets": 1,
		"ListPackets": {
			"packet_1": {
				"Altitude": 2217,
				"Datetime": "12546542156",
				"EventCode": {
					"Code": 1,
					"Name": "Input 1 Active"
				},
				"Latitude": 18.950273513793945,
				"Longitude": -97.92288970947266,
				"Speed": 0,
				"Extras": {
					"StartSignal": "$$",
					"Identifier": "f",
					"DataLength": "167",
					"CommandType": "AAA",
					"Rest": "1,18.950273,-97.922888,12546542156,V,0,13,0,69,0.0,2217,358868041,192062311,334|3|7663|00AA7FAB,0000,0001|0000|0000|01A5|0514,,,3,,,108,106*C6",
					"PositioningStatus": "V",
					"NumberOfSatellites": 0,
					"GsmSignalStrength": 13,
					"Direction": 69,
					"Hdop": 0,
					"Mileage": 358868041,
					"RunTime": 192062311,
					"BaseStationInfo": {
						"cellId": "00AA7FAB",
						"lac": "7663",
						"mmc": "334",
						"mnc": "3",
						"rxLevel": -1
					},
					"IoPortStatus": "0000",
					"AnalogInputs": {
						"AD1": "0001",
						"AD2": "0000",
						"AD3": "0000",
						"AD4": "01A5",
						"AD5": "0514"
					},
					"AssistedEventInfo": "",
					"CustomizedData": "",
					"ProtocolVersion": 3,
					"FuelPercentage": "",
					"TemperatureSensor": "",
					"MaxAcceleration": 108,
					"MaxDesceleration": 106,
					"Checksum": "C6"
				}
			}
		}
	}`

	// Call the Initialize function
	output, err := jono.Initialize(jsonAAA)

	// Unmarshal expected and actual output
	var expectedMap, outputMap map[string]interface{}
	json.Unmarshal([]byte(expectedOutput), &expectedMap)
	json.Unmarshal([]byte(output), &outputMap)

	// Remove the Message field if it is not critical
	delete(expectedMap, "Message")
	delete(outputMap, "Message")

	// Assert no error occurred
	assert.NoError(t, err, "Expected no error from Initialize function")

	// Compare the maps without Message
	assert.Equal(t, expectedMap, outputMap, "Expected output does not match actual output")
}

const jsonCCE = `{
	"StartSignal": "$$",
	"Identifier": "a",
	"DataLength": "139",
	"IMEI": "866811062546604",
	"CommandType": "CCE",
	"Rest": "\u0000\u0000\u0000\u0000\u0001\u0000i\u0000\u0017\u0000\u0005\u0005\u0000\u0006\u0000\u0007\b\u0014\u0000\u0015\u0002\t\b\u0000\u0000\t\u0000\u0000\n\u0000\u0000\u000b\u0000\u0000\u0016\b\u0000\u0017\u0006\u0000\u0019\u0001\u0000\u001a\ufffd\u0004@#\u0000\u0006\u0002\ufffd\ufffd)\u0001\u0003\ufffd\u0026\u0016\ufffd\u0004㿷.\f\u0000\u0000\u0000\u0000\r\b\ufffd\u0003\u0000\u001c\u0001 \u0000\u0000\u0003\u000e\fN\u0001\u0014\u0000Z\u0002\u0002~K\u0002\u0000\u0000I\t\u0004\u0003\u0000\u0000\u0000\u0000\u0000\u0000\u0000K\u0005\u0001\u0001\u00024G*F7\r\n",
	"Message": "$$a139,866811062546604,CCE,\u0000\u0000\u0000\u0000\u0001\u0000i\u0000\u0017\u0000\u0005\u0005\u0000\u0006\u0000\u0007\b\u0014\u0000\u0015\u0002\t\b\u0000\u0000\t\u0000\u0000\n\u0000\u0000\u000b\u0000\u0000\u0016\b\u0000\u0017\u0006\u0000\u0019\u0001\u0000\u001a\ufffd\u0004@#\u0000\u0006\u0002\ufffd\ufffd)\u0001\u0003\ufffd\u0026\u0016\ufffd\u0004㿷.\f\u0000\u0000\u0000\u0000\r\b\ufffd\u0003\u0000\u001c\u0001 \u0000\u0000\u0003\u000e\fN\u0001\u0014\u0000Z\u0002\u0002~K\u0002\u0000\u0000I\t\u0004\u0003\u0000\u0000\u0000\u0000\u0000\u0000\u0000K\u0005\u0001\u0001\u00024G*F7\r\n",
	"RemainingCacheRecords": 0,
	"DataPackets": 1,
	"ListPackets": {
		"packet_1": {
			"AD1": 0,
			"AD2": 0,
			"AD4": 0,
			"AD5": 12,
			"Altitude": 0,
			"BaseStationInfo": {"cellId": 38501913, "lac": 602, "mmc": 334, "mnc": 20, "rxLevel": 0},
			"CameraStatus": {"camerasNumber": 4, "status": "11"},
			"CurrentNetworkInfo": {"decriptorLen": 2, "descriptor": "4G", "type": "01", "version": "01"},
			"Datetime": "2024-11-01T16:35:47Z",
			"Direction": 0,
			"EventCode": {"Code": 35, "Name": "Track By Time Interval"},
			"GsmSignalStrength": 8,
			"Hdop": 0,
			"IoPortStatus": "10",
			"Latitude": 19.52101,
			"Longitude": -99.211608,
			"Mileage": 0,
			"NumberOfSatellites": 0,
			"OutputPortStatus": "0",
			"PositioningStatus": false,
			"RunTime": 250632,
			"Speed": 0,
			"SystemFlag": false
		}
	}
}`

func TestInitializeWithCCE(t *testing.T) {
	expectedOutput := `{
		"IMEI": "866811062546604",
		"DataPackets": 1,
		"ListPackets": {
			"packet_1": {
				"Altitude": 0,
				"Datetime": "2024-11-01T16:35:47Z",
				"EventCode": {"Code": 35, "Name": "Track By Time Interval"},
				"Latitude": 19.52101,
				"Longitude": -99.211608,
				"Speed": 0,
				"Extras": {
					"AD1": 0,
					"AD2": 0,
					"AD4": 0,
					"AD5": 12,
					"BaseStationInfo": {"cellId": 38501913, "lac": 602, "mmc": 334, "mnc": 20, "rxLevel": 0},
					"CameraStatus": {"camerasNumber": 4, "status": "11"},
					"CurrentNetworkInfo": {"decriptorLen": 2, "descriptor": "4G", "type": "01", "version": "01"},
					"Direction": 0,
					"GsmSignalStrength": 8,
					"Hdop": 0,
					"IoPortStatus": "10",
					"Mileage": 0,
					"NumberOfSatellites": 0,
					"OutputPortStatus": "0",
					"PositioningStatus": false,
					"RunTime": 250632,
					"SystemFlag": false
				}
			}
		}
	}`

	// Ejecutar la función que se está probando
	output, err := jono.Initialize(jsonCCE)
	assert.NoError(t, err, "Expected no error from Initialize function")

	// Parsear el JSON esperado y el obtenido como mapas
	var expectedMap, outputMap map[string]interface{}
	json.Unmarshal([]byte(expectedOutput), &expectedMap)
	json.Unmarshal([]byte(output), &outputMap)

	// Eliminar el campo Message de ambos mapas
	delete(expectedMap, "Message")
	delete(outputMap, "Message")

	// Comparar los mapas sin el campo Message
	assert.Equal(t, expectedMap, outputMap, "Expected output does not match actual output, excluding Message field")
}

const ruptela = `{
	"Altitude":6067.2,
	"Datetime":"2011-10-17T22:41:48",
	"Direction":"0.1",
	"EventCode": {
		"Code": 773,
		"Name": "Unknown"
	},
	"Hdop":"0.7",
	"IMEI":"9223372036854775807",
	"Latitude":34.3980544,
	"Longitude":-70.4057062,
	"NumberOfSatellites":"0",
	"Speed":11
	}`

func TestInitializeWithRuptela(t *testing.T) {
	expectedOutput := `{
		"IMEI": "9223372036854775807",
		"DataPackets": 1,
		"ListPackets": {
			"packet_1": {
				"Altitude": 6067.2,
				"Datetime": "2011-10-17T22:41:48",
				"EventCode": {"Code": 773, "Name": "Unknown"},
				"Latitude": 34.3980544,
				"Longitude": -70.4057062,
				"Speed": 11,
				"Extras": {
					"Hdop": "0.7",
					"NumberOfSatellites": "0",
					"Direction":"0.1"
				}
			}
		}
	}`

	output, err := jono.Initialize(ruptela)
	assert.NoError(t, err, "Expected no error from Initialize function")

	var expectedMap, outputMap map[string]interface{}
	json.Unmarshal([]byte(expectedOutput), &expectedMap)
	json.Unmarshal([]byte(output), &outputMap)

	delete(expectedMap, "Message")
	delete(outputMap, "Message")

	assert.Equal(t, expectedMap, outputMap, "Expected output does not match actual output, excluding Message field")
}
