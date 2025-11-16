# Red Team C2 Framework - API Documentation

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Beacon Endpoints](#beacon-endpoints)
4. [Operator Endpoints](#operator-endpoints)
5. [Data Models](#data-models)
6. [Error Handling](#error-handling)
7. [Rate Limiting](#rate-limiting)
8. [Examples](#examples)

---

## Overview

The Red Team C2 Framework provides a REST API for communication between implants, the teamserver, and operators. All endpoints return JSON responses and expect JSON payloads where applicable.

### Base URL

```
http://localhost:8443  # Development
https://c2.example.com:8443  # Production
```

### API Version

Current version: `v1`

All endpoints are currently unversioned. Future versions will use `/api/v2/` prefix.

### Content Type

All requests must include:
```
Content-Type: application/json
```

All responses will be:
```
Content-Type: application/json
```

---

## Authentication

### Current Implementation

**Status**: ⚠️ Not implemented in v1.0.0

All endpoints are currently **unprotected**. This is acceptable for:
- Isolated lab environments
- Docker deployments with network isolation
- Testing purposes

### Future Implementation (v1.1.0)

Will support:
- **API Keys**: Bearer token authentication
- **JWT**: JSON Web Tokens for operators
- **mTLS**: Mutual TLS for implants

**Example** (planned):
```http
Authorization: Bearer <api_key_or_jwt_token>
```

---

## Beacon Endpoints

Endpoints used by implants to communicate with the teamserver.

### Check-in

Register a new beacon or update an existing one.

**Endpoint**: `POST /api/beacon/{beacon_id}/checkin`

**Parameters**:
- `beacon_id` (path, required): Unique identifier for the beacon

**Request Body**:
```json
{
  "beacon_id": "abc123def456",
  "timestamp": 1700000000,
  "hostname": "target-host",
  "username": "user1",
  "os": "windows",
  "ip": "192.168.1.100",
  "data": "base64_encrypted_system_info",
  "metadata": {
    "version": "1.0.0",
    "arch": "amd64"
  }
}
```

**Response**: `200 OK`
```json
{
  "status": "success",
  "message": "Beacon registered",
  "beacon_id": "abc123def456"
}
```

**Errors**:
- `400 Bad Request`: Invalid payload
- `500 Internal Server Error`: Database error

**Example**:
```bash
curl -X POST http://localhost:8443/api/beacon/beacon123/checkin \
  -H "Content-Type: application/json" \
  -d '{
    "beacon_id": "beacon123",
    "timestamp": 1700000000,
    "data": "encrypted_data_here"
  }'
```

---

### Get Commands

Retrieve pending commands for a beacon.

**Endpoint**: `GET /api/commands/{beacon_id}`

**Parameters**:
- `beacon_id` (path, required): Beacon identifier

**Response**: `200 OK`
```json
[
  {
    "command_id": "cmd_789",
    "command_type": "exec",
    "command_data": "base64_encrypted_command",
    "metadata": {
      "priority": "normal"
    }
  },
  {
    "command_id": "cmd_790",
    "command_type": "sleep",
    "command_data": "encrypted_interval",
    "metadata": null
  }
]
```

**Empty response** (no pending commands):
```json
[]
```

**Side Effects**:
- Updates beacon's `last_seen` timestamp
- Marks retrieved commands as `executed = true`

**Example**:
```bash
curl http://localhost:8443/api/commands/beacon123
```

---

### Submit Output

Send command execution results back to teamserver.

**Endpoint**: `POST /api/beacon/{beacon_id}/output`

**Parameters**:
- `beacon_id` (path, required): Beacon identifier

**Request Body**:
```json
{
  "beacon_id": "beacon123",
  "timestamp": 1700000100,
  "data": "base64_encrypted_output",
  "metadata": {
    "command_id": "cmd_789",
    "status": "success",
    "exit_code": 0,
    "duration_ms": 1234
  }
}
```

**Response**: `200 OK`
```json
{
  "status": "success",
  "message": "Output stored"
}
```

**Example**:
```bash
curl -X POST http://localhost:8443/api/beacon/beacon123/output \
  -H "Content-Type: application/json" \
  -d '{
    "beacon_id": "beacon123",
    "timestamp": 1700000100,
    "data": "encrypted_output",
    "metadata": {
      "command_id": "cmd_789",
      "status": "success"
    }
  }'
```

---

## Operator Endpoints

Endpoints used by operators to manage beacons and issue commands.

### List Beacons

Retrieve all registered beacons.

**Endpoint**: `GET /api/operator/beacons`

**Query Parameters**:
- `status` (optional): Filter by status (`active`, `lost`, `terminated`)
- `limit` (optional): Maximum number of results (default: 100)
- `offset` (optional): Pagination offset (default: 0)

**Response**: `200 OK`
```json
{
  "count": 2,
  "beacons": [
    {
      "id": "beacon123",
      "hostname": "target-host-1",
      "username": "admin",
      "os": "windows",
      "ip": "192.168.1.100",
      "first_seen": "2024-11-16T10:00:00Z",
      "last_seen": "2024-11-16T10:05:00Z",
      "status": "active"
    },
    {
      "id": "beacon456",
      "hostname": "target-host-2",
      "username": "user",
      "os": "linux",
      "ip": "192.168.1.101",
      "first_seen": "2024-11-16T09:30:00Z",
      "last_seen": "2024-11-16T10:04:00Z",
      "status": "active"
    }
  ]
}
```

**Example**:
```bash
# List all beacons
curl http://localhost:8443/api/operator/beacons

# Filter by status
curl "http://localhost:8443/api/operator/beacons?status=active"

# Pagination
curl "http://localhost:8443/api/operator/beacons?limit=10&offset=20"
```

---

### Submit Command

Queue a command for execution by a beacon.

**Endpoint**: `POST /api/operator/command`

**Request Body**:
```json
{
  "beacon_id": "beacon123",
  "command_type": "exec",
  "command_data": "whoami"
}
```

**Command Types**:
- `exec`: Execute shell command
- `sleep`: Update beacon interval
- `exit`: Terminate beacon
- `upload`: Upload file to beacon (future)
- `download`: Download file from beacon (future)

**Response**: `200 OK`
```json
{
  "status": "success",
  "command_id": "cmd_789",
  "message": "Command queued"
}
```

**Example**:
```bash
curl -X POST http://localhost:8443/api/operator/command \
  -H "Content-Type: application/json" \
  -d '{
    "beacon_id": "beacon123",
    "command_type": "exec",
    "command_data": "whoami"
  }'
```

---

### Get Outputs

Retrieve command outputs from a beacon.

**Endpoint**: `GET /api/operator/outputs/{beacon_id}`

**Parameters**:
- `beacon_id` (path, required): Beacon identifier

**Query Parameters**:
- `limit` (optional): Maximum results (default: 100)
- `since` (optional): ISO 8601 timestamp (only outputs after this time)

**Response**: `200 OK`
```json
{
  "count": 2,
  "outputs": [
    {
      "id": "output_001",
      "beacon_id": "beacon123",
      "command_id": "cmd_789",
      "output_data": "username\\nDOMAIN\\admin",
      "timestamp": "2024-11-16T10:05:30Z",
      "status": "success"
    },
    {
      "id": "output_002",
      "beacon_id": "beacon123",
      "command_id": "cmd_790",
      "output_data": "Sleep interval updated to 120s",
      "timestamp": "2024-11-16T10:06:00Z",
      "status": "success"
    }
  ]
}
```

**Example**:
```bash
# Get all outputs
curl http://localhost:8443/api/operator/outputs/beacon123

# Get recent outputs
curl "http://localhost:8443/api/operator/outputs/beacon123?limit=10"

# Get outputs since timestamp
curl "http://localhost:8443/api/operator/outputs/beacon123?since=2024-11-16T10:00:00Z"
```

---

## Data Models

### Beacon

```json
{
  "id": "string (primary key)",
  "hostname": "string | null",
  "username": "string | null",
  "os": "string | null (windows, linux, darwin)",
  "ip": "string | null",
  "first_seen": "datetime (ISO 8601)",
  "last_seen": "datetime (ISO 8601)",
  "status": "string (active, lost, terminated)"
}
```

**Status Values**:
- `active`: Beaconing regularly
- `lost`: No check-in for >5 minutes
- `terminated`: Received exit command

### Command

```json
{
  "id": "string (UUID)",
  "beacon_id": "string (foreign key)",
  "command_type": "string",
  "command_data": "string (encrypted)",
  "created_at": "datetime (ISO 8601)",
  "executed": "boolean",
  "executed_at": "datetime (ISO 8601) | null"
}
```

### Output

```json
{
  "id": "string (UUID)",
  "beacon_id": "string (foreign key)",
  "command_id": "string (foreign key) | null",
  "output_data": "string (decrypted by teamserver)",
  "timestamp": "datetime (ISO 8601)",
  "status": "string (success, error)"
}
```

---

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional context (optional)"
  }
}
```

### HTTP Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful request |
| 400 | Bad Request | Invalid payload or parameters |
| 404 | Not Found | Beacon or resource not found |
| 500 | Internal Server Error | Database error or server failure |
| 503 | Service Unavailable | Database unavailable |

### Common Errors

**Invalid JSON**:
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "INVALID_JSON",
    "message": "Failed to parse JSON payload"
  }
}
```

**Beacon Not Found**:
```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "error": {
    "code": "BEACON_NOT_FOUND",
    "message": "Beacon with ID 'beacon999' does not exist"
  }
}
```

**Database Error**:
```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json

{
  "error": {
    "code": "DATABASE_ERROR",
    "message": "Failed to connect to database",
    "details": "Connection timeout after 30s"
  }
}
```

---

## Rate Limiting

### Current Implementation

**Status**: ⚠️ Not implemented in v1.0.0

No rate limiting is currently enforced.

### Future Implementation (v1.2.0)

Planned limits:
- **Beacons**: 1 request per 5 seconds per beacon_id
- **Operators**: 100 requests per minute per API key
- **Global**: 10,000 requests per hour

**Headers** (planned):
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1700000000
```

**Rate Limit Exceeded**:
```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Retry after 60 seconds.",
    "retry_after": 60
  }
}
```

---

## Examples

### Python Client Example

```python
import requests
import json

class C2Client:
    def __init__(self, base_url):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()

    def list_beacons(self):
        """List all active beacons"""
        response = self.session.get(f"{self.base_url}/api/operator/beacons")
        response.raise_for_status()
        return response.json()

    def submit_command(self, beacon_id, command_type, command_data):
        """Submit a command to a beacon"""
        payload = {
            "beacon_id": beacon_id,
            "command_type": command_type,
            "command_data": command_data
        }
        response = self.session.post(
            f"{self.base_url}/api/operator/command",
            json=payload
        )
        response.raise_for_status()
        return response.json()

    def get_outputs(self, beacon_id, limit=10):
        """Get command outputs from a beacon"""
        params = {"limit": limit}
        response = self.session.get(
            f"{self.base_url}/api/operator/outputs/{beacon_id}",
            params=params
        )
        response.raise_for_status()
        return response.json()

# Usage
client = C2Client("http://localhost:8443")

# List beacons
beacons = client.list_beacons()
print(f"Active beacons: {beacons['count']}")

# Execute command
result = client.submit_command("beacon123", "exec", "whoami")
print(f"Command queued: {result['command_id']}")

# Get outputs
outputs = client.get_outputs("beacon123")
for output in outputs['outputs']:
    print(f"Output: {output['output_data']}")
```

### cURL Examples

```bash
# Beacon check-in
curl -X POST http://localhost:8443/api/beacon/beacon123/checkin \
  -H "Content-Type: application/json" \
  -d '{"beacon_id": "beacon123", "timestamp": 1700000000}'

# Get commands (as beacon)
curl http://localhost:8443/api/commands/beacon123

# Submit output (as beacon)
curl -X POST http://localhost:8443/api/beacon/beacon123/output \
  -H "Content-Type: application/json" \
  -d '{
    "beacon_id": "beacon123",
    "timestamp": 1700000100,
    "data": "command_output",
    "metadata": {"command_id": "cmd_789", "status": "success"}
  }'

# List beacons (as operator)
curl http://localhost:8443/api/operator/beacons | jq

# Submit command (as operator)
curl -X POST http://localhost:8443/api/operator/command \
  -H "Content-Type: application/json" \
  -d '{
    "beacon_id": "beacon123",
    "command_type": "exec",
    "command_data": "whoami"
  }' | jq

# Get outputs (as operator)
curl http://localhost:8443/api/operator/outputs/beacon123 | jq
```

### Go Client Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type C2Client struct {
    BaseURL string
    Client  *http.Client
}

func (c *C2Client) ListBeacons() (map[string]interface{}, error) {
    resp, err := c.Client.Get(c.BaseURL + "/api/operator/beacons")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result, nil
}

func (c *C2Client) SubmitCommand(beaconID, cmdType, cmdData string) error {
    payload := map[string]string{
        "beacon_id":    beaconID,
        "command_type": cmdType,
        "command_data": cmdData,
    }

    jsonData, _ := json.Marshal(payload)
    resp, err := c.Client.Post(
        c.BaseURL+"/api/operator/command",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    return nil
}

func main() {
    client := &C2Client{
        BaseURL: "http://localhost:8443",
        Client:  &http.Client{},
    }

    // List beacons
    beacons, _ := client.ListBeacons()
    fmt.Printf("Beacons: %v\n", beacons)

    // Submit command
    err := client.SubmitCommand("beacon123", "exec", "whoami")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

---

## OpenAPI Specification

Future versions will include a full OpenAPI 3.0 specification for automatic client generation.

**Planned location**: `/api/docs/openapi.json`

---

## Changelog

### v1.0.0 (2024-11-16)
- Initial API implementation
- Beacon endpoints: check-in, get commands, submit output
- Operator endpoints: list beacons, submit command, get outputs
- JSON request/response format
- Basic error handling

### Future Versions

**v1.1.0** (Planned):
- Authentication (API keys, JWT)
- WebSocket support for real-time updates
- File upload/download endpoints

**v1.2.0** (Planned):
- Rate limiting
- API versioning (`/api/v2/`)
- Pagination improvements
- Filtering and sorting

**v1.3.0** (Planned):
- GraphQL endpoint
- Batch operations
- Async task status tracking

---

## Support

- **Issues**: https://github.com/Raoof128/CRTIAE/issues
- **Documentation**: https://github.com/Raoof128/CRTIAE
- **Security**: See SECURITY_POLICY.md

---

**API Version**: 1.0.0
**Last Updated**: 2024-11-16
**Maintainer**: Red Team C2 Project
