package jono_test

import (
	"encoding/json"
	"huabaoprotocol/features/jono/usecases"
	"testing"
)

// 📌 TestGetDataJono verifica que GetDataJono devuelva JSON válido sin errores
func TestGetDataJono(t *testing.T) {
	// 📌 JSON de prueba con valores mínimos requeridos
	inputJSON := `{
		"IMEI": "123456789012345",
		"Message": "Test Message",
		"DataPackets": 1,
		"ListPackets": {
			"packet_1": {
				"Altitude": 100,
				"Datetime": "2024-02-07T12:00:00Z",
				"EventCode": 35,
				"Latitude": 19.4326,
				"Longitude": -99.1332,
				"Speed": 50,
				"PositioningStatus": "A",
				"IoPortStatus": "0000",
				"AnalogInputs": {
					"AD1": "100",
					"AD2": "200"
				}
			}
		}
	}`

	// 📌 Ejecutar la función
	output, err := usecases.GetDataJono(inputJSON)

	// 📌 Verificar que no haya errores
	if err != nil {
		t.Fatalf("GetDataJono returned an error: %v", err)
	}

	// 📌 Verificar que la salida sea un JSON válido
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// 📌 Verificar que los campos principales existen
	requiredKeys := []string{"IMEI", "Message", "DataPackets", "ListPackets"}
	for _, key := range requiredKeys {
		if _, exists := result[key]; !exists {
			t.Errorf("Missing key in output JSON: %s", key)
		}
	}

	// 📌 Verificar que "ListPackets" contiene los paquetes esperados
	if packets, ok := result["ListPackets"].(map[string]interface{}); ok {
		if len(packets) != 1 {
			t.Errorf("Expected 1 packet, got %d", len(packets))
		}

		// 📌 Verificar que el paquete tiene los valores esperados
		if packet, ok := packets["packet_1"].(map[string]interface{}); ok {
			// Check for required fields in the packet
			requiredPacketKeys := []string{
				"Altitude", "Datetime", "EventCode", "Latitude", "Longitude",
				"Speed", "PositioningStatus", "IoPortStatus", "AnalogInputs",
				"Direction", "NumberOfSatellites", "GSMSignalStrength", "HDOP",
				"BaseStationInfo", "AdditionalAlertInfoADASDMS", "BluetoothBeaconA",
				"SystemFlag", "TemperatureSensor",
			}

			for _, key := range requiredPacketKeys {
				if _, exists := packet[key]; !exists {
					t.Errorf("Missing key in packet_1: %s", key)
				}
			}

			// Verify EventCode is an object with Code and Name
			if eventCode, ok := packet["EventCode"].(map[string]interface{}); ok {
				if _, exists := eventCode["Code"]; !exists {
					t.Errorf("EventCode is missing 'Code' field")
				}
				if _, exists := eventCode["Name"]; !exists {
					t.Errorf("EventCode is missing 'Name' field")
				}
			} else {
				t.Errorf("EventCode is not a map[string]interface{}")
			}

			// Verify IoPortStatus is an object
			if ioStatus, ok := packet["IoPortStatus"].(map[string]interface{}); ok {
				if len(ioStatus) == 0 {
					t.Errorf("IoPortStatus map is empty")
				}
			} else {
				t.Errorf("IoPortStatus is not a map[string]interface{}")
			}
		} else {
			t.Errorf("packet_1 is missing or not a valid object")
		}
	} else {
		t.Errorf("ListPackets is missing or not a map")
	}
}

// 📌 TestGetDataJono_MissingOptionalFields verifica que los campos opcionales sean `null`
func TestGetDataJono_MissingOptionalFields(t *testing.T) {
	// 📌 JSON de prueba SIN los valores opcionales
	inputJSON := `{
		"IMEI": "123456789012345",
		"Message": "Test Message",
		"DataPackets": 1,
		"ListPackets": {
			"packet_1": {
				"Altitude": 100,
				"Datetime": "2024-02-07T12:00:00Z",
				"EventCode": 35,
				"Latitude": 19.4326,
				"Longitude": -99.1332,
				"Speed": 50
			}
		}
	}`

	// 📌 Ejecutar la función
	output, err := usecases.GetDataJono(inputJSON)

	// 📌 Verificar que no haya errores
	if err != nil {
		t.Fatalf("GetDataJono returned an error: %v", err)
	}

	// 📌 Verificar que la salida sea un JSON válido
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// 📌 Verificar que "ListPackets" contiene los paquetes esperados
	if packets, ok := result["ListPackets"].(map[string]interface{}); ok {
		if packet, ok := packets["packet_1"].(map[string]interface{}); ok {
			// All optional fields should be present but can be null
			requiredFields := []string{
				"PositioningStatus", "AnalogInputs", "IoPortStatus",
				"BaseStationInfo", "SystemFlag", "OutputPortStatus",
				"InputPortStatus", "TemperatureSensor", "AdditionalAlertInfoADASDMS",
			}

			for _, key := range requiredFields {
				if _, exists := packet[key]; !exists {
					t.Errorf("Expected %s to be present (even if null), but it's missing", key)
				}
			}
		} else {
			t.Errorf("packet_1 is missing or not a valid object")
		}
	} else {
		t.Errorf("ListPackets is missing or not a map")
	}
}
