package utils

import (
	"errors"
	"strings"

	"github.com/MaddSystems/jonobridge/common/models"
)

// GetUrl retrieves URL information for a given IMEI
// First tries from local file, then from API if needed
func GetUrl(imei string) (string, error) {
	imei = strings.TrimSpace(imei)

	// Try from file first
	if url, err := getUrlFromFile(imei); err == nil {
		return url, nil
	}

	// Fall back to API
	// Assuming getPlatesFromAPI also fetches/updates the URL field if it's part of the same dataset
	return getPlatesFromAPI(imei) // This might need adjustment if URL data comes from a different source or endpoint
}

// Helper to get url from file
func getUrlFromFile(imei string) (string, error) {
	data, err := LoadFromFile(platesFileName) // Assuming platesFileName contains URL data as well
	if err != nil {
		return "", err
	}

	return searchImeiInData4Url(data, imei)
}

// Search for IMEI in the data model to get URL
func searchImeiInData4Url(data *models.PlatesModel, imei string) (string, error) {
	for _, item := range data.Imeis {
		if strings.TrimSpace(item.Imei) == imei {
			return item.Url, nil
		}
	}
	return "", errors.New("IMEI not found or URL not available")
}
