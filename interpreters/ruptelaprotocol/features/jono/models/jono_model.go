package models

import (
	"encoding/json"
)

type EventCode struct {
	Code int    `json:"Code"`
	Name string `json:"Name"`
}

type Packet struct {
	Altitude  int                    `json:"Altitude"`
	Datetime  string                 `json:"Datetime"`
	EventCode EventCode              `json:"EventCode"`
	Latitude  float64                `json:"Latitude"`
	Longitude float64                `json:"Longitude"`
	Speed     int                    `json:"Speed"`
	Extras    map[string]interface{} `json:"Extras,omitempty"`
}

type ParsedModel struct {
	IMEI        string            `json:"IMEI"`
	Message     string            `json:"Message"`
	DataPackets int               `json:"DataPackets"`
	ListPackets map[string]Packet `json:"ListPackets"`
}

func (p *ParsedModel) UnmarshalJSON(data []byte) error {

	type Alias ParsedModel
	aux := &struct {
		*Alias
		ListPackets map[string]json.RawMessage `json:"ListPackets"`
	}{
		Alias:       (*Alias)(p),
		ListPackets: make(map[string]json.RawMessage),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.ListPackets = make(map[string]Packet)

	for key, raw := range aux.ListPackets {
		var packet Packet
		if err := json.Unmarshal(raw, &packet); err != nil {
			return err
		}
		p.ListPackets[key] = packet
	}

	return nil
}
