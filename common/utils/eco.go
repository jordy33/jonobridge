package utils

import (
	"errors"
	"strings"

	"github.com/MaddSystems/jonobridge/common/models"
)

// GetPlates retrieves plates information for a given IMEI
// First tries from local file, then from API if needed

func GetEco(imei string) (string, error) {
	imei = strings.TrimSpace(imei)

	// Try from file first
	if plates, err := getEcoFromFile(imei); err == nil {
		return plates, nil
	}

	// Fall back to API
	return getPlatesFromAPI(imei)
}

// Helper to get plates from file
func getEcoFromFile(imei string) (string, error) {
	data, err := LoadFromFile(platesFileName)
	if err != nil {
		return "", err
	}

	return searchImeiInData4Eco(data, imei)
}

// Search for IMEI in the data model
func searchImeiInData4Eco(data *models.PlatesModel, imei string) (string, error) {
	for _, item := range data.Imeis {
		if strings.TrimSpace(item.Imei) == imei {
			return item.Eco, nil
		}
	}
	return "", errors.New("IMEI no encontrado")
}
