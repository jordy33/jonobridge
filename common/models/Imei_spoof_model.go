package models

import "encoding/json"

func UnmarshalImeiSpoofModel(data []byte) (ImeiSpoofModel, error) {
	var r ImeiSpoofModel
	r.IMEIMap = make(map[string]string)
	err := json.Unmarshal(data, &r.IMEIMap)
	return r, err
}

func (r *ImeiSpoofModel) Marshal() ([]byte, error) {
	return json.Marshal(r.IMEIMap)
}

type ImeiSpoofModel struct {
	IMEIMap map[string]string // Map of IMEI to Spoof IMEI
}

// GetSpoofIMEI returns the spoof IMEI for a given IMEI if it exists
func (r *ImeiSpoofModel) GetSpoofIMEI(imei string) (string, bool) {
	spoofIMEI, exists := r.IMEIMap[imei]
	return spoofIMEI, exists
}

// SetSpoofIMEI sets or updates a mapping between IMEI and spoof IMEI
func (r *ImeiSpoofModel) SetSpoofIMEI(imei, spoofIMEI string) {
	if r.IMEIMap == nil {
		r.IMEIMap = make(map[string]string)
	}
	r.IMEIMap[imei] = spoofIMEI
}
