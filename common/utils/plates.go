package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/MaddSystems/jonobridge/common/models"
)

const platesFileName = "data_plates.json"

// GetPlates retrieves plates information for a given IMEI
// First tries from local file, then from API if needed
func GetPlates(imei string) (string, error) {
	imei = strings.TrimSpace(imei)

	// Try from file first
	if plates, err := getPlatesFromFile(imei); err == nil {
		return plates, nil
	}

	// Fall back to API
	return getPlatesFromAPI(imei)
}

// Helper to get plates from file
func getPlatesFromFile(imei string) (string, error) {
	data, err := LoadFromFile(platesFileName)
	if err != nil {
		return "", err
	}

	return searchImeiInData(data, imei)
}

// Helper to get plates from API
func getPlatesFromAPI(imei string) (string, error) {
	apiURL := os.Getenv("PLATES_URL")
	// VPrint("Plates URL: %s", apiURL)

	// Fetch API data
	apiResponse, err := fetchFromURL(apiURL)
	if err != nil {
		return "", fmt.Errorf("API fetch error: %w", err)
	}

	// Parse response
	data, err := LoadFromString(apiResponse)
	if err != nil {
		return "", fmt.Errorf("API response parsing error: %w", err)
	}

	// Save for future use
	SaveToFile(data, platesFileName)

	// Search for the IMEI
	return searchImeiInData(data, imei)
}

// Search for IMEI in the data model
func searchImeiInData(data *models.PlatesModel, imei string) (string, error) {
	for _, item := range data.Imeis {
		if strings.TrimSpace(item.Imei) == imei {
			return item.Plates, nil
		}
	}
	return "", errors.New("IMEI no encontrado")
}

// LoadFromFile loads plates data from a file
func LoadFromFile(fileName string) (*models.PlatesModel, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var plates models.PlatesModel
	if err := json.Unmarshal(byteValue, &plates); err != nil {
		return nil, err
	}

	return &plates, nil
}

// SaveToFile saves plates data to a file
func SaveToFile(plates *models.PlatesModel, fileName string) error {
	jsonData, err := json.MarshalIndent(plates, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, jsonData, 0644)
}

// LoadFromString creates a PlatesModel from a JSON string
func LoadFromString(jsonData string) (*models.PlatesModel, error) {
	var plates models.PlatesModel
	err := json.Unmarshal([]byte(jsonData), &plates)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling plates: %v", err)
	}
	return &plates, nil
}

// fetchFromURL retrieves data from an external API
func fetchFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
