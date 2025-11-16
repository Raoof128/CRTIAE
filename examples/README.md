# Red Team C2 Framework - Code Examples

This directory contains practical code examples demonstrating how to use and extend the Red Team C2 Framework.

## Directory Structure

```
examples/
├── basic_deployment/       # Simple deployment example
├── custom_module/          # Creating custom implant modules
├── api_client/             # API client library examples
└── integration_tests/      # End-to-end integration tests
```

## Examples Overview

### 1. Basic Deployment

**Directory**: `basic_deployment/`

Demonstrates the simplest way to deploy the C2 framework for testing:
- Docker Compose deployment
- Single implant connection
- Basic command execution workflow

**Use Case**: Learning the framework, quick testing

**Files**:
- `deploy.sh` - Automated deployment script
- `test_connection.sh` - Verify implant connectivity
- `README.md` - Step-by-step guide

### 2. Custom Module

**Directory**: `custom_module/`

Shows how to add custom functionality to the implant:
- Creating new command modules
- Integrating with implant architecture
- Testing custom modules

**Use Case**: Extending implant capabilities for specific engagements

**Files**:
- `screenshot.go` - Custom screenshot module
- `keylogger.go` - Custom keylogging module (educational)
- `module_template.go` - Template for new modules
- `README.md` - Development guide

### 3. API Client

**Directory**: `api_client/`

Examples of building custom controllers using the REST API:
- Python API client library
- Go API client library
- Web dashboard example

**Use Case**: Building custom interfaces, automation, integration

**Files**:
- `python_client.py` - Full-featured Python client
- `go_client.go` - Go client implementation
- `web_dashboard.html` - Simple web interface
- `README.md` - API usage guide

### 4. Integration Tests

**Directory**: `integration_tests/`

End-to-end testing scenarios:
- Full deployment testing
- Multi-implant scenarios
- Network simulation

**Use Case**: Validation, CI/CD, quality assurance

**Files**:
- `test_full_stack.py` - Complete stack test
- `test_multi_beacon.py` - Multiple beacon simulation
- `docker-compose.test.yml` - Test environment
- `README.md` - Testing guide

## Quick Start

### Run Basic Deployment Example

```bash
cd examples/basic_deployment
./deploy.sh
```

### Test Custom Module

```bash
cd examples/custom_module
go build -o custom_implant .
./custom_implant --test
```

### Use API Client

```bash
cd examples/api_client
python3 python_client.py --server http://localhost:8443 list-beacons
```

## Learning Path

**For Beginners**:
1. Start with `basic_deployment/` - Learn deployment and basic operations
2. Explore `api_client/` - Understand the API and build simple tools
3. Try `custom_module/` - Extend the implant with new features

**For Advanced Users**:
1. Review `custom_module/` - Build sophisticated modules
2. Study `integration_tests/` - Set up automated testing
3. Create your own examples - Contribute back to the project

## Security Considerations

All examples are for **authorized security testing only**:

- **Get Written Permission**: Never deploy without explicit authorization
- **Stay in Scope**: Only test authorized systems
- **Document Everything**: Log all activities
- **Clean Up**: Remove all implants and artifacts after testing

## Contributing Examples

We welcome contributions of new examples! Please:

1. Follow the directory structure pattern
2. Include comprehensive README.md
3. Add comments explaining key concepts
4. Test thoroughly before submitting
5. Document security considerations

See `CONTRIBUTING.md` for full guidelines.

## Example Difficulty Levels

- **Beginner**: `basic_deployment/`
- **Intermediate**: `api_client/`, `integration_tests/`
- **Advanced**: `custom_module/`

## Support

- **Documentation**: See `/docs` directory
- **Issues**: https://github.com/Raoof128/CRTIAE/issues
- **Security**: See `SECURITY_POLICY.md`

## Legal Notice

This framework is intended for:
- Authorized penetration testing
- Security research
- Educational purposes
- Red team exercises with proper authorization

**Unauthorized use is illegal and unethical.**

---

**Remember**: Always act ethically and legally when using these examples.
