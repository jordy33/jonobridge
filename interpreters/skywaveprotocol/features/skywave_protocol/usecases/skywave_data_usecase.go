package usecases

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"skywaveprotocol/features/skywave_protocol/models"
	"strconv"
	"strings"
	"time"
)

func ParseXML(s *models.GetReturnMessagesResult, data []byte) error {
	err := xml.Unmarshal(data, s)
	if err != nil {
		return err
	}
	return nil
}

func ReturnedMessagesJson(s *models.GetReturnMessagesResult) ([]byte, error) {
	if len(s.Messages.ReturnedMessages) != 0 {
		data, err := json.MarshalIndent(s.Messages.ReturnedMessages, " ", "    ")
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		return nil, fmt.Errorf("messages is empty")
	}
}

func ReturnedMessagesBridge(s *models.GetReturnMessagesResult) ([]models.PayloadBridge, error) {
	if len(s.Messages.ReturnedMessages) != 0 {
		messages := make([]models.PayloadBridge, 0)
		for _, mess := range s.Messages.ReturnedMessages {
			switch mess.Payload.Name {
			case "DistanceCell", "StationaryIntervalSat", "MovingIntervalSat", "MovingEnd", "MovingStart", "IgnitionOn", "StationaryIntervalCell":
				payload := models.PayloadBridge{}
				for _, field := range mess.Payload.Fields.Fields {
					switch field.Name {
					case "Latitude":
						payload.Latitude = field.Value
					case "Longitude":
						payload.Longitude = field.Value
					case "Speed":
						payload.Speed = field.Value
					case "Heading":
						payload.Heading = field.Value
					case "EventTime":
						payload.EventTime = field.Value
					default:
						continue
					}
				}
				payload.ID = mess.ID
				payload.MessageUTC = mess.MessageUTC
				payload.ReceiveUTC = mess.ReceiveUTC
				payload.Type = mess.Payload.Name
				payload.SIN = mess.SIN
				payload.MobileID = mess.MobileID
				payload.Min = mess.Payload.Min
				payload.RegionName = mess.RegionName
				payload.OtaMessageSize = mess.OtaMessageSize
				messages = append(messages, payload)
			default:
				continue
			}
		}
		return messages, nil
	} else {
		return nil, fmt.Errorf("messages is empty")
	}
}

type SkywaveDoc struct {
	Access_id uint64
	Password  string
	From_id   uint64
}

func FromBridgePayload(sky models.PayloadBridge) (map[string]string, error) {
	Output_Map := map[string]string{}
	if len(sky.Latitude) >= 7 && len(sky.Longitude) >= 7 {
		// Divide latitude and longitude
		var latdegrees string
		var latdecimal string
		var londegrees string
		var londecimal string
		if len(sky.Latitude) == 7 {
			latdegrees = sky.Latitude[:4]
			latdecimal = sky.Latitude[4:]
		} else {
			latdegrees = sky.Latitude[:5]
			latdecimal = sky.Latitude[5:]
		}
		if len(sky.Longitude) == 7 {
			londegrees = sky.Longitude[:4]
			londecimal = sky.Longitude[4:]
		} else {
			londegrees = sky.Longitude[:5]
			londecimal = sky.Longitude[5:]
		}

		// Processing steps:
		// 1) Get number with two digits
		// 2) Get decimal remainder
		// 3) Apply sign
		// 4) Divide by 60000
		// 5) Add decimal remainder to the two decimal places in float type
		// 6) Set the result as string

		// Step 1
		latdegreesfloat, err := strconv.ParseFloat(latdegrees, 64)
		if err != nil {
			return Output_Map, err
		}
		latdegreesres := latdegreesfloat / 60
		latpartone := fmt.Sprintf("%0.2f", latdegreesres)
		latdetwodec, err := strconv.ParseFloat(latpartone, 64)
		if err != nil {
			return Output_Map, err
		}

		// Step 2,3,4
		latdecimalfloat, err := strconv.ParseFloat(latdecimal, 64)
		if err != nil {
			return Output_Map, err
		}
		latdecimalres := latdecimalfloat / 60000

		// Step 3
		if strings.Contains(latdegrees, "-") {
			latdecimalres *= -1
		}

		// Step 5
		lat := latdetwodec + latdecimalres
		lat -= 0.003333

		// Process longitude similarly
		londegreesfloat, err := strconv.ParseFloat(londegrees, 64)
		if err != nil {
			return Output_Map, err
		}
		londegreesres := londegreesfloat / 60
		lonpartone := fmt.Sprintf("%0.2f", londegreesres)
		londetwodec, err := strconv.ParseFloat(lonpartone, 64)
		if err != nil {
			return Output_Map, err
		}

		londecimalfloat, err := strconv.ParseFloat(londecimal, 64)
		if err != nil {
			return Output_Map, err
		}
		londecimalres := londecimalfloat / 60000

		if strings.Contains(londegrees, "-") {
			londecimalres *= -1
		}

		lon := londetwodec + londecimalres

		// Date parsing
		datepartial := strings.Replace(sky.ReceiveUTC, " ", "T", 1) + "Z"
		d, err := time.Parse(time.RFC3339, datepartial)
		d2 := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d", d.Year(), d.Month(), d.Day(),
			d.Hour(), d.Minute(), d.Second())

		if err != nil {
			return Output_Map, err
		}

		// Parse Speed
		speed, err := strconv.ParseFloat(sky.Speed, 64)
		if err != nil {
			return Output_Map, err
		}

		// Parse Heading
		dir, err := strconv.ParseUint(sky.Heading, 10, 16)
		if err != nil {
			return Output_Map, err
		}

		Output_Map["Imei"] = sky.MobileID
		Output_Map["Latitude"] = fmt.Sprintf("%v", lat)
		Output_Map["Longitude"] = fmt.Sprintf("%v", lon)
		Output_Map["Speed"] = fmt.Sprintf("%v", speed)
		Output_Map["Driving direction"] = fmt.Sprintf("%v", uint16(dir))
		Output_Map["Date Time"] = d2
		Output_Map["GPS position status"] = "A"
		Output_Map["Protocol Version"] = "3"
		Output_Map["Event code"] = "35"
		Output_Map["Command Type"] = "AAA"
		Output_Map["Altitude"] = "21.232345"

		return Output_Map, nil
	} else {
		return Output_Map, fmt.Errorf("invalid length of latitude or longitude: %d %d", len(sky.Latitude), len(sky.Longitude))
	}
}

func (d *SkywaveDoc) GetDoc() ([]byte, error) {
	url := fmt.Sprintf("https://isatdatapro.skywave.com/GLGW/GWServices_v1/RestMessages.svc/get_return_messages.xml/?access_id=%d&password=%s&from_id=%d", d.Access_id, d.Password, d.From_id)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("ERROR NIL", resp.StatusCode)
			return nil, err
		}
		return bodyBytes, nil
	} else {
		return nil, fmt.Errorf("response with status code %d", resp.StatusCode)
	}
}
