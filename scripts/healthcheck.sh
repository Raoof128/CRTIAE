#!/bin/bash
#
# Health Check Script
# Verifies that all C2 Framework services are running correctly
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

# Track overall health
HEALTH_STATUS=0

# Check Docker services
check_docker_services() {
    print_header "Checking Docker Services"

    # Check if docker-compose is running
    if ! docker-compose ps &> /dev/null; then
        print_error "Docker Compose not running or not available"
        HEALTH_STATUS=1
        return
    fi

    # Check each service
    SERVICES=("database" "teamserver" "controller")

    for service in "${SERVICES[@]}"; do
        if docker-compose ps | grep -q "$service.*Up"; then
            print_success "$service is running"
        else
            print_error "$service is not running"
            HEALTH_STATUS=1
        fi
    done
}

# Check database connectivity
check_database() {
    print_header "Checking Database"

    # Check if database container is healthy
    if docker-compose ps database | grep -q "healthy"; then
        print_success "Database is healthy"
    elif docker-compose ps database | grep -q "Up"; then
        print_info "Database is running but not yet healthy"
    else
        print_error "Database is not running"
        HEALTH_STATUS=1
        return
    fi

    # Test database connection
    if docker-compose exec -T database psql -U c2user -d c2_database -c "SELECT 1;" &> /dev/null; then
        print_success "Database connection successful"
    else
        print_error "Database connection failed"
        HEALTH_STATUS=1
    fi

    # Check database size
    DB_SIZE=$(docker-compose exec -T database psql -U c2user -d c2_database -t -c "SELECT pg_size_pretty(pg_database_size('c2_database'));" 2>/dev/null | xargs || echo "unknown")
    print_info "Database size: $DB_SIZE"
}

# Check teamserver API
check_teamserver() {
    print_header "Checking Teamserver"

    # Check if teamserver is running
    if ! docker-compose ps teamserver | grep -q "Up"; then
        print_error "Teamserver is not running"
        HEALTH_STATUS=1
        return
    fi

    print_success "Teamserver container is running"

    # Check if port is listening
    if netstat -tlnp 2>/dev/null | grep -q ":8443" || lsof -i :8443 &> /dev/null; then
        print_success "Teamserver is listening on port 8443"
    else
        print_error "Teamserver is not listening on port 8443"
        HEALTH_STATUS=1
        return
    fi

    # Test HTTP endpoint
    if curl -k -s -o /dev/null -w "%{http_code}" https://localhost:8443/health 2>/dev/null | grep -q "200\|404"; then
        print_success "Teamserver HTTPS endpoint is accessible"
    else
        print_error "Teamserver HTTPS endpoint is not accessible"
        HEALTH_STATUS=1
    fi

    # Test API endpoint
    if curl -k -s https://localhost:8443/api/beacons &> /dev/null; then
        print_success "Teamserver API is responding"
    else
        print_error "Teamserver API is not responding"
        HEALTH_STATUS=1
    fi
}

# Check beacons
check_beacons() {
    print_header "Checking Beacons"

    # Get beacon count
    BEACON_COUNT=$(docker-compose exec -T controller python main.py beacons 2>/dev/null | grep -c "beacon-" || echo "0")

    if [ "$BEACON_COUNT" -gt "0" ]; then
        print_success "Active beacons: $BEACON_COUNT"
    else
        print_info "No active beacons (this may be normal)"
    fi

    # Check for stale beacons (last seen > 10 minutes ago)
    # This would require parsing the beacon list, skip for now
    print_info "Note: Manual beacon health check recommended"
}

# Check disk space
check_disk_space() {
    print_header "Checking Disk Space"

    # Check available disk space
    DISK_USAGE=$(df -h . | awk 'NR==2 {print $5}' | tr -d '%')

    if [ "$DISK_USAGE" -lt 80 ]; then
        print_success "Disk usage: ${DISK_USAGE}%"
    elif [ "$DISK_USAGE" -lt 90 ]; then
        print_info "Disk usage: ${DISK_USAGE}% (warning threshold)"
    else
        print_error "Disk usage: ${DISK_USAGE}% (critical)"
        HEALTH_STATUS=1
    fi
}

# Check memory usage
check_memory() {
    print_header "Checking Memory Usage"

    if command -v free &> /dev/null; then
        MEMORY_USAGE=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100}')
        print_info "Memory usage: ${MEMORY_USAGE}%"
    else
        print_info "Memory check not available on this system"
    fi
}

# Check logs for errors
check_logs() {
    print_header "Checking Recent Logs"

    # Check teamserver logs for errors
    ERROR_COUNT=$(docker-compose logs --tail=100 teamserver 2>/dev/null | grep -ci "error" || echo "0")

    if [ "$ERROR_COUNT" -eq 0 ]; then
        print_success "No recent errors in teamserver logs"
    else
        print_info "Found $ERROR_COUNT error entries in recent logs"
        print_info "Review with: docker-compose logs teamserver"
    fi
}

# Check network connectivity
check_network() {
    print_header "Checking Network"

    # Check if ports are accessible
    PORTS=("5432" "8443")

    for port in "${PORTS[@]}"; do
        if nc -z localhost $port 2>/dev/null || timeout 1 bash -c "cat < /dev/null > /dev/tcp/localhost/$port" 2>/dev/null; then
            print_success "Port $port is accessible"
        else
            print_error "Port $port is not accessible"
            HEALTH_STATUS=1
        fi
    done
}

# Print summary
print_summary() {
    print_header "Health Check Summary"

    if [ $HEALTH_STATUS -eq 0 ]; then
        echo -e "${GREEN}✓ All health checks passed${NC}"
        echo ""
        echo "System Status:"
        docker-compose ps
        exit 0
    else
        echo -e "${RED}✗ Health check failed${NC}"
        echo ""
        echo "Failed checks detected. Review the errors above."
        echo ""
        echo "Troubleshooting:"
        echo "  - View logs: docker-compose logs"
        echo "  - Restart services: docker-compose restart"
        echo "  - Check configuration: cat .env"
        echo ""
        exit 1
    fi
}

# Main execution
main() {
    print_header "C2 Framework Health Check"

    # Navigate to project root
    if [ -f docker-compose.yml ]; then
        print_info "Found docker-compose.yml in current directory"
    elif [ -f ../docker-compose.yml ]; then
        cd ..
        print_info "Changed to parent directory"
    else
        print_error "docker-compose.yml not found"
        exit 1
    fi

    check_docker_services
    check_database
    check_teamserver
    check_beacons
    check_disk_space
    check_memory
    check_logs
    check_network
    print_summary
}

# Run
main
