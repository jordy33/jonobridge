package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type ElasticLogData struct {
	Client      string  `json:"client"`
	MessengerId string  `json:"messenger_id"`
	DateTime    string  `json:"date_time"`
	Type        string  `json:"type"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Alt         int     `json:"alt"`
}

func SendToElastic(logData ElasticLogData) error {
	elasticURL := os.Getenv("ELASTIC_URL")
	if elasticURL == "" {
		elasticURL = "http://elasticserver.dwim.mx:9200/xpot/_doc"
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		return fmt.Errorf("error marshaling log data: %v", err)
	}

	req, err := http.NewRequest("POST", elasticURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating elastic request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending to elastic: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("elastic error: status code %d", resp.StatusCode)
	}

	return nil
}
