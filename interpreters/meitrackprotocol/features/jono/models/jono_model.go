package models

import (
	"encoding/json"
	"fmt"
)

// 📌 BaseStationInfo contiene información de la estación base
type BaseStationInfo struct {
	MCC    *string `json:"MCC"`
	MNC    *string `json:"MNC"`
	LAC    *string `json:"LAC"`
	CellID *string `json:"CellID"`
}

// 📌 AnalogInputs contiene las entradas analógicas
type AnalogInputs struct {
	AD1  *string `json:"AD1"`
	AD2  *string `json:"AD2"`
	AD3  *string `json:"AD3"`
	AD4  *string `json:"AD4"`
	AD5  *string `json:"AD5"`
	AD6  *string `json:"AD6"`
	AD7  *string `json:"AD7"`
	AD8  *string `json:"AD8"`
	AD9  *string `json:"AD9"`
	AD10 *string `json:"AD10"`
}

// 📌 EventCode contiene el código del evento
type EventCode struct {
	Code int    `json:"Code"`
	Name string `json:"Name"`
}

// 📌 OutputPortStatus contiene el estado de los puertos de salida
type OutputPortStatus struct {
	Output1 *string `json:"Output1"`
	Output2 *string `json:"Output2"`
	Output3 *string `json:"Output3"`
	Output4 *string `json:"Output4"`
	Output5 *string `json:"Output5"`
	Output6 *string `json:"Output6"`
	Output7 *string `json:"Output7"`
	Output8 *string `json:"Output8"`
}

// 📌 InputPortStatus contiene el estado de los puertos de entrada
type InputPortStatus struct {
	Input1 *string `json:"Input1"`
	Input2 *string `json:"Input2"`
	Input3 *string `json:"Input3"`
	Input4 *string `json:"Input4"`
	Input5 *string `json:"Input5"`
	Input6 *string `json:"Input6"`
	Input7 *string `json:"Input7"`
	Input8 *string `json:"Input8"`
}

// 📌 SystemFlag contiene banderas del sistema
type SystemFlag struct {
	EEP2                *string `json:"EEP2"`
	ACC                 *string `json:"ACC"`
	AntiTheft           *string `json:"AntiTheft"`
	VibrationFlag       *string `json:"VibrationFlag"`
	MovingFlag          *string `json:"MovingFlag"`
	ExternalPowerSupply *string `json:"ExternalPowerSupply"`
	Charging            *string `json:"Charging"`
	SleepMode           *string `json:"SleepMode"`
	FMS                 *string `json:"FMS"`
	FMSFunction         *string `json:"FMSFunction"`
	SystemFlagExtras    *string `json:"SystemFlagExtras"`
}

// 📌 TemperatureSensor representa un sensor de temperatura
type TemperatureSensor struct {
	SensorNumber *string `json:"SensorNumber"`
	Value        *string `json:"Value"`
}

// 📌 CameraStatus representa el estado de una cámara
type CameraStatus struct {
	CameraNumber *string `json:"CameraNumber"`
	Status       *string `json:"Status"`
}

// 📌 CurrentNetworkInfo representa información de red
type CurrentNetworkInfo struct {
	Version    *string `json:"Version"`
	Type       *string `json:"Type"`
	Descriptor *string `json:"Descriptor"`
}

// 📌 FatigueDrivingInformation representa información sobre fatiga del conductor
type FatigueDrivingInformation struct {
	Version    *string `json:"Version"`
	Type       *string `json:"Type"`
	Descriptor *string `json:"Descriptor"`
}

// 📌 AdditionalAlertInfoADASDMS representa alertas adicionales
type AdditionalAlertInfoADASDMS struct {
	AlarmProtocol *string `json:"AlarmProtocol"`
	AlarmType     *string `json:"AlarmType"`
	PhotoName     *string `json:"PhotoName"`
}

// 📌 BluetoothBeacon representa información de beacons Bluetooth
type BluetoothBeacon struct {
	Version        *string `json:"Version"`
	DeviceName     *string `json:"DeviceName"`
	MAC            *string `json:"MAC"`
	BatteryPower   *string `json:"BatteryPower"`
	SignalStrength *string `json:"SignalStrength"`
}

// 📌 TemperatureAndHumidity representa sensores de temperatura y humedad
type TemperatureAndHumidity struct {
	DeviceName           *string `json:"DeviceName"`
	MAC                  *string `json:"MAC"`
	BatteryPower         *string `json:"BatteryPower"`
	Temperature          *string `json:"Temperature"`
	Humidity             *string `json:"Humidity"`
	AlertHighTemperature *string `json:"AlertHighTemperature"`
	AlertLowTemperature  *string `json:"AlertLowTemperature"`
	AlertHighHumidity    *string `json:"AlertHighHumidity"`
	AlertLowHumidity     *string `json:"AlertLowHumidity"`
}

// 📌 IoPortsStatus contiene el estado de los puertos de entrada/salida con valores predeterminados en 0
type IoPortsStatus struct {
	Port1 int `json:"Port1"`
	Port2 int `json:"Port2"`
	Port3 int `json:"Port3"`
	Port4 int `json:"Port4"`
	Port5 int `json:"Port5"`
	Port6 int `json:"Port6"`
	Port7 int `json:"Port7"`
	Port8 int `json:"Port8"`
}

// 📌 Packet contiene toda la información de un paquete
type Packet struct {
	Altitude                     int                         `json:"Altitude"`
	Datetime                     *string                     `json:"Datetime"`
	EventCode                    EventCode                   `json:"EventCode"`
	Latitude                     *float64                    `json:"Latitude"`
	Longitude                    *float64                    `json:"Longitude"`
	Speed                        *int                        `json:"Speed"`
	RunTime                      *int                        `json:"RunTime"`
	Direction                    *int                        `json:"Direction"`
	HDOP                         *float64                    `json:"HDOP"`
	Mileage                      *int                        `json:"Mileage"`
	PositioningStatus            *string                     `json:"PositioningStatus"`
	NumberOfSatellites           int                         `json:"NumberOfSatellites"`
	GSMSignalStrength            *int                        `json:"GSMSignalStrength"` // Added GSM signal strength field
	AnalogInputs                 *AnalogInputs               `json:"AnalogInputs"`
	IoPortStatus                 *IoPortsStatus              `json:"IoPortStatus"`
	BaseStationInfo              *BaseStationInfo            `json:"BaseStationInfo"`
	OutputPortStatus             *OutputPortStatus           `json:"OutputPortStatus"`
	InputPortStatus              *InputPortStatus            `json:"InputPortStatus"`
	SystemFlag                   *SystemFlag                 `json:"SystemFlag"`
	TemperatureSensor            *TemperatureSensor          `json:"TemperatureSensor"`
	CameraStatus                 *CameraStatus               `json:"CameraStatus"`
	CurrentNetworkInfo           *CurrentNetworkInfo         `json:"CurrentNetworkInfo"`
	FatigueDrivingInformation    *FatigueDrivingInformation  `json:"FatigueDrivingInformation"`
	AdditionalAlertInfoADASDMS   *AdditionalAlertInfoADASDMS `json:"AdditionalAlertInfoADASDMS"`
	BluetoothBeaconA             *BluetoothBeacon            `json:"BluetoothBeaconA"`
	BluetoothBeaconB             *BluetoothBeacon            `json:"BluetoothBeaconB"`
	TemperatureAndHumiditySensor *TemperatureAndHumidity     `json:"TemperatureAndHumiditySensor"`
}

// 📌 ParsedModel representa el modelo final con paquetes
type ParsedModel struct {
	IMEI        *string           `json:"IMEI"`
	Message     *string           `json:"Message"`
	DataPackets *int              `json:"DataPackets"`
	ListPackets map[string]Packet `json:"ListPackets"`
}

// 📌 Método para convertir `ParsedModel` a JSON normal
func (p *ParsedModel) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// 📌 Método para convertir `ParsedModel` a JSON indentado (legible)
func (p *ParsedModel) ToPrettyJSON() (string, error) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal pretty JSON: %w", err)
	}
	return string(data), nil
}
