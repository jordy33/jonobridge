package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MaddSystems/jonobridge/common/models"
)

const imei_spoofFileName = "data_imei_spoof.json"

// ImeiSpoofMapping represents a mapping between original and spoofed IMEI
type ImeiSpoofMapping struct {
	Original string `json:"original"`
	Spoofed  string `json:"spoofed"`
}

// ImeiSpoofData contains a list of IMEI mappings
type ImeiSpoofData struct {
	Imeis []ImeiSpoofMapping `json:"imeis"`
}

// GetImei_spoof retrieves imei_spoof information for a given IMEI
// First tries from local file, then from API if needed
func GetImei_spoof(imei string) (string, error) {
	imei = strings.TrimSpace(imei)

	// Try from file first
	if imei_spoof, err := getImei_spoofFromFile(imei); err == nil {
		return imei_spoof, nil
	}

	// Fall back to API
	return getImei_spoofFromAPI(imei)
}

// Helper to get imei_spoof from file
func getImei_spoofFromFile(imei string) (string, error) {
	data, err := LoadFromFileImeiSpoof(imei_spoofFileName)
	if err != nil {
		return "", err
	}

	return searchImeiInDataImeiSpoof(data, imei)
}

// Helper to get imei_spoof from API
func getImei_spoofFromAPI(imei string) (string, error) {
	apiURL := os.Getenv("SPOOF_IMEI_URL")
	if apiURL == "" {
		apiURL = "https://pluto.dudewhereismy.com.mx/virtualimeis?appId=244"
	}
	// VPrint("Imei_spoof URL: %s", apiURL)

	// Fetch API data
	apiResponse, err := fetchFromURLImeiSpoof(apiURL)
	if err != nil {
		return "", fmt.Errorf("API fetch error: %w", err)
	}

	// Parse response
	data, err := LoadFromStringImeiSpoof(apiResponse)
	if err != nil {
		return "", fmt.Errorf("API response parsing error: %w", err)
	}

	// Save for future use
	SaveToFileImeiSpoof(data, imei_spoofFileName)

	// Search for the IMEI
	return searchImeiInDataImeiSpoof(data, imei)
}

// Search for IMEI in the data model
func searchImeiInDataImeiSpoof(data *models.ImeiSpoofModel, imei string) (string, error) {
	spoofImei, exists := data.GetSpoofIMEI(imei)
	if !exists {
		return "", errors.New("IMEI not found")
	}
	return spoofImei, nil
}

// LoadFromFileImeiSpoof loads imei_spoof data from a file
func LoadFromFileImeiSpoof(fileName string) (*models.ImeiSpoofModel, error) {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var imeiSpoofModel models.ImeiSpoofModel
	if err := json.Unmarshal(byteValue, &imeiSpoofModel.IMEIMap); err != nil {
		return nil, err
	}

	return &imeiSpoofModel, nil
}

// SaveToFileImeiSpoof saves imei_spoof data to a file
func SaveToFileImeiSpoof(imei_spoof *models.ImeiSpoofModel, fileName string) error {
	jsonData, err := json.MarshalIndent(imei_spoof.IMEIMap, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, jsonData, 0644)
}

// LoadFromStringImeiSpoof creates a Imei_spoofModel from a JSON string
func LoadFromStringImeiSpoof(jsonData string) (*models.ImeiSpoofModel, error) {
	var imeiSpoofModel models.ImeiSpoofModel
	imeiSpoofModel.IMEIMap = make(map[string]string)

	err := json.Unmarshal([]byte(jsonData), &imeiSpoofModel.IMEIMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling imei_spoof: %v", err)
	}

	return &imeiSpoofModel, nil
}

// fetchFromURLImeiSpoof retrieves data from an external API
func fetchFromURLImeiSpoof(url string) (string, error) {
	// Create HTTP client with timeout to prevent hanging
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make the request
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// FetchAndSaveImeiMappings retrieves IMEI mappings from the endpoint and converts them to our format
func FetchAndSaveImeiMappings(fileName string) error {
	// Get URL from environment
	apiURL := os.Getenv("SPOOF_IMEI_URL")
	if apiURL == "" {
		apiURL = "https://pluto.dudewhereismy.com.mx/virtualimeis?appId=244"
	}

	// Fetch API data
	apiResponse, err := fetchFromURLImeiSpoof(apiURL)
	if err != nil {
		return fmt.Errorf("API fetch error: %w", err)
	}

	// Parse as map (original format)
	var endpointData map[string]string
	if err := json.Unmarshal([]byte(apiResponse), &endpointData); err != nil {
		return fmt.Errorf("failed to parse endpoint data: %w", err)
	}

	// Convert to our format
	spoofData := ImeiSpoofData{
		Imeis: make([]ImeiSpoofMapping, 0, len(endpointData)),
	}

	for original, spoofed := range endpointData {
		spoofData.Imeis = append(spoofData.Imeis, ImeiSpoofMapping{
			Original: original,
			Spoofed:  spoofed,
		})
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(spoofData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	// Write to file
	err = ioutil.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		// If we can't write to the specified location, try the root directory
		if os.IsPermission(err) || os.IsNotExist(err) {
			rootPath := "/data_spoof.json"
			VPrint("Could not write to %s, trying %s", fileName, rootPath)
			return ioutil.WriteFile(rootPath, jsonData, 0644)
		}
		return err
	}
	return nil
}

// GetSpoofimeiFromJson retrieves a spoofed IMEI from the JSON file
func GetSpoofimeiFromJson(imei string, fileName string) (string, error) {
	// Try multiple common locations for the file
	possiblePaths := []string{
		fileName,                               // Direct path provided
		"/data_spoof.json",                     // Root of container
		os.Getenv("HOME") + "/data_spoof.json", // Home directory
	}

	var jsonData []byte
	var err error
	var foundPath string

	// Try each path until we find the file
	for _, path := range possiblePaths {
		if _, statErr := os.Stat(path); statErr == nil {
			// File exists at this path
			foundPath = path
			jsonData, err = ioutil.ReadFile(path)
			if err == nil {
				// Successfully read the file
				break
			}
		}
	}

	// File not found in any location, create it at root for container compatibility
	if jsonData == nil {
		//VPrint("IMEI mapping file not found, fetching from endpoint...")
		// Use the root path for containers
		savePath := "/data_spoof.json"
		if err := FetchAndSaveImeiMappings(savePath); err != nil {
			return "", fmt.Errorf("failed to fetch and save IMEI mappings: %w", err)
		}

		// Now read the newly created file
		foundPath = savePath
		jsonData, err = ioutil.ReadFile(savePath)
		if err != nil {
			return "", fmt.Errorf("failed to read JSON file after creation: %w", err)
		}
	}

	//VPrint("Using IMEI mapping file at: %s", foundPath)

	// Parse the JSON data
	var data ImeiSpoofData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return "", fmt.Errorf("failed to parse JSON data: %w", err)
	}

	// Find the spoofed IMEI for the given IMEI
	for _, mapping := range data.Imeis {
		if mapping.Original == imei {
			return mapping.Spoofed, nil
		}
	}

	// If we couldn't find the IMEI in our file, try fetching fresh data
	VPrint("IMEI %s not found in local data, fetching fresh data...", imei)
	if err := FetchAndSaveImeiMappings(foundPath); err != nil {
		return "", fmt.Errorf("failed to refresh IMEI mappings: %w", err)
	}

	// Read refreshed data
	jsonData, err = ioutil.ReadFile(foundPath)
	if err != nil {
		return "", fmt.Errorf("failed to read refreshed JSON file: %w", err)
	}

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return "", fmt.Errorf("failed to parse refreshed JSON data: %w", err)
	}

	// Try to find the IMEI again
	for _, mapping := range data.Imeis {
		if mapping.Original == imei {
			return mapping.Spoofed, nil
		}
	}

	return "", fmt.Errorf("no spoofed IMEI found for IMEI %s", imei)
}
