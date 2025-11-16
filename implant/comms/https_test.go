package comms

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewHTTPSChannel tests the creation of HTTPS channel
func TestNewHTTPSChannel(t *testing.T) {
	config := &HTTPSConfig{
		ServerURL:      "https://example.com",
		BeaconInterval: 60,
		Jitter:         30,
		UserAgent:      "Mozilla/5.0",
		Insecure:       true,
	}

	channel, err := NewHTTPSChannel(config, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTPS channel: %v", err)
	}

	if channel == nil {
		t.Fatal("HTTPS channel is nil")
	}

	if channel.config.ServerURL != config.ServerURL {
		t.Errorf("Server URL mismatch: got %s, want %s", channel.config.ServerURL, config.ServerURL)
	}
}

// TestCalculateJitter tests jitter calculation
func TestCalculateJitter(t *testing.T) {
	config := &HTTPSConfig{
		ServerURL:      "https://example.com",
		BeaconInterval: 60,
		Jitter:         50, // 50% jitter
		UserAgent:      "Mozilla/5.0",
		Insecure:       true,
	}

	channel, err := NewHTTPSChannel(config, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTPS channel: %v", err)
	}

	// Calculate jitter multiple times to test randomness
	for i := 0; i < 10; i++ {
		interval := channel.calculateJitter()

		// With 50% jitter, interval should be between 30s and 90s (60s ± 50%)
		minInterval := 30 * time.Second
		maxInterval := 90 * time.Second

		if interval < minInterval || interval > maxInterval {
			t.Errorf("Jitter out of range: got %v, want between %v and %v",
				interval, minInterval, maxInterval)
		}
	}
}

// TestBuildRequest tests HTTP request building
func TestBuildRequest(t *testing.T) {
	config := &HTTPSConfig{
		ServerURL:      "https://example.com",
		BeaconInterval: 60,
		Jitter:         30,
		UserAgent:      "Custom-Agent/1.0",
		Insecure:       true,
	}

	channel, err := NewHTTPSChannel(config, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTPS channel: %v", err)
	}

	beaconID := "test-beacon-123"
	data := []byte("test-payload")

	req, err := channel.buildRequest("POST", "/api/checkin", beaconID, data)
	if err != nil {
		t.Fatalf("Failed to build request: %v", err)
	}

	if req.Method != "POST" {
		t.Errorf("Method mismatch: got %s, want POST", req.Method)
	}

	if req.Header.Get("User-Agent") != config.UserAgent {
		t.Errorf("User-Agent mismatch: got %s, want %s",
			req.Header.Get("User-Agent"), config.UserAgent)
	}

	if req.Header.Get("X-Beacon-ID") != beaconID {
		t.Errorf("Beacon ID header mismatch: got %s, want %s",
			req.Header.Get("X-Beacon-ID"), beaconID)
	}

	if req.Header.Get("Content-Type") != "application/octet-stream" {
		t.Errorf("Content-Type mismatch: got %s, want application/octet-stream",
			req.Header.Get("Content-Type"))
	}
}

// TestCheckinRequest tests the checkin request
func TestCheckinRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/checkin" {
			t.Errorf("Path mismatch: got %s, want /api/checkin", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Method mismatch: got %s, want POST", r.Method)
		}

		beaconID := r.Header.Get("X-Beacon-ID")
		if beaconID == "" {
			t.Error("Missing X-Beacon-ID header")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","commands":[]}`))
	}))
	defer server.Close()

	config := &HTTPSConfig{
		ServerURL:      server.URL,
		BeaconInterval: 60,
		Jitter:         30,
		UserAgent:      "Test-Agent/1.0",
		Insecure:       true,
	}

	channel, err := NewHTTPSChannel(config, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTPS channel: %v", err)
	}

	beaconID := "test-beacon-456"
	metadata := map[string]string{
		"hostname": "test-host",
		"username": "test-user",
		"os":       "linux",
	}

	commands, err := channel.Checkin(beaconID, metadata)
	if err != nil {
		t.Fatalf("Checkin failed: %v", err)
	}

	if commands == nil {
		t.Error("Commands slice is nil")
	}
}

// TestSubmitOutput tests output submission
func TestSubmitOutput(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/output" {
			t.Errorf("Path mismatch: got %s, want /api/output", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Method mismatch: got %s, want POST", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	config := &HTTPSConfig{
		ServerURL:      server.URL,
		BeaconInterval: 60,
		Jitter:         30,
		UserAgent:      "Test-Agent/1.0",
		Insecure:       true,
	}

	channel, err := NewHTTPSChannel(config, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTPS channel: %v", err)
	}

	beaconID := "test-beacon-789"
	commandID := "cmd-123"
	output := "command output here"

	err = channel.SubmitOutput(beaconID, commandID, output)
	if err != nil {
		t.Fatalf("SubmitOutput failed: %v", err)
	}
}

// TestHTTPError tests error handling for HTTP errors
func TestHTTPError(t *testing.T) {
	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server error"))
	}))
	defer server.Close()

	config := &HTTPSConfig{
		ServerURL:      server.URL,
		BeaconInterval: 60,
		Jitter:         30,
		UserAgent:      "Test-Agent/1.0",
		Insecure:       true,
	}

	channel, err := NewHTTPSChannel(config, nil)
	if err != nil {
		t.Fatalf("Failed to create HTTPS channel: %v", err)
	}

	beaconID := "test-beacon-error"
	metadata := map[string]string{
		"hostname": "test-host",
	}

	_, err = channel.Checkin(beaconID, metadata)
	if err == nil {
		t.Error("Expected error for HTTP 500, got nil")
	}
}

// TestInvalidServerURL tests error handling for invalid URLs
func TestInvalidServerURL(t *testing.T) {
	config := &HTTPSConfig{
		ServerURL:      "://invalid-url",
		BeaconInterval: 60,
		Jitter:         30,
		UserAgent:      "Test-Agent/1.0",
		Insecure:       true,
	}

	_, err := NewHTTPSChannel(config, nil)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

// BenchmarkCheckin benchmarks the checkin operation
func BenchmarkCheckin(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","commands":[]}`))
	}))
	defer server.Close()

	config := &HTTPSConfig{
		ServerURL:      server.URL,
		BeaconInterval: 60,
		Jitter:         30,
		UserAgent:      "Bench-Agent/1.0",
		Insecure:       true,
	}

	channel, _ := NewHTTPSChannel(config, nil)
	beaconID := "bench-beacon"
	metadata := map[string]string{
		"hostname": "bench-host",
		"username": "bench-user",
		"os":       "linux",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		channel.Checkin(beaconID, metadata)
	}
}

// BenchmarkSubmitOutput benchmarks output submission
func BenchmarkSubmitOutput(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	config := &HTTPSConfig{
		ServerURL:      server.URL,
		BeaconInterval: 60,
		Jitter:         30,
		UserAgent:      "Bench-Agent/1.0",
		Insecure:       true,
	}

	channel, _ := NewHTTPSChannel(config, nil)
	beaconID := "bench-beacon"
	commandID := "bench-cmd"
	output := "benchmark output data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		channel.SubmitOutput(beaconID, commandID, output)
	}
}
