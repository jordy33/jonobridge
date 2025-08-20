package models

import (
	"meitrackprotocol/features/meitrack_protocol/config"
	"meitrackprotocol/features/meitrack_protocol/helpers"
)

type CCEModel struct {
	GeneralModel
	RemainingCacheRecords int
	DataPackets           int
	ListPackets           map[string]any
}

type IDModel struct {
	Name       string
	Conversion func(string) any
}

var IDOneByte = map[string]IDModel{
	"01":   {"EventCode", func(hexString string) interface{} { return helpers.FetchInEventCodes(hexString, config.EventCodes) }},
	"05":   {"PositioningStatus", helpers.BooleanValue},
	"06":   {"NumberOfSatellites", helpers.HexToInt},
	"07":   {"GsmSignalStrength", helpers.HexToInt},
	"14":   {"OutputPortStatus", helpers.HexToBinary},
	"15":   {"IoPortStatus", helpers.HexToBinary},
	"1b":   {"GeoFenceNumber", helpers.HexToInt},
	"27":   {"TemperatureSensor", helpers.HexToInt},
	"93":   {"ClutchSwitch", helpers.HexToInt},
	"94":   {"TachographPerformance", helpers.HexToInt},
	"95":   {"ParkingBrakeSwitch", helpers.HexToInt},
	"96":   {"CruiseControl", helpers.HexToInt},
	"97":   {"AcceleratorPedalPosition", helpers.HexToInt},
	"9d":   {"FuelPercentage", helpers.HexToInt},
	"9e":   {"ActualEngineTorque", helpers.HexToInt},
	"a1":   {"ActualEngineTorqueLoadAtCurrentSpeed", helpers.HexToInt},
	"fe69": {"BateryPercentage", helpers.HexToInt},
}

var IDTwoBytes = map[string]IDModel{
	"08": {"Speed", helpers.HexToLittleEndianDecimal},
	"09": {"Direction", helpers.HexToLittleEndianDecimal},
	"0a": {"HDOP", helpers.HexToLittleEndianDecimal}, // Changed from Hdop to HDOP
	"0b": {"Altitude", helpers.HexToLittleEndianDecimal},
	"16": {"AD1", helpers.DivideByHundred},
	"17": {"AD2", helpers.DivideByHundred},
	"18": {"AD3", helpers.DivideByHundred},
	"19": {"AD4", helpers.DivideByHundred},
	"1a": {"AD5", helpers.DivideByHundred},
	"29": {"FuelPercentage", helpers.Percentage},
	"40": {"EventCode", func(hexString string) interface{} { return helpers.FetchInEventCodes(hexString, config.EventCodes) }},
	"41": {"Unknown", helpers.HexToLittleEndianDecimal},
	"91": {"VehicleSpeedBasedOnTachograph", helpers.HexToLittleEndianDecimal},
	"92": {"VehicleSpeedBasedOnWheel", helpers.HexToLittleEndianDecimal},
	"99": {"EngineSpeed", helpers.HexToLittleEndianDecimal},
	"9c": {"EngineCoolantTemperature", helpers.HexToLittleEndianDecimal},
	"9f": {"AmbientAirTemperature", helpers.HexToLittleEndianDecimal},
}

var IDFourBytes = map[string]IDModel{
	"02": {"Latitude", helpers.LatLngValue},
	"03": {"Longitude", helpers.LatLngValue},
	"04": {"Datetime", helpers.DateAndTime},
	"0c": {"Mileage", helpers.HexToLittleEndianDecimal},
	"0d": {"RunTime", helpers.HexToLittleEndianDecimal},
	"1c": {"SystemFlag", helpers.BooleanValue},
	"25": {"RfidId", helpers.HexToLittleEndianDecimal},
	"98": {"TotalFuelConsumption", helpers.HexToLittleEndianDecimal},
	"9a": {"TotalEngineRunTime", helpers.DivideByTen},
	"9b": {"HighResolutionVehicleDistance", helpers.HexToLittleEndianDecimal},
	"a0": {"HighResolutionTotalFuelConsumption", helpers.DivideByThousand},
	"a2": {"FuelConsumptionRate", helpers.DivideByHundred},
	"a3": {"AxleWeight", helpers.DivideByTen},
	"a4": {"ServiceDistance", helpers.HexToLittleEndianDecimal},
	"a5": {"InstantaneousFuelConsumption", helpers.DivideByThousand},
}

var IDUndefinedBytes = map[string]IDModel{
	"0e": {"BaseStationInfo", helpers.BaseStationInfo},
	"28": {"PictureName", helpers.PictureName},
	"2a": {"TemperatureSensor1", helpers.TemperatureSensor},
	"2b": {"TemperatureSensor2", helpers.TemperatureSensor},
	"2c": {"TemperatureSensor3", helpers.TemperatureSensor},
	"2d": {"TemperatureSensor4", helpers.TemperatureSensor},
	"2e": {"TemperatureSensor5", helpers.TemperatureSensor},
	"2f": {"TemperatureSensor6", helpers.TemperatureSensor},
	"30": {"TemperatureSensor7", helpers.TemperatureSensor},
	"31": {"TemperatureSensor8", helpers.TemperatureSensor},
	"39": {"MagneticCardReader", helpers.HexToInt},
	"49": {"CameraStatus", helpers.CameraStatus},
	"4b": {"CurrentNetworkInfo", helpers.NetworkInformation},
	"fe2D": {"FatigueDrivingInformation", func(hexString string) interface{} {
		return helpers.FatigueDrivingInfo(hexString, config.AlarmTypesFatiqueDriving)
	}},
	"fe31": {"AdditionalAlertInfoADASDMS", func(hexString string) interface{} {
		return helpers.AdditionalAlertInfo(hexString, config.AdditionalAlarmTypeFirstProtocol, config.AdditionalAlarmTypesSecondProtocol)
	}},
	"fe70": {"AdditionalInfoBluetoothDevice", func(hexString string) interface{} {
		return helpers.FetchInEventCodes(hexString, config.AlarmTypesBluetooth)
	}},
	"fe71": {"BluetoothBeaconA", helpers.BluetoothBeacon},
	"fe72": {"BluetoothBeaconB", helpers.BluetoothBeacon},
	"fe73": {"TemperatureAndHumiditySensor", helpers.TemperatureAndHumidity},
}
