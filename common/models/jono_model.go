package models

import "time"

type JonoModel struct {
	IMEI        string                `json:"IMEI"`
	Message     *string               `json:"Message"`
	DataPackets int                   `json:"DataPackets"`
	ListPackets map[string]DataPacket `json:"ListPackets"`
}

type DataPacket struct {
	Altitude                     int                         `json:"Altitude"`
	Datetime                     time.Time                   `json:"Datetime"`
	EventCode                    EventCode                   `json:"EventCode"`
	Latitude                     float64                     `json:"Latitude"`
	Longitude                    float64                     `json:"Longitude"`
	Speed                        int                         `json:"Speed"`
	RunTime                      int                         `json:"RunTime"`
	FuelPercentage               int                         `json:"FuelPercentage"`
	Direction                    int                         `json:"Direction"`
	HDOP                         float64                     `json:"HDOP"`
	Mileage                      int                         `json:"Mileage"`
	PositioningStatus            string                      `json:"PositioningStatus"`
	NumberOfSatellites           int                         `json:"NumberOfSatellites"`
	GSMSignalStrength            *int                        `json:"GSMSignalStrength"`
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

type EventCode struct {
	Code int    `json:"Code"`
	Name string `json:"Name"`
}

type BaseStationInfo struct {
	MCC    *string `json:"MCC"`
	MNC    *string `json:"MNC"`
	LAC    *string `json:"LAC"`
	CellID *string `json:"CellID"`
}

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

type TemperatureSensor struct {
	SensorNumber *string `json:"SensorNumber"`
	Value        *string `json:"Value"`
}

type CameraStatus struct {
	CameraNumber *string `json:"CameraNumber"`
	Status       *string `json:"Status"`
}

type CurrentNetworkInfo struct {
	Version    *string `json:"Version"`
	Type       *string `json:"Type"`
	Descriptor *string `json:"Descriptor"`
}

type FatigueDrivingInformation struct {
	Version    *string `json:"Version"`
	Type       *string `json:"Type"`
	Descriptor *string `json:"Descriptor"`
}

type AdditionalAlertInfoADASDMS struct {
	AlarmProtocol *string `json:"AlarmProtocol"`
	AlarmType     *string `json:"AlarmType"`
	PhotoName     *string `json:"PhotoName"`
}

type BluetoothBeacon struct {
	Version        *string `json:"Version"`
	DeviceName     *string `json:"DeviceName"`
	MAC            *string `json:"MAC"`
	BatteryPower   *string `json:"BatteryPower"`
	SignalStrength *string `json:"SignalStrength"`
}

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
