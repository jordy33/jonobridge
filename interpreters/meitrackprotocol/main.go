package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"meitrackprotocol/features/jono"
	"meitrackprotocol/features/meitrack_protocol"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	verbose = flag.Bool("v", false, "Enable verbose logging") // Verbose flag
)

// CircuitBreaker implements a simple circuit breaker pattern
type CircuitBreaker struct {
	maxFailures  int64
	resetTimeout time.Duration
	failures     int64
	lastFailTime time.Time
	state        int32 // 0: closed, 1: open, 2: half-open
	mutex        sync.RWMutex
}

func NewCircuitBreaker(maxFailures int64, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.RLock()
	state := atomic.LoadInt32(&cb.state)
	cb.mutex.RUnlock()

	// If circuit is open, check if we should transition to half-open
	if state == 1 { // open
		cb.mutex.Lock()
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			atomic.StoreInt32(&cb.state, 2) // half-open
		}
		cb.mutex.Unlock()
		if atomic.LoadInt32(&cb.state) == 1 {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	err := fn()

	if err != nil {
		newFailures := atomic.AddInt64(&cb.failures, 1)
		cb.mutex.Lock()
		cb.lastFailTime = time.Now()
		if newFailures >= cb.maxFailures {
			atomic.StoreInt32(&cb.state, 1) // open
		}
		cb.mutex.Unlock()
		return err
	}

	// Success - reset failures and close circuit
	atomic.StoreInt64(&cb.failures, 0)
	atomic.StoreInt32(&cb.state, 0) // closed
	return nil
}

// HealthMonitor tracks application health metrics
type HealthMonitor struct {
	messagesProcessed int64
	errorsCount       int64
	lastMessageTime   time.Time
	startTime         time.Time
	mutex             sync.RWMutex
}

func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		startTime: time.Now(),
	}
}

func (hm *HealthMonitor) RecordMessage() {
	atomic.AddInt64(&hm.messagesProcessed, 1)
	hm.mutex.Lock()
	hm.lastMessageTime = time.Now()
	hm.mutex.Unlock()
}

func (hm *HealthMonitor) RecordError() {
	atomic.AddInt64(&hm.errorsCount, 1)
}

func (hm *HealthMonitor) GetStats() (int64, int64, time.Time, time.Duration) {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	return atomic.LoadInt64(&hm.messagesProcessed),
		atomic.LoadInt64(&hm.errorsCount),
		hm.lastMessageTime,
		time.Since(hm.startTime)
}

type TrackerData struct {
	Payload    string `json:"payload"`
	RemoteAddr string `json:"remoteaddr"`
}

type TrackerAssign struct {
	Imei       string `json:"imei"`
	Protocol   string `json:"protocol"`
	RemoteAddr string `json:"remoteaddr"`
}

// Helper function to print verbose logs if enabled
func vPrint(format string, v ...interface{}) {
	if *verbose {
		log.Printf(format, v...)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MQTTClient wraps the MQTT client with additional functionality
type MQTTClient struct {
	client         mqtt.Client
	brokerURL      string
	clientID       string
	verbose        bool
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	publishTimeout time.Duration
	maxGoroutines  int
	semaphore      chan struct{}
	healthMonitor  *HealthMonitor
	circuitBreaker *CircuitBreaker
	lastHeartbeat  time.Time
	heartbeatMutex sync.RWMutex
}

// NewMQTTClient creates a new MQTT client with the given configuration
func NewMQTTClient(brokerHost string, clientID string, verbose bool) (*MQTTClient, error) {
	if brokerHost == "" {
		return nil, fmt.Errorf("broker host cannot be empty")
	}

	brokerURL := fmt.Sprintf("tcp://%s:1883", brokerHost)
	ctx, cancel := context.WithCancel(context.Background())

	// Limit concurrent goroutines to prevent resource exhaustion
	maxGoroutines := runtime.NumCPU() * 2
	if maxGoroutines < 4 {
		maxGoroutines = 4
	}

	return &MQTTClient{
		brokerURL:      brokerURL,
		clientID:       clientID,
		verbose:        verbose,
		ctx:            ctx,
		cancel:         cancel,
		publishTimeout: 30 * time.Second,
		maxGoroutines:  maxGoroutines,
		semaphore:      make(chan struct{}, maxGoroutines),
		healthMonitor:  NewHealthMonitor(),
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second), // 5 failures in 30 seconds opens circuit
		lastHeartbeat:  time.Now(),
	}, nil
}

// Connect establishes a connection to the MQTT broker
func (m *MQTTClient) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.brokerURL)

	// Generate a unique client ID
	subscribe_topic := "meitrack" // You might want to make this configurable
	clientID := fmt.Sprintf("meitrackprotocol_%s_%s_%d",
		subscribe_topic,
		os.Getenv("HOSTNAME"),
		time.Now().UnixNano()%100000)
	opts.SetClientID(clientID)

	// Configure settings for multiple listeners
	opts.SetCleanSession(false) // Maintain persistent session
	opts.SetAutoReconnect(true) // Auto reconnect on connection loss
	opts.SetKeepAlive(60 * time.Second)
	opts.SetOrderMatters(true) // Maintain message order
	opts.SetResumeSubs(true)   // Resume stored subscriptions
	opts.SetDefaultPublishHandler(m.messageHandler)

	// Add connection lost handler
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		if m.verbose {
			vPrint("MQTT connection lost: %v. Will attempt to reconnect...", err)
		}
	})

	// Add reconnect handler
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if m.verbose {
			vPrint("MQTT connection established/re-established")
		}
	})

	m.client = mqtt.NewClient(opts)

	// Try to connect with retries and timeout
	maxRetries := 10
	retryDelay := 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-m.ctx.Done():
			return fmt.Errorf("connection cancelled")
		default:
		}

		if token := m.client.Connect(); token.WaitTimeout(30*time.Second) && token.Error() != nil {
			if m.verbose {
				vPrint("Error connecting to MQTT broker at %s (attempt %d/%d): %v. Retrying in %v...",
					m.brokerURL, attempt, maxRetries, token.Error(), retryDelay)
			}
			if attempt == maxRetries {
				return fmt.Errorf("failed to connect after %d attempts: %v", maxRetries, token.Error())
			}
			time.Sleep(retryDelay)
			continue
		}
		if m.verbose {
			vPrint("Successfully connected to MQTT broker!")
		}
		break
	}

	return nil
}

// Subscribe subscribes to the specified topic
func (m *MQTTClient) Subscribe(topic string, qos byte) error {
	if token := m.client.Subscribe(topic, qos, nil); token.Wait() && token.Error() != nil {
		return fmt.Errorf("error subscribing to topic %s: %v", topic, token.Error())
	}
	if m.verbose {
		vPrint("Subscribed to topic: %s", topic)
	}
	return nil
}

// Shutdown gracefully shuts down the MQTT client
func (m *MQTTClient) Shutdown() {
	if m.verbose {
		vPrint("Shutting down MQTT client...")
	}

	// Cancel context to stop all goroutines
	m.cancel()

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if m.verbose {
			vPrint("All goroutines finished")
		}
	case <-time.After(10 * time.Second):
		if m.verbose {
			vPrint("Timeout waiting for goroutines to finish")
		}
	}

	// Disconnect MQTT client
	if m.client.IsConnected() {
		m.client.Disconnect(1000) // 1 second timeout
	}

	if m.verbose {
		vPrint("MQTT client shutdown complete")
	}
}

// UpdateHeartbeat updates the last heartbeat time
func (m *MQTTClient) UpdateHeartbeat() {
	m.heartbeatMutex.Lock()
	m.lastHeartbeat = time.Now()
	m.heartbeatMutex.Unlock()
}

// GetLastHeartbeat returns the last heartbeat time
func (m *MQTTClient) GetLastHeartbeat() time.Time {
	m.heartbeatMutex.RLock()
	defer m.heartbeatMutex.RUnlock()
	return m.lastHeartbeat
}

// IsHealthy checks if the client is healthy based on heartbeat and connection status
func (m *MQTTClient) IsHealthy() bool {
	if !m.client.IsConnected() {
		return false
	}

	lastHeartbeat := m.GetLastHeartbeat()
	return time.Since(lastHeartbeat) < 2*time.Minute // Consider unhealthy if no activity for 2 minutes
}

// Publish publishes a message to the specified topic with timeout and retry logic
func (m *MQTTClient) Publish(topic string, payload interface{}) error {
	// Use circuit breaker to prevent cascading failures
	return m.circuitBreaker.Call(func() error {
		// Create a context with timeout for this publish operation
		ctx, cancel := context.WithTimeout(m.ctx, m.publishTimeout)
		defer cancel()

		// Check if client is connected
		if !m.client.IsConnected() {
			m.healthMonitor.RecordError()
			return fmt.Errorf("MQTT client is not connected")
		}

		// Channel to receive the result
		done := make(chan error, 1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					done <- fmt.Errorf("panic in publish: %v", r)
				}
			}()

			token := m.client.Publish(topic, 0, false, payload)
			if token.WaitTimeout(m.publishTimeout) {
				done <- token.Error()
			} else {
				done <- fmt.Errorf("publish timeout for topic %s", topic)
			}
		}()

		select {
		case err := <-done:
			if err != nil {
				m.healthMonitor.RecordError()
				if m.verbose {
					vPrint("Failed to publish to MQTT topic %s: %v", topic, err)
				}
				return err
			}
			if m.verbose {
				vPrint("Successfully published to MQTT topic: %s", topic)
			}
			return nil
		case <-ctx.Done():
			m.healthMonitor.RecordError()
			return fmt.Errorf("publish cancelled or timed out for topic %s", topic)
		}
	})
}

// messageHandler handles incoming MQTT messages
func (m *MQTTClient) messageHandler(client mqtt.Client, msg mqtt.Message) {
	// Update heartbeat to show we're processing messages
	m.UpdateHeartbeat()
	m.healthMonitor.RecordMessage()

	// Check if context is cancelled
	select {
	case <-m.ctx.Done():
		return
	default:
	}

	// Acquire semaphore to limit concurrent processing
	select {
	case m.semaphore <- struct{}{}:
		defer func() { <-m.semaphore }()
	case <-m.ctx.Done():
		return
	case <-time.After(5 * time.Second): // Don't wait forever
		if m.verbose {
			vPrint("Dropping message due to semaphore timeout")
		}
		m.healthMonitor.RecordError()
		return
	}

	// Add to wait group for graceful shutdown
	m.wg.Add(1)
	defer m.wg.Done()

	// Add recovery to prevent panics from crashing the handler
	defer func() {
		if r := recover(); r != nil {
			m.healthMonitor.RecordError()
			log.Printf("Recovered from panic in message handler: %v", r)
			debug.PrintStack()
		}
	}()

	// Set a timeout for message processing
	processingCtx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if msg.Topic() == "tracker/from-udp" {
			m.processUDPMessage(msg)
		} else {
			m.processTCPMessage(msg)
		}
	}()

	select {
	case <-done:
		// Processing completed normally
	case <-processingCtx.Done():
		m.healthMonitor.RecordError()
		if m.verbose {
			vPrint("Message processing timed out for topic: %s", msg.Topic())
		}
	}
}

// processUDPMessage handles UDP messages
func (m *MQTTClient) processUDPMessage(msg mqtt.Message) {
	defer func() {
		if r := recover(); r != nil {
			m.healthMonitor.RecordError()
			log.Printf("Panic in processUDPMessage: %v", r)
		}
	}()

	if len(msg.Payload()) == 0 {
		if m.verbose {
			vPrint("Ignoring empty UDP message")
		}
		return
	}

	if len(msg.Payload()) > 100*1024 { // 100KB limit
		if m.verbose {
			vPrint("Message too large, ignoring UDP message of size: %d", len(msg.Payload()))
		}
		m.healthMonitor.RecordError()
		return
	}

	trackerPayload := string(msg.Payload())

	var trackerData string
	// Try to decode as hex, if it fails, use the original message
	bytes, err := hex.DecodeString(trackerPayload)
	if err != nil {
		trackerData = trackerPayload
		if m.verbose {
			vPrint("error: %s", err)
		}
	} else {
		trackerData = string(bytes)
	}

	// Preliminary check to see if the message is a valid format
	if !strings.HasPrefix(trackerData, "$$") && !strings.HasPrefix(trackerData, "@@") {
		if m.verbose {
			vPrint("Ignoring message with invalid protocol format: %s", trackerData)
		}
		return
	}

	fields := strings.Split(trackerData, ",")
	if len(fields) <= 2 {
		if m.verbose {
			vPrint("Not enough fields in message: %d", len(fields))
		}
		return
	}

	dataMeitrack, err := meitrack_protocol.Initialize(trackerData)
	if err != nil {
		m.healthMonitor.RecordError()
		if m.verbose {
			vPrint("Error initializing Meitrack protocol: %v", err)
		}
		return
	}

	jonoNormalize, err := jono.Initialize(dataMeitrack)
	if err != nil {
		m.healthMonitor.RecordError()
		if m.verbose {
			vPrint("Error initializing Jono protocol: %v", err)
		}
		return
	}

	// Use a separate goroutine for publishing with context cancellation
	m.wg.Add(1)
	go func(jonoData string) {
		defer m.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				m.healthMonitor.RecordError()
				log.Printf("Panic in UDP publish goroutine: %v", r)
			}
		}()

		select {
		case <-m.ctx.Done():
			return
		default:
		}

		// Publish to jonoprotocol topic
		if err := m.Publish("tracker/jonoprotocol", jonoData); err != nil {
			log.Printf("Error publishing to jonoprotocol: %v", err)
			return
		}
		if m.verbose {
			// Create a compact version of the JSON for logging
			var jsonObj map[string]interface{}
			if err := json.Unmarshal([]byte(jonoData), &jsonObj); err == nil {
				// Re-marshal without indentation
				if compactJSON, err := json.Marshal(jsonObj); err == nil {
					vPrint("Jono Protocol: %s", string(compactJSON))
				} else {
					vPrint("Jono Protocol: %s", jonoData) // Fallback to pretty JSON if compact fails
				}
			} else {
				vPrint("Jono Protocol: %s", jonoData) // Fallback to pretty JSON if unmarshaling fails
			}
		}
	}(jonoNormalize)
}

// processTCPMessage handles TCP messages
func (m *MQTTClient) processTCPMessage(msg mqtt.Message) {
	defer func() {
		if r := recover(); r != nil {
			m.healthMonitor.RecordError()
			log.Printf("Panic in processTCPMessage: %v", r)
		}
	}()

	if len(msg.Payload()) == 0 {
		if m.verbose {
			vPrint("Ignoring empty TCP message")
		}
		return
	}

	if len(msg.Payload()) > 100*1024 { // 100KB limit
		if m.verbose {
			vPrint("Message too large, ignoring TCP message of size: %d", len(msg.Payload()))
		}
		m.healthMonitor.RecordError()
		return
	}

	// Handle JSON messages
	var json_data TrackerData
	if err := json.Unmarshal(msg.Payload(), &json_data); err != nil {
		m.healthMonitor.RecordError()
		log.Printf("Error unmarshaling JSON: %v", err)
		return
	}

	if json_data.Payload == "" {
		if m.verbose {
			vPrint("Empty payload in JSON message")
		}
		return
	}

	payload := json_data.Payload
	remote_addr := json_data.RemoteAddr
	trackerPayload := payload

	var trackerData string
	// Try to decode as hex, if it fails, use the original message
	bytes, err := hex.DecodeString(trackerPayload)
	if err != nil {
		trackerData = trackerPayload
		if m.verbose {
			vPrint("error: %s", err)
		}
	} else {
		trackerData = string(bytes)
	}

	if m.verbose {
		vPrint("Received message on topic %s: %s", msg.Topic(), msg.Payload())
		tracker_bytes := []byte(trackerData)
		vPrint("Received message on topic : %s", hex.EncodeToString(tracker_bytes[:min(32, len(tracker_bytes))]))
	}

	// Preliminary check to see if the message is a valid format
	if !strings.HasPrefix(trackerData, "$$") && !strings.HasPrefix(trackerData, "@@") {
		if m.verbose {
			vPrint("Ignoring message with invalid protocol format: %s", trackerData)
		}
		return
	}

	fields := strings.Split(trackerData, ",")
	if len(fields) <= 2 {
		if m.verbose {
			vPrint("Not enough fields in TCP message: %d", len(fields))
		}
		return
	}

	dataMeitrack, err := meitrack_protocol.Initialize(trackerData)
	if err != nil {
		m.healthMonitor.RecordError()
		if m.verbose {
			vPrint("Error initializing Meitrack protocol: %v", err)
		}
		return
	}

	if len(fields) < 2 {
		if m.verbose {
			vPrint("Invalid IMEI field in message")
		}
		return
	}
	imei := fields[1]

	if len(imei) < 10 || len(imei) > 20 { // Basic IMEI validation
		if m.verbose {
			vPrint("Invalid IMEI format: %s", imei)
		}
		m.healthMonitor.RecordError()
		return
	}

	jonoNormalize, err := jono.Initialize(dataMeitrack)
	if err != nil {
		m.healthMonitor.RecordError()
		if m.verbose {
			vPrint("Error initializing Jono protocol: %v", err)
		}
		return
	}

	// Publish to jonoprotocol topic
	if err := m.Publish("tracker/jonoprotocol", jonoNormalize); err != nil {
		log.Printf("Error publishing to jonoprotocol: %v", err)
		return
	}
	if m.verbose {
		// Create a compact version of the JSON for logging
		var jsonObj map[string]interface{}
		if err := json.Unmarshal([]byte(jonoNormalize), &jsonObj); err == nil {
			// Re-marshal without indentation
			if compactJSON, err := json.Marshal(jsonObj); err == nil {
				vPrint("Jono Protocol: %s", string(compactJSON))
			} else {
				vPrint("Jono Protocol: %s", jonoNormalize) // Fallback to pretty JSON if compact fails
			}
		} else {
			vPrint("Jono Protocol: %s", jonoNormalize) // Fallback to pretty JSON if unmarshaling fails
		}
	}
	tracker_data_json := TrackerAssign{
		Imei:       imei,
		Protocol:   "meitrack",
		RemoteAddr: remote_addr,
	}
	// Marshal the data to JSON before publishing
	assignImeiJson, err := json.Marshal(tracker_data_json)
	if err != nil {
		m.healthMonitor.RecordError()
		log.Printf("Error marshaling assign-imei data: %v", err)
		return
	}
	// Publish to assign-imei2remoteaddr topic
	if err := m.Publish("tracker/assign-imei2remoteaddr", assignImeiJson); err != nil {
		log.Printf("Error publishing to assign-imei2remoteaddr: %v", err)
	}
	if m.verbose {
		log.Printf("Publishing to tracker/assign-imei2remoteaddr: %s", assignImeiJson)
	}
}

func main() {
	// Parse command-line flags
	flag.Parse()

	// Set up logging with timestamps
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Set garbage collection target
	debug.SetGCPercent(100) // More aggressive GC

	// Set max procs to prevent resource exhaustion
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Get MQTT broker host from environment
	mqttBrokerHost := os.Getenv("MQTT_BROKER_HOST")
	if mqttBrokerHost == "" {
		log.Fatal("MQTT_BROKER_HOST environment variable not set")
	}

	// Create and configure MQTT client
	mqttClient, err := NewMQTTClient(mqttBrokerHost, "go_mqtt_client", *verbose)
	if err != nil {
		log.Fatal("Failed to create MQTT client:", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Set up shutdown handling
	shutdown := make(chan struct{})
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		mqttClient.Shutdown()
		close(shutdown)
	}()

	// Connect to the broker
	if err := mqttClient.Connect(); err != nil {
		log.Fatal("Failed to connect to MQTT broker:", err)
	}

	// Subscribe to the connection topic
	if err := mqttClient.Subscribe("tracker/from-tcp", 1); err != nil {
		log.Fatal("Failed to subscribe to tracker/from-tcp:", err)
	}

	// Subscribe to the raw topic
	if err := mqttClient.Subscribe("tracker/from-udp", 1); err != nil {
		log.Fatal("Failed to subscribe to tracker/from-udp:", err)
	}

	log.Println("MQTT client started successfully. Waiting for messages...")

	// Add comprehensive health check and monitoring
	healthTicker := time.NewTicker(1 * time.Minute)
	defer healthTicker.Stop()

	statsTicker := time.NewTicker(5 * time.Minute)
	defer statsTicker.Stop()

	deadlockTicker := time.NewTicker(10 * time.Second)
	defer deadlockTicker.Stop()

	go func() {
		for {
			select {
			case <-healthTicker.C:
				if !mqttClient.IsHealthy() {
					log.Printf("Health check FAILED: Client is unhealthy!")
					log.Printf("Connected: %v, Last heartbeat: %v",
						mqttClient.client.IsConnected(),
						mqttClient.GetLastHeartbeat())

					// Try to reconnect first
					if !mqttClient.client.IsConnected() {
						log.Printf("Attempting to reconnect...")
						if err := mqttClient.Connect(); err != nil {
							log.Printf("Reconnection failed: %v", err)
						}
					}

					// Force restart regardless of reconnection result
					// This ensures the system recovers even if it becomes unresponsive
					log.Printf("=== SYSTEM IS UNHEALTHY - RESTARTING ===")
					
					// Graceful shutdown to clean up resources
					mqttClient.Shutdown()
					
					// Force restart by exiting with error code 1
					os.Exit(1)
				} else if *verbose {
					vPrint("Health check: MQTT client is healthy")
				}

			case <-statsTicker.C:
				messages, errors, lastMsg, uptime := mqttClient.healthMonitor.GetStats()
				log.Printf("Stats - Messages: %d, Errors: %d, Last Message: %v, Uptime: %v, Goroutines: %d",
					messages, errors, lastMsg, uptime, runtime.NumGoroutine())

				// Force garbage collection if too many goroutines
				if runtime.NumGoroutine() > 1000 {
					log.Printf("High goroutine count detected, forcing GC")
					runtime.GC()
				}

			case <-deadlockTicker.C:
				// Check for potential deadlocks by monitoring goroutine count
				if runtime.NumGoroutine() > 10000 {
					log.Printf("CRITICAL: Very high goroutine count: %d - possible goroutine leak!", runtime.NumGoroutine())
				}

			case <-mqttClient.ctx.Done():
				return
			}
		}
	}()

	// Memory monitoring
	go func() {
		memTicker := time.NewTicker(2 * time.Minute)
		defer memTicker.Stop()

		for {
			select {
			case <-memTicker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				// Log if memory usage is high
				if m.Alloc > 100*1024*1024 { // 100MB
					log.Printf("High memory usage: Alloc=%d KB, Sys=%d KB, Goroutines=%d",
						m.Alloc/1024, m.Sys/1024, runtime.NumGoroutine())
				}

				// Force GC if memory usage is very high
				if m.Alloc > 500*1024*1024 { // 500MB
					log.Printf("Forcing garbage collection due to high memory usage")
					runtime.GC()
				}

			case <-mqttClient.ctx.Done():
				return
			}
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	log.Println("Application shutdown complete")
}
