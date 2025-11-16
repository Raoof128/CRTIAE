# Contributing to Red Team C2 Framework

First off, thank you for considering contributing to the Red Team C2 Framework! This project is an educational cybersecurity portfolio demonstrating purple team methodologies.

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## How Can I Contribute?

### Reporting Bugs

**Before submitting a bug report:**
- Check the [DEBUG_REPORT.md](DEBUG_REPORT.md) for known issues
- Search existing [GitHub Issues](https://github.com/Raoof128/CRTIAE/issues)
- Ensure you're using the latest version

**How to submit a good bug report:**

```markdown
**Environment:**
- OS: [e.g., Ubuntu 22.04, Windows 11, macOS 14]
- Component: [Implant, Teamserver, Controller]
- Version: [e.g., v1.0.0]
- Deployment: [Docker, Manual]

**Description:**
Clear and concise description of the bug.

**Steps to Reproduce:**
1. Step one
2. Step two
3. See error

**Expected Behavior:**
What you expected to happen.

**Actual Behavior:**
What actually happened.

**Logs:**
```
Paste relevant logs here
```

**Additional Context:**
Any other information about the problem.
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful**
- **List any alternative solutions** you've considered

### Pull Requests

**Process:**

1. **Fork the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/CRTIAE.git
   cd CRTIAE
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Follow the coding standards (see below)
   - Add tests for new functionality
   - Update documentation as needed

4. **Test your changes**
   ```bash
   # Go tests
   cd implant && go test ./...

   # Rust tests
   cd teamserver && cargo test

   # Python tests
   cd controller && python -m pytest
   ```

5. **Commit with conventional commits**
   ```bash
   git commit -m "feat: add new beaconing protocol"
   git commit -m "fix: resolve DNS resolver timeout"
   git commit -m "docs: update API documentation"
   ```

6. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

**PR Guidelines:**
- Clearly describe the problem and solution
- Include relevant issue numbers
- Update CHANGELOG.md
- Ensure CI/CD passes
- Include blue team detection methods for offensive features

## Coding Standards

### Go (Implant)

```go
// Use gofmt for formatting
gofmt -w .

// Follow effective Go guidelines
// - Clear variable names
// - Error handling on every error
// - Documentation comments for exported functions

// Example:
// ExecuteCommand runs a shell command with timeout protection
// and returns the output or error. Implements cross-platform
// command execution for Windows, Linux, and macOS.
func (ce *CommandExecutor) ExecuteCommand(cmd string) (*CommandResult, error) {
    // Implementation
}
```

**Standards:**
- Use `gofmt` for formatting
- Run `go vet` before committing
- Add comments for exported functions
- Handle all errors explicitly
- Use meaningful variable names

### Rust (Teamserver)

```rust
// Use rustfmt for formatting
cargo fmt

// Follow Rust style guide
// - Use snake_case for functions
// - Use CamelCase for types
// - Document public APIs

/// Register a new beacon in the database
///
/// # Arguments
/// * `beacon_id` - Unique identifier for the beacon
/// * `hostname` - Target system hostname
///
/// # Returns
/// Result indicating success or database error
pub async fn register_beacon(
    &self,
    beacon_id: &str,
    hostname: Option<String>,
) -> Result<(), sqlx::Error> {
    // Implementation
}
```

**Standards:**
- Use `cargo fmt` for formatting
- Run `cargo clippy` before committing
- Document public APIs with `///`
- Use `Result` for error handling
- Prefer `&str` over `String` when possible

### Python (Controller)

```python
# Use black for formatting
black controller/

# Follow PEP 8
# - Use snake_case for functions
# - Use CamelCase for classes
# - 4-space indentation
# - Maximum line length: 88 (black default)

def execute_command(beacon_id: str, command: str) -> bool:
    """
    Execute a command on the specified beacon.

    Args:
        beacon_id: Unique identifier of the target beacon
        command: Shell command to execute

    Returns:
        True if command was successfully queued, False otherwise

    Raises:
        RequestException: If teamserver is unreachable
    """
    # Implementation
```

**Standards:**
- Use `black` for formatting
- Use type hints
- Add docstrings (Google style)
- Follow PEP 8
- Use `pylint` or `flake8`

### SQL

```sql
-- Use uppercase for SQL keywords
-- Use snake_case for table/column names
-- Include comments for complex queries
-- Use parameterized queries (never string concatenation)

CREATE TABLE beacons (
    id VARCHAR(255) PRIMARY KEY,
    hostname VARCHAR(255),
    -- Timestamp of first beacon
    first_seen TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for frequent queries
CREATE INDEX idx_beacons_last_seen ON beacons(last_seen DESC);
```

## Purple Team Requirement

**CRITICAL**: All offensive capabilities MUST include corresponding defensive documentation.

### When adding offensive features:

1. **Document detection methods** in `blue-team-guide/`
2. **Add YARA rules** if applicable
3. **Provide network signatures** (Suricata/Snort)
4. **Include behavioral analytics** (Splunk queries)
5. **Document mitigations** for defenders

**Example:**

```go
// New feature: SMB beaconing
// Located: implant/comms/smb.go

// BLUE TEAM DETECTION:
// - Monitor: Unusual SMB connections to external IPs
// - Signature: Named pipe patterns (\\.\pipe\beacon_*)
// - Behavioral: Periodic SMB traffic from non-file-sharing processes
// - Mitigation: Firewall rules blocking outbound SMB
```

And corresponding documentation:

```markdown
<!-- blue-team-guide/DETECTION_SIGNATURES.md -->

### SMB Beaconing Detection

**Indicators:**
- Named pipe connections with suspicious patterns
- Periodic SMB traffic (60-180s intervals)
- SMB from unexpected processes

**YARA Rule:**
```yara
rule C2_SMB_Beacon {
    strings:
        $pipe = "\\\\.\\pipe\\beacon_" wide
    condition:
        $pipe
}
```
```

## Testing Requirements

### Unit Tests

All new code must include unit tests:

```go
// implant/comms/crypto_test.go
func TestEncryptDecrypt(t *testing.T) {
    ce, _ := NewCryptoEngine(nil)
    plaintext := []byte("test data")

    encrypted, err := ce.Encrypt(plaintext)
    if err != nil {
        t.Fatalf("Encryption failed: %v", err)
    }

    decrypted, err := ce.Decrypt(encrypted)
    if err != nil {
        t.Fatalf("Decryption failed: %v", err)
    }

    if !bytes.Equal(plaintext, decrypted) {
        t.Errorf("Expected %v, got %v", plaintext, decrypted)
    }
}
```

### Integration Tests

Complex features should include integration tests:

```bash
# tests/integration/test_beacon.sh
#!/bin/bash

# Start teamserver
docker-compose up -d teamserver

# Deploy test implant
./build/implant --server http://localhost:8443 --interval 5 &
IMPLANT_PID=$!

# Wait for beacon
sleep 10

# Verify beacon registered
BEACONS=$(curl -s http://localhost:8443/api/operator/beacons | jq '.count')
if [ "$BEACONS" -eq 1 ]; then
    echo "✓ Beacon registered"
else
    echo "✗ Beacon not found"
    exit 1
fi

# Cleanup
kill $IMPLANT_PID
docker-compose down
```

## Documentation Requirements

### Code Comments

- **Go**: Document all exported functions
- **Rust**: Document all public APIs
- **Python**: Docstrings for all functions/classes

### README Updates

If your change affects usage, update:
- README.md (main documentation)
- QUICKSTART.md (if deployment changes)
- API.md (if API changes)

### CHANGELOG

Add entry to CHANGELOG.md:

```markdown
## [Unreleased]

### Added
- SMB beaconing protocol for C2 communication (#42)

### Fixed
- DNS resolver timeout on slow networks (#45)

### Changed
- Increased default beacon interval to 120s (#47)
```

## Commit Message Guidelines

Follow [Conventional Commits](https://www.conventionalcommits.org/):

**Format:**
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, no logic change)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(implant): add SMB beaconing protocol

Implements Named Pipe-based C2 communication for Windows targets.
Includes jitter and error handling. Blue team detection methods
documented in DETECTION_SIGNATURES.md.

Closes #42
```

```
fix(teamserver): resolve database connection pool exhaustion

Connection pool was not releasing connections on error.
Added proper error handling and connection cleanup.

Fixes #45
```

## Development Setup

### Prerequisites

- Go 1.21+
- Rust 1.75+
- Python 3.11+
- Docker & Docker Compose
- PostgreSQL 15+

### Initial Setup

```bash
# Clone repository
git clone https://github.com/Raoof128/CRTIAE.git
cd CRTIAE

# Install dependencies
cd implant && go mod download && cd ..
cd teamserver && cargo build && cd ..
cd controller && pip install -r requirements.txt && cd ..

# Run tests
./scripts/run_tests.sh
```

### Pre-commit Checklist

Before submitting a PR:

- [ ] Code compiles without errors
- [ ] Tests pass (`./scripts/run_tests.sh`)
- [ ] Code formatted (`gofmt`, `cargo fmt`, `black`)
- [ ] Linting passes (`go vet`, `cargo clippy`, `pylint`)
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Blue team documentation added (if offensive feature)
- [ ] Commit messages follow conventional commits

## Security Contributions

### Vulnerability Disclosure

**DO NOT** open public issues for security vulnerabilities.

Instead:
1. Email security concerns privately (see SECURITY_POLICY.md)
2. Allow 90 days for patch development
3. Coordinate public disclosure

### Security Features

When adding security features:
- Document threat model
- Explain defense-in-depth approach
- Provide blue team perspective
- Include detection/mitigation methods

## Community

### Communication

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: Questions, ideas, showcase
- **Pull Requests**: Code contributions

### Recognition

Contributors are recognized in:
- CONTRIBUTORS.md (all contributors)
- Git commit history
- Release notes (significant contributions)

## License

By contributing, you agree that your contributions will be licensed under the MIT License (see LICENSE).

## Questions?

- Check [README.md](README.md) for project overview
- Read [QUICKSTART.md](QUICKSTART.md) for deployment
- Review [SECURITY_POLICY.md](SECURITY_POLICY.md) for responsible use
- Open a [GitHub Discussion](https://github.com/Raoof128/CRTIAE/discussions) for questions

---

**Thank you for contributing to making cybersecurity education better!**
