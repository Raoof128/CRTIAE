# API Client Examples

This directory contains examples of building custom controllers and tools using the Red Team C2 Framework REST API.

## Overview

The teamserver exposes a REST API for all operations:
- Listing beacons
- Submitting commands
- Retrieving output
- Managing infrastructure

These examples show how to build custom clients in different languages.

## Examples

### 1. Python Client Library (`python_client.py`)

Full-featured Python client with object-oriented interface.

**Features**:
- Type hints and dataclasses
- Automatic retry and error handling
- Execute-and-wait convenience method
- SSL verification control

**Usage**:

```python
from python_client import C2Client

# Create client
client = C2Client("https://teamserver:8443")

# List beacons
beacons = client.beacons.list()
for beacon in beacons:
    print(beacon)

# Submit command
result = client.commands.submit(beacon_id, "whoami")
print(f"Command ID: {result['command_id']}")

# Execute and wait for output
output = client.commands.execute_and_wait(beacon_id, "ls -la", timeout=300)
print(output)
```

**Command-line interface**:

```bash
# List beacons
python3 python_client.py --server https://localhost:8443 list

# Execute command
python3 python_client.py --server https://localhost:8443 exec --beacon BEACON_ID --command "whoami"

# Execute and wait for output
python3 python_client.py --server https://localhost:8443 exec --beacon BEACON_ID --command "whoami" --wait

# Health check
python3 python_client.py --server https://localhost:8443 health
```

### 2. Go Client Library (`go_client.go`)

Idiomatic Go client library with strong typing.

**Features**:
- Idiomatic Go error handling
- Configurable timeouts
- TLS configuration
- Concurrent-safe

**Usage**:

```go
package main

import (
    "fmt"
    "time"
    "github.com/Raoof128/red-team-c2/examples/api_client"
)

func main() {
    // Create client
    client := NewC2Client("https://teamserver:8443")

    // List beacons
    beacons, err := client.ListBeacons()
    if err != nil {
        panic(err)
    }

    // Submit command
    commandID, err := client.SubmitCommand(beaconID, "whoami")

    // Execute and wait
    output, err := client.ExecuteAndWait(beaconID, "ls -la", 5*time.Minute)
    fmt.Println(output)
}
```

**Build and run**:

```bash
go run go_client.go
```

### 3. Web Dashboard (`web_dashboard.html`)

Simple web-based interface for monitoring beacons.

**Features**:
- Real-time beacon list
- Command submission
- Output retrieval
- Auto-refresh

**Usage**:

```bash
# Serve with Python
python3 -m http.server 8000

# Open in browser
open http://localhost:8000/web_dashboard.html
```

## API Reference

### Base URL

```
https://teamserver:8443
```

### Endpoints

#### List Beacons

```http
GET /api/beacons
```

Response:
```json
{
  "beacons": [
    {
      "id": "beacon-123",
      "hostname": "target-host",
      "username": "user",
      "os": "linux",
      "ip": "192.168.1.100",
      "first_seen": "2024-01-15T10:00:00Z",
      "last_seen": "2024-01-15T10:30:00Z",
      "status": "active"
    }
  ]
}
```

#### Submit Command

```http
POST /api/commands
Content-Type: application/json

{
  "beacon_id": "beacon-123",
  "command": "whoami"
}
```

Response:
```json
{
  "command_id": "cmd-456",
  "status": "queued"
}
```

#### Get Output

```http
GET /api/output/{beacon_id}
GET /api/output/{beacon_id}?command_id={command_id}
```

Response:
```json
{
  "outputs": [
    {
      "command_id": "cmd-456",
      "beacon_id": "beacon-123",
      "output": "command output here",
      "timestamp": "2024-01-15T10:31:00Z"
    }
  ]
}
```

## Integration Examples

### Automated Scanning

```python
from python_client import C2Client
import time

client = C2Client("https://teamserver:8443")

# Get all active beacons
beacons = client.beacons.active(minutes=5)

# Run command on all beacons
for beacon in beacons:
    print(f"Scanning {beacon.hostname}...")
    client.commands.submit(beacon.id, "nmap -sV localhost")
    time.sleep(2)  # Rate limiting

# Collect results
time.sleep(300)  # Wait for scans
for beacon in beacons:
    outputs = client.commands.get_output(beacon.id)
    for output in outputs:
        print(f"\n[{beacon.hostname}]:\n{output.output}")
```

### Monitoring Script

```python
from python_client import C2Client
import time

client = C2Client("https://teamserver:8443")

while True:
    beacons = client.beacons.active(minutes=5)
    print(f"Active beacons: {len(beacons)}")

    for beacon in beacons:
        print(f"  {beacon.hostname} - {beacon.last_seen}")

    time.sleep(60)  # Check every minute
```

### Go Automation

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    client := NewC2Client("https://teamserver:8443")

    // Monitor beacons
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        beacons, _ := client.ActiveBeacons(5)
        fmt.Printf("Active beacons: %d\n", len(beacons))

        // Run periodic tasks
        for _, beacon := range beacons {
            client.SubmitCommand(beacon.ID, "ps aux")
        }
    }
}
```

## Error Handling

### Python

```python
from python_client import C2Client
import requests

client = C2Client("https://teamserver:8443")

try:
    beacons = client.beacons.list()
except requests.Timeout:
    print("Request timed out")
except requests.ConnectionError:
    print("Connection failed")
except Exception as e:
    print(f"Error: {e}")
```

### Go

```go
beacons, err := client.ListBeacons()
if err != nil {
    log.Printf("Failed to list beacons: %v", err)
    return
}
```

## Best Practices

1. **Error Handling**: Always handle network errors gracefully
2. **Timeouts**: Set appropriate timeouts for long-running operations
3. **Rate Limiting**: Don't overwhelm the teamserver
4. **Logging**: Log all operations for audit trail
5. **Security**: Validate all input, use TLS in production

## Next Steps

- Build custom automation scripts
- Create monitoring dashboards
- Integrate with other security tools
- Develop custom command modules

## Support

- **API Documentation**: `API.md`
- **Architecture**: `ARCHITECTURE.md`
- **Issues**: https://github.com/Raoof128/CRTIAE/issues
