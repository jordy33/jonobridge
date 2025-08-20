package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"crypto/tls"
	"sync"
)

// ElasticLogData represents the data structure for Elasticsearch logging
type ElasticLogData struct {
	Client     string `json:"client"`
	IMEI       string `json:"imei"`
	Payload    string `json:"payload"`
	Time       string `json:"time"`
	StatusCode int    `json:"status-code"`
	StatusText string `json:"status-text"`
}

// IndexSettings represents the settings for creating an Elasticsearch index
type IndexSettings struct {
	Settings struct {
		NumberOfShards   int `json:"number_of_shards"`
		NumberOfReplicas int `json:"number_of_replicas"`
	} `json:"settings"`
}

// indexCache keeps track of indices that have been verified to exist
var (
	indexCache = make(map[string]bool)
	cacheMutex = sync.RWMutex{}
)

// isIndexCached checks if an index is already cached as existing
func isIndexCached(indexName string) bool {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return indexCache[indexName]
}

// cacheIndex marks an index as existing in the cache
func cacheIndex(indexName string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	indexCache[indexName] = true
}

// SendToElastic sends log data to Elasticsearch with proper indexing
func SendToElastic(logData ElasticLogData, customerName string) error {
	// Obtener la URL base de Elasticsearch
	elasticBaseURL := os.Getenv("ELASTIC_URL")
	if elasticBaseURL == "" {
		elasticBaseURL = "https://opensearch.madd.com.mx:9200"
	}

	// Convertir customerName a snake_case en minúsculas
	customerName = ToSnakeCase(customerName)

	// Construir la URL dinámica del índice
	indexName := customerName // Just use the customer name as the index
	elasticURL := fmt.Sprintf("%s/%s/_doc", elasticBaseURL, indexName)

	// Check cache first to avoid unnecessary requests
	if !isIndexCached(indexName) {
		// Check if index exists, create if it doesn't
		exists, err := checkIndexExists(indexName)
		if err != nil {
			VPrint("Error checking index existence: %v", err)
			return fmt.Errorf("error checking index existence: %v", err)
		}

		if !exists {
			VPrint("Index '%s' does not exist, creating it...", indexName)
			if err := createIndex(indexName); err != nil {
				VPrint("Error creating index: %v", err)
				return fmt.Errorf("error creating index: %v", err)
			}
		}
		
		// Cache the index as existing
		cacheIndex(indexName)
	}

	// Debug: Verificar datos antes de enviar
	//VPrint("Elastic URL: %s", elasticURL)
	//VPrint("Log Data: %+v", logData)

	// Convertir los datos a JSON
	jsonData, err := json.Marshal(logData)
	if err != nil {
		VPrint("Error marshaling log data: %v", err)
		return fmt.Errorf("error marshaling log data: %v", err)
	}

	// Debug: Ver JSON generado
	//VPrint("JSON Data to send: %s", string(jsonData))

	// Crear la solicitud HTTP
	req, err := http.NewRequest("POST", elasticURL, bytes.NewBuffer(jsonData))
	if err != nil {
		VPrint("Error creating elastic request: %v", err)
		return fmt.Errorf("error creating elastic request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication if available
	username := os.Getenv("ELASTIC_USER")
	password := os.Getenv("ELASTIC_PASSWORD")
		if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "GPSc0ntr0l1"
	}
	
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// Create HTTP client with TLS config for HTTPS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		VPrint("Error sending to Elasticsearch: %v", err)
		return fmt.Errorf("error sending to elastic: %v", err)
	}
	defer resp.Body.Close()

	// Verificar el código de respuesta
	if resp.StatusCode >= 400 {
		VPrint("Elastic error: status code %d", resp.StatusCode)
		return fmt.Errorf("elastic error: status code %d", resp.StatusCode)
	}

	VPrint("Data successfully sent to Elasticsearch with status code %d", resp.StatusCode)
	return nil
}

// ToSnakeCase converts a string to snake_case format (lowercase with underscores)
func ToSnakeCase(input string) string {
	// Reemplazar espacios por guiones bajos
	re := regexp.MustCompile(`\s+`)
	snake := re.ReplaceAllString(strings.TrimSpace(input), "_")

	// Convertir a minúsculas
	return strings.ToLower(snake)
}

// checkIndexExists checks if an index exists in Elasticsearch
func checkIndexExists(indexName string) (bool, error) {
	elasticBaseURL := os.Getenv("ELASTIC_URL")
	if elasticBaseURL == "" {
		elasticBaseURL = "https://opensearch.madd.com.mx:9200"
	}

	checkURL := fmt.Sprintf("%s/%s", elasticBaseURL, indexName)
	
	req, err := http.NewRequest("HEAD", checkURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating index check request: %v", err)
	}

	// Add authentication - use default credentials if env vars not set
	username := os.Getenv("ELASTIC_USER")
	password := os.Getenv("ELASTIC_PASSWORD")
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "GPSc0ntr0l1"
	}
	req.SetBasicAuth(username, password)

	// Create HTTP client with TLS config for HTTPS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error checking index existence: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// createIndex creates a new index with replicas=0
func createIndex(indexName string) error {
	elasticBaseURL := os.Getenv("ELASTIC_URL")
	if elasticBaseURL == "" {
		elasticBaseURL = "https://opensearch.madd.com.mx:9200"
	}

	createURL := fmt.Sprintf("%s/%s", elasticBaseURL, indexName)
	
	// Debug: Print the URL being used
	//VPrint("Creating index at URL: %s", createURL)

	// Create index settings with shards=1 and replicas=0
	indexSettings := IndexSettings{}
	indexSettings.Settings.NumberOfShards = 1
	indexSettings.Settings.NumberOfReplicas = 0

	jsonData, err := json.Marshal(indexSettings)
	if err != nil {
		return fmt.Errorf("error marshaling index settings: %v", err)
	}

	// Debug: Print the JSON being sent
	//VPrint("Index settings JSON: %s", string(jsonData))

	req, err := http.NewRequest("PUT", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating index creation request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication - use default credentials if env vars not set
	username := os.Getenv("ELASTIC_USER")
	password := os.Getenv("ELASTIC_PASSWORD")
	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "GPSc0ntr0l1"
	}
	req.SetBasicAuth(username, password)
	//VPrint("Using authentication with username: %s", username)

	// Create HTTP client with TLS config for HTTPS
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error creating index: %v", err)
	}
	defer resp.Body.Close()

	// Debug: Read and print the response body
	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)
	//VPrint("Index creation response status: %d", resp.StatusCode)
	//VPrint("Index creation response body: %s", responseBody.String())

	if resp.StatusCode >= 400 {
		return fmt.Errorf("error creating index: status code %d, response: %s", resp.StatusCode, responseBody.String())
	}

	//VPrint("Index '%s' created successfully with replicas=0", indexName)
	return nil
}
