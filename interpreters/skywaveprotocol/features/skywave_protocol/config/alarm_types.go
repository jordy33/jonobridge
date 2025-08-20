package config

var AdditionalAlarmTypesSecondProtocol = map[string]any{
	"01": CodeModel{Code: 1, Name: "Look left"},
	"02": CodeModel{Code: 2, Name: "Look right"},
	"03": CodeModel{Code: 3, Name: "Raise head"},
	"04": CodeModel{Code: 4, Name: "Lower head"},
	"05": CodeModel{Code: 5, Name: "Drowsiness"},
	"06": CodeModel{Code: 6, Name: "Yawning"},
	"07": CodeModel{Code: 7, Name: "Calling"},
	"08": CodeModel{Code: 8, Name: "Smoking"},
	"09": CodeModel{Code: 9, Name: "Drinking"},
	"0a": CodeModel{Code: 10, Name: "Driver absence"},
	"0b": CodeModel{Code: 11, Name: "Camera occlusion"},
	"80": CodeModel{Code: 128, Name: "Forward collision"},
	"81": CodeModel{Code: 129, Name: "Distance detection"},
	"82": CodeModel{Code: 130, Name: "Left lane departure"},
	"83": CodeModel{Code: 131, Name: "Right lane departure"},
	"84": CodeModel{Code: 132, Name: "Front vehicle started"},
}

var AdditionalAlarmTypeFirstProtocol = map[string]any{
	"01": CodeModel{Code: 1, Name: "Close eyes"},
	"02": CodeModel{Code: 2, Name: "Yawning"},
	"03": CodeModel{Code: 3, Name: "Not defined"},
	"04": CodeModel{Code: 4, Name: "Lower head"},
	"05": CodeModel{Code: 5, Name: "Look left or right"},
	"06": CodeModel{Code: 6, Name: "Driver absence"},
	"07": CodeModel{Code: 7, Name: "Calling"},
	"08": CodeModel{Code: 8, Name: "Smoking"},
	"09": CodeModel{Code: 9, Name: "Camera occlusion"},
	"0a": CodeModel{Code: 10, Name: "Forward Collision Warning (FCW)"},
	"0b": CodeModel{Code: 11, Name: "Urban Forward Collision Warning (UFCW)"},
	"0c": CodeModel{Code: 12, Name: "Left Lane Departure Warning"},
	"0d": CodeModel{Code: 13, Name: "Right Lane Departure Warning"},
	"0e": CodeModel{Code: 14, Name: "Headway Monitoring and Warning (HMW)"},
	"0f": CodeModel{Code: 15, Name: "TTC 1"},
}

var AlarmTypesFatiqueDriving = map[string]any{
	"02": CodeModel{Code: 2, Name: "Moderate fatigue"},
	"03": CodeModel{Code: 3, Name: "Severe fatigue"},
	"04": CodeModel{Code: 4, Name: "Distraccion alert"},
	"05": CodeModel{Code: 5, Name: "Distraccion alert"},
	"06": CodeModel{Code: 6, Name: "Absence alert"},
	"07": CodeModel{Code: 7, Name: "Smoking alert"},
	"08": CodeModel{Code: 8, Name: "Yawning alert"},
}

var AlarmTypesBluetooth = map[string]any{
	"01": CodeModel{Code: 1, Name: "Low battery alert for the temperature and humidity sensor"},
	"02": CodeModel{Code: 2, Name: "High temperature alert for the temperature and humidity sensor"},
	"03": CodeModel{Code: 3, Name: "Low temperature alert for the temperature and humidity sensor"},
	"04": CodeModel{Code: 4, Name: "High humidity alert for the temperature and humidity sensor"},
	"05": CodeModel{Code: 5, Name: "Low humidity alert for the temperature and humidity sensor"},
	"06": CodeModel{Code: 6, Name: "Signal lost alert for the temperature and humidity sensor"},
	"07": CodeModel{Code: 7, Name: "Signal recovery alert for the temperature and humidity sensor"},
	"08": CodeModel{Code: 8, Name: "Low battery alert for the Bluetooth beacon"},
	"09": CodeModel{Code: 9, Name: "Bluetooth beacon lost alert"},
	"10": CodeModel{Code: 10, Name: "Bluetooth beacon found alert"},
}
