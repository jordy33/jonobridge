package models

type CCCModel struct {
	GeneralModel
	ProtocolVersion         string
	PacketLength            int
	NumberOrRemainingCaches int
	EventCode               any
	Latitude                float64
	Longitude               float64
	Datetime                string
	PositioningStatus       string
	NumberOfSatellites      int
	Speed                   int
	Direction               int
	HorizontalPositioning   int
	Altitude                int
	Mileage                 int
	RunTime                 int
	BaseStationInfo         any
	IoPortStatus            string
	AnalogsInput            any
	GeoFenceNumber          int
}
