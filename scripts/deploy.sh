#!/bin/bash
#
# Deployment Script
# Automated deployment for development and production environments
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

# Default environment
ENVIRONMENT="${1:-development}"

# Validate environment
if [[ "$ENVIRONMENT" != "development" && "$ENVIRONMENT" != "production" ]]; then
    print_error "Invalid environment: $ENVIRONMENT"
    echo "Usage: $0 [development|production]"
    exit 1
fi

print_header "Deploying C2 Framework - $ENVIRONMENT"

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker not installed"
        exit 1
    fi
    print_success "Docker installed"

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose not installed"
        exit 1
    fi
    print_success "Docker Compose installed"

    # Check Docker daemon
    if ! docker info &> /dev/null; then
        print_error "Docker daemon not running"
        exit 1
    fi
    print_success "Docker daemon running"
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

# Build components
build_components() {
    print_header "Building Components"

    if [ "$ENVIRONMENT" == "production" ]; then
        print_info "Building with production optimizations..."
        docker-compose build --no-cache
    else
        print_info "Building with development settings..."
        docker-compose build
    fi

    print_success "Build complete"
}

# Configure environment
configure_environment() {
    print_header "Configuring Environment"

    # Create .env if it doesn't exist
    if [ ! -f .env ]; then
        print_info "Creating .env file..."
        cp .env.example .env 2>/dev/null || cat > .env << EOF
C2_ENV=$ENVIRONMENT
DATABASE_URL=postgres://c2user:c2password@database:5432/c2_database
SERVER_ADDR=0.0.0.0:8443
RUST_LOG=debug
EOF
        print_success ".env file created"
    else
        print_info ".env file already exists"
    fi

    # Update environment in .env
    if grep -q "^C2_ENV=" .env; then
        sed -i.bak "s/^C2_ENV=.*/C2_ENV=$ENVIRONMENT/" .env
        rm .env.bak
    else
        echo "C2_ENV=$ENVIRONMENT" >> .env
    fi

    print_success "Environment configured: $ENVIRONMENT"
}

# Start services
start_services() {
    print_header "Starting Services"

    if [ "$ENVIRONMENT" == "production" ]; then
        print_info "Starting in production mode..."
        docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d 2>/dev/null || docker-compose up -d
    else
        print_info "Starting in development mode..."
        docker-compose up -d
    fi

    print_success "Services started"
}

# Wait for services
wait_for_services() {
    print_header "Waiting for Services"

    print_info "Waiting for database..."
    MAX_RETRIES=30
    RETRY_COUNT=0

    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if docker-compose ps database | grep -q "healthy\|Up"; then
            print_success "Database ready"
            break
        fi
        RETRY_COUNT=$((RETRY_COUNT + 1))
        echo -n "."
        sleep 2
    done

    if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
        print_error "Database failed to start"
        docker-compose logs database
        exit 1
    fi

    print_info "Waiting for teamserver..."
    sleep 10

    if docker-compose ps teamserver | grep -q "Up"; then
        print_success "Teamserver ready"
    else
        print_error "Teamserver failed to start"
        docker-compose logs teamserver
        exit 1
    fi
}

# Run health checks
run_health_checks() {
    print_header "Running Health Checks"

    # Check if healthcheck script exists
    if [ -f scripts/healthcheck.sh ]; then
        bash scripts/healthcheck.sh
    else
        print_info "Health check script not found, running basic checks..."

        # Basic service check
        if docker-compose ps | grep -q "teamserver.*Up"; then
            print_success "Teamserver is running"
        else
            print_error "Teamserver is not running"
        fi

        if docker-compose ps | grep -q "database.*Up"; then
            print_success "Database is running"
        else
            print_error "Database is not running"
        fi
    fi
}

# Print deployment info
print_deployment_info() {
    print_header "Deployment Complete"

    echo "Environment: $ENVIRONMENT"
    echo ""
    echo "Services:"
    docker-compose ps
    echo ""
    echo "Access points:"
    echo "  - Teamserver: https://localhost:8443"
    echo "  - Database:   localhost:5432"
    echo "  - Controller: docker-compose exec controller python main.py"
    echo ""
    echo "Useful commands:"
    echo "  - View logs:       docker-compose logs -f"
    echo "  - List beacons:    docker-compose exec controller python main.py beacons"
    echo "  - Stop services:   docker-compose down"
    echo "  - Health check:    ./scripts/healthcheck.sh"
    echo ""

    if [ "$ENVIRONMENT" == "production" ]; then
        echo -e "${YELLOW}PRODUCTION NOTES:${NC}"
        echo "  - Ensure firewall is configured"
        echo "  - Use valid TLS certificates"
        echo "  - Change default passwords"
        echo "  - Enable monitoring"
        echo "  - Set up backups"
        echo ""
    fi

    print_success "Deployment successful!"
}

# Main execution
main() {
    # Navigate to project root
    if [ -f docker-compose.yml ]; then
        print_info "Using current directory"
    elif [ -f ../docker-compose.yml ]; then
        cd ..
        print_info "Changed to project root"
    else
        print_error "docker-compose.yml not found"
        exit 1
    fi

    check_prerequisites
    stop_existing
    build_components
    configure_environment
    start_services
    wait_for_services
    run_health_checks
    print_deployment_info
}

# Run
main
