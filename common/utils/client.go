package utils

import (
	"errors"
	"strings"

	"github.com/MaddSystems/jonobridge/common/models"
)

// GetClient retrieves client information for a given IMEI
// First tries from local file, then from API if needed
func GetClient(imei string) (string, error) {
	imei = strings.TrimSpace(imei)

	// Try from file first
	if client, err := getClientFromFile(imei); err == nil {
		return client, nil
	}

	// Fall back to API
	return getPlatesFromAPI(imei)
}

// Helper to get client from file
func getClientFromFile(imei string) (string, error) {
	data, err := LoadFromFile(platesFileName)
	if err != nil {
		return "", err
	}

	return searchImeiInData4Client(data, imei)
}

// Search for IMEI in the data model to get Client
func searchImeiInData4Client(data *models.PlatesModel, imei string) (string, error) {
	for _, item := range data.Imeis {
		if strings.TrimSpace(item.Imei) == imei {
			return item.Client, nil
		}
	}
	return "", errors.New("IMEI no encontrado")
}
