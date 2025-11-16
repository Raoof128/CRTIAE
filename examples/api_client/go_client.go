package main

// Red Team C2 Framework - Go API Client Library
//
// Full-featured Go client for programmatic access to the C2 infrastructure.
// Provides idiomatic Go interface to all teamserver endpoints.
//
// Usage:
//     import "github.com/Raoof128/red-team-c2/examples/api_client"
//
//     client := NewC2Client("https://teamserver:8443")
//     beacons, err := client.ListBeacons()
//     err = client.SubmitCommand(beaconID, "whoami")

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Beacon represents a C2 beacon/implant
type Beacon struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	Username  string    `json:"username"`
	OS        string    `json:"os"`
	IP        string    `json:"ip"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Status    string    `json:"status"`
}

// Command represents a queued command
type Command struct {
	ID        string    `json:"id"`
	BeaconID  string    `json:"beacon_id"`
	Command   string    `json:"command"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// Output represents command output
type Output struct {
	CommandID string    `json:"command_id"`
	BeaconID  string    `json:"beacon_id"`
	Output    string    `json:"output"`
	Timestamp time.Time `json:"timestamp"`
}

// C2Client is the main API client
type C2Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewC2Client creates a new C2 API client
func NewC2Client(baseURL string) *C2Client {
	return &C2Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // For testing only
				},
			},
		},
	}
}

// WithTimeout sets custom timeout for the client
func (c *C2Client) WithTimeout(timeout time.Duration) *C2Client {
	c.HTTPClient.Timeout = timeout
	return c
}

// HealthCheck checks if teamserver is healthy
func (c *C2Client) HealthCheck() error {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unhealthy status: %d", resp.StatusCode)
	}

	return nil
}

// ListBeacons retrieves all active beacons
func (c *C2Client) ListBeacons() ([]Beacon, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/beacons")
	if err != nil {
		return nil, fmt.Errorf("failed to list beacons: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Beacons []Beacon `json:"beacons"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Beacons, nil
}

// GetBeacon retrieves a specific beacon by ID
func (c *C2Client) GetBeacon(beaconID string) (*Beacon, error) {
	beacons, err := c.ListBeacons()
	if err != nil {
		return nil, err
	}

	for _, beacon := range beacons {
		if beacon.ID == beaconID {
			return &beacon, nil
		}
	}

	return nil, fmt.Errorf("beacon not found: %s", beaconID)
}

// ActiveBeacons retrieves beacons active within the last N minutes
func (c *C2Client) ActiveBeacons(minutes int) ([]Beacon, error) {
	beacons, err := c.ListBeacons()
	if err != nil {
		return nil, err
	}

	var active []Beacon
	threshold := time.Now().Add(-time.Duration(minutes) * time.Minute)

	for _, beacon := range beacons {
		if beacon.LastSeen.After(threshold) {
			active = append(active, beacon)
		}
	}

	return active, nil
}

// SubmitCommand submits a command to a beacon
func (c *C2Client) SubmitCommand(beaconID, command string) (string, error) {
	payload := map[string]string{
		"beacon_id": beaconID,
		"command":   command,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/api/commands",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("failed to submit command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		CommandID string `json:"command_id"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.CommandID, nil
}

// GetOutput retrieves command output for a beacon
func (c *C2Client) GetOutput(beaconID string, commandID string) ([]Output, error) {
	url := fmt.Sprintf("%s/api/output/%s", c.BaseURL, beaconID)
	if commandID != "" {
		url += fmt.Sprintf("?command_id=%s", commandID)
	}

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get output: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Outputs []Output `json:"outputs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Outputs, nil
}

// ExecuteAndWait executes a command and waits for output
func (c *C2Client) ExecuteAndWait(beaconID, command string, timeout time.Duration) (string, error) {
	// Submit command
	commandID, err := c.SubmitCommand(beaconID, command)
	if err != nil {
		return "", err
	}

	// Poll for output
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			outputs, err := c.GetOutput(beaconID, commandID)
			if err != nil {
				return "", err
			}

			if len(outputs) > 0 {
				return outputs[0].Output, nil
			}

			if time.Now().After(deadline) {
				return "", fmt.Errorf("timeout waiting for output")
			}

		case <-time.After(timeout):
			return "", fmt.Errorf("timeout waiting for output")
		}
	}
}

// Example usage
func main() {
	// Create client
	client := NewC2Client("https://localhost:8443")

	// Health check
	if err := client.HealthCheck(); err != nil {
		fmt.Printf("Health check failed: %v\n", err)
	} else {
		fmt.Println("✓ Teamserver is healthy")
	}

	// List beacons
	beacons, err := client.ListBeacons()
	if err != nil {
		fmt.Printf("Error listing beacons: %v\n", err)
		return
	}

	fmt.Printf("Active Beacons: %d\n", len(beacons))
	for _, beacon := range beacons {
		fmt.Printf("  %s - %s@%s (%s)\n",
			beacon.ID, beacon.Username, beacon.Hostname, beacon.OS)
	}

	// Submit command (if beacons exist)
	if len(beacons) > 0 {
		beaconID := beacons[0].ID
		fmt.Printf("\nExecuting command on %s...\n", beaconID)

		output, err := client.ExecuteAndWait(beaconID, "whoami", 5*time.Minute)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Output:\n%s\n", output)
		}
	}
}
