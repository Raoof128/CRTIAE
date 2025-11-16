#!/bin/bash
#
# Setup Script - One-command environment setup
# Installs dependencies and configures the Red Team C2 Framework
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_info() { echo -e "${YELLOW}→${NC} $1"; }
print_header() { echo -e "\n${BLUE}==== $1 ====${NC}\n"; }

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [ -f /etc/debian_version ]; then
            OS="debian"
        elif [ -f /etc/redhat-release ]; then
            OS="redhat"
        else
            OS="linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
    else
        print_error "Unsupported OS: $OSTYPE"
        exit 1
    fi
    print_success "Detected OS: $OS"
}

# Check Go installation
check_go() {
    print_header "Checking Go Installation"

    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        print_success "Go installed: $GO_VERSION"

        # Check version >= 1.21
        MAJOR=$(echo $GO_VERSION | sed 's/go//' | cut -d. -f1)
        MINOR=$(echo $GO_VERSION | sed 's/go//' | cut -d. -f2)

        if [ "$MAJOR" -ge 1 ] && [ "$MINOR" -ge 21 ]; then
            print_success "Go version is sufficient (>= 1.21)"
        else
            print_error "Go version too old. Need >= 1.21"
            install_go
        fi
    else
        print_info "Go not found"
        install_go
    fi
}

# Install Go
install_go() {
    print_info "Installing Go 1.21..."

    if [ "$OS" == "macos" ]; then
        brew install go@1.21
    elif [ "$OS" == "debian" ]; then
        wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
        rm go1.21.0.linux-amd64.tar.gz
        export PATH=$PATH:/usr/local/go/bin
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    else
        print_error "Please install Go manually from https://go.dev/dl/"
        exit 1
    fi

    print_success "Go installed"
}

# Check Rust installation
check_rust() {
    print_header "Checking Rust Installation"

    if command -v cargo &> /dev/null; then
        RUST_VERSION=$(rustc --version)
        print_success "Rust installed: $RUST_VERSION"
    else
        print_info "Rust not found"
        install_rust
    fi
}

# Install Rust
install_rust() {
    print_info "Installing Rust..."

    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source "$HOME/.cargo/env"

    print_success "Rust installed"
}

# Check Python installation
check_python() {
    print_header "Checking Python Installation"

    if command -v python3 &> /dev/null; then
        PYTHON_VERSION=$(python3 --version)
        print_success "Python installed: $PYTHON_VERSION"

        # Check version >= 3.11
        MAJOR=$(python3 -c 'import sys; print(sys.version_info.major)')
        MINOR=$(python3 -c 'import sys; print(sys.version_info.minor)')

        if [ "$MAJOR" -ge 3 ] && [ "$MINOR" -ge 11 ]; then
            print_success "Python version is sufficient (>= 3.11)"
        else
            print_error "Python version too old. Need >= 3.11"
        fi
    else
        print_error "Python3 not found. Please install Python 3.11+"
        exit 1
    fi
}

# Check Docker installation
check_docker() {
    print_header "Checking Docker Installation"

    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version)
        print_success "Docker installed: $DOCKER_VERSION"

        # Check if Docker daemon is running
        if docker info &> /dev/null; then
            print_success "Docker daemon is running"
        else
            print_error "Docker daemon is not running"
            print_info "Start Docker and try again"
            exit 1
        fi
    else
        print_info "Docker not found"
        install_docker
    fi

    # Check Docker Compose
    if command -v docker-compose &> /dev/null; then
        COMPOSE_VERSION=$(docker-compose --version)
        print_success "Docker Compose installed: $COMPOSE_VERSION"
    else
        print_error "Docker Compose not found"
        install_docker_compose
    fi
}

# Install Docker
install_docker() {
    print_info "Installing Docker..."

    if [ "$OS" == "macos" ]; then
        print_error "Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
        exit 1
    elif [ "$OS" == "debian" ]; then
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        rm get-docker.sh
        sudo usermod -aG docker $USER
        print_success "Docker installed. Please log out and back in for group changes"
    else
        print_error "Please install Docker manually from https://docs.docker.com/get-docker/"
        exit 1
    fi
}

# Install Docker Compose
install_docker_compose() {
    print_info "Installing Docker Compose..."

    if [ "$OS" == "macos" ]; then
        brew install docker-compose
    else
        sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
    fi

    print_success "Docker Compose installed"
}

# Install PostgreSQL
check_postgresql() {
    print_header "Checking PostgreSQL"

    if command -v psql &> /dev/null; then
        PSQL_VERSION=$(psql --version)
        print_success "PostgreSQL client installed: $PSQL_VERSION"
    else
        print_info "PostgreSQL client not found (optional for local development)"
    fi
}

# Setup project dependencies
setup_dependencies() {
    print_header "Setting Up Project Dependencies"

    # Go dependencies
    print_info "Installing Go dependencies..."
    cd implant
    go mod download
    cd ..
    print_success "Go dependencies installed"

    # Rust dependencies
    print_info "Fetching Rust dependencies..."
    cd teamserver
    cargo fetch
    cd ..
    print_success "Rust dependencies fetched"

    # Python dependencies
    print_info "Installing Python dependencies..."
    cd controller
    python3 -m pip install --user -r requirements.txt
    cd ..
    print_success "Python dependencies installed"
}

# Initialize database
init_database() {
    print_header "Initializing Database"

    print_info "Database will be initialized on first Docker deployment"
    print_success "Database configuration ready"
}

# Create environment file
create_env_file() {
    print_header "Creating Environment File"

    if [ -f .env ]; then
        print_info ".env file already exists"
    else
        cat > .env << EOF
# Red Team C2 Framework - Environment Configuration

# Environment (development/production)
C2_ENV=development

# Database
DATABASE_URL=postgres://c2user:c2password@localhost:5432/c2_database
POSTGRES_USER=c2user
POSTGRES_PASSWORD=c2password
POSTGRES_DB=c2_database

# Teamserver
SERVER_ADDR=0.0.0.0:8443
RUST_LOG=debug

# Controller
C2_SERVER_URL=http://localhost:8443

# Security
TLS_CERT_PATH=/etc/c2/certs/server.crt
TLS_KEY_PATH=/etc/c2/certs/server.key
EOF
        print_success ".env file created"
        print_info "Review and customize .env file for your environment"
    fi
}

# Print summary
print_summary() {
    print_header "Setup Complete"

    echo "All prerequisites are installed:"
    echo "  ✓ Go $(go version | awk '{print $3}')"
    echo "  ✓ Rust $(rustc --version | awk '{print $2}')"
    echo "  ✓ Python $(python3 --version | awk '{print $2}')"
    echo "  ✓ Docker $(docker --version | awk '{print $3}' | tr -d ',')"
    echo ""
    echo "Next steps:"
    echo ""
    echo "1. Review configuration:"
    echo "   cat .env"
    echo ""
    echo "2. Build components:"
    echo "   make build"
    echo ""
    echo "3. Deploy with Docker:"
    echo "   docker-compose up -d"
    echo ""
    echo "4. Or use deployment script:"
    echo "   ./scripts/deploy.sh development"
    echo ""
    print_success "Happy hacking!"
}

# Main execution
main() {
    print_header "Red Team C2 Framework - Setup"

    detect_os
    check_go
    check_rust
    check_python
    check_docker
    check_postgresql
    setup_dependencies
    init_database
    create_env_file
    print_summary
}

# Run
main
