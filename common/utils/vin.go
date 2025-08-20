package utils

import (
	"errors"
	"strings"

	"github.com/MaddSystems/jonobridge/common/models"
)

// GetPlates retrieves plates information for a given IMEI
// First tries from local file, then from API if needed
func GetVin(imei string) (string, error) {
	imei = strings.TrimSpace(imei)

	// Try from file first
	if vin, err := getVinFromFile(imei); err == nil {
		return vin, nil
	}

	// Fall back to API
	return getPlatesFromAPI(imei)
}

// Helper to get plates from file
func getVinFromFile(imei string) (string, error) {
	data, err := LoadFromFile(platesFileName)
	if err != nil {
		return "", err
	}

	return searchImeiInData4Vin(data, imei)
}

// Search for IMEI in the data model
func searchImeiInData4Vin(data *models.PlatesModel, imei string) (string, error) {
	for _, item := range data.Imeis {
		if strings.TrimSpace(item.Imei) == imei {
			return item.Vin, nil
		}
	}
	return "", errors.New("IMEI no encontrado")
}
