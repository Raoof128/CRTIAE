#!/bin/bash
#
# Backup Script
# Creates backups of database and configuration files
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

# Configuration
BACKUP_DIR="${1:-./backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="c2_backup_${TIMESTAMP}"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"

# Create backup directory
create_backup_dir() {
    print_header "Preparing Backup"

    mkdir -p "$BACKUP_DIR"
    mkdir -p "$BACKUP_PATH"

    print_success "Backup directory created: $BACKUP_PATH"
}

# Backup database
backup_database() {
    print_header "Backing Up Database"

    print_info "Creating database dump..."

    docker-compose exec -T database pg_dump -U c2user -d c2_database > "${BACKUP_PATH}/database.sql"

    if [ -f "${BACKUP_PATH}/database.sql" ]; then
        DB_SIZE=$(du -h "${BACKUP_PATH}/database.sql" | awk '{print $1}')
        print_success "Database backed up (${DB_SIZE})"
    else
        print_error "Database backup failed"
        exit 1
    fi
}

# Backup configuration files
backup_config() {
    print_header "Backing Up Configuration"

    # Backup .env file
    if [ -f .env ]; then
        cp .env "${BACKUP_PATH}/.env"
        print_success ".env file backed up"
    else
        print_info ".env file not found (skipping)"
    fi

    # Backup docker-compose files
    cp docker-compose.yml "${BACKUP_PATH}/docker-compose.yml"
    print_success "docker-compose.yml backed up"

    # Backup any custom configs
    if [ -d config ]; then
        cp -r config "${BACKUP_PATH}/config"
        print_success "Config directory backed up"
    fi
}

# Backup logs (optional)
backup_logs() {
    print_header "Backing Up Logs"

    print_info "Exporting recent logs..."

    docker-compose logs --tail=1000 > "${BACKUP_PATH}/docker_logs.txt" 2>&1 || true

    if [ -f "${BACKUP_PATH}/docker_logs.txt" ]; then
        LOG_SIZE=$(du -h "${BACKUP_PATH}/docker_logs.txt" | awk '{print $1}')
        print_success "Logs backed up (${LOG_SIZE})"
    else
        print_info "No logs to backup"
    fi
}

# Create metadata
create_metadata() {
    print_header "Creating Metadata"

    cat > "${BACKUP_PATH}/backup_info.txt" << EOF
Red Team C2 Framework - Backup Information
==========================================

Backup Date: $(date)
Backup Name: ${BACKUP_NAME}
Hostname: $(hostname)
User: $(whoami)

Components:
- Database dump: database.sql
- Configuration: .env, docker-compose.yml
- Logs: docker_logs.txt

Restoration:
1. Extract backup archive
2. Copy configuration files to project directory
3. Restore database: psql -U c2user -d c2_database < database.sql
4. Restart services: docker-compose up -d

EOF

    print_success "Metadata created"
}

# Compress backup
compress_backup() {
    print_header "Compressing Backup"

    print_info "Creating archive..."

    cd "$BACKUP_DIR"
    tar -czf "${BACKUP_NAME}.tar.gz" "$BACKUP_NAME"

    if [ -f "${BACKUP_NAME}.tar.gz" ]; then
        ARCHIVE_SIZE=$(du -h "${BACKUP_NAME}.tar.gz" | awk '{print $1}')
        print_success "Backup compressed (${ARCHIVE_SIZE})"

        # Remove uncompressed directory
        rm -rf "$BACKUP_NAME"

        # Return to original directory
        cd - > /dev/null

        # Calculate checksum
        CHECKSUM=$(sha256sum "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz" | awk '{print $1}')
        echo "$CHECKSUM" > "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz.sha256"
        print_success "Checksum: ${CHECKSUM:0:16}..."
    else
        print_error "Compression failed"
        exit 1
    fi
}

# Cleanup old backups
cleanup_old_backups() {
    print_header "Cleaning Up Old Backups"

    # Keep only last 7 backups
    KEEP_COUNT=7

    BACKUP_COUNT=$(ls -1 "${BACKUP_DIR}"/c2_backup_*.tar.gz 2>/dev/null | wc -l || echo 0)

    if [ "$BACKUP_COUNT" -gt "$KEEP_COUNT" ]; then
        print_info "Found $BACKUP_COUNT backups, keeping $KEEP_COUNT most recent..."

        ls -1t "${BACKUP_DIR}"/c2_backup_*.tar.gz | tail -n +$((KEEP_COUNT + 1)) | while read backup; do
            rm -f "$backup" "${backup}.sha256"
            print_info "Removed old backup: $(basename $backup)"
        done

        print_success "Old backups cleaned up"
    else
        print_info "Backup count within limit ($BACKUP_COUNT/$KEEP_COUNT)"
    fi
}

# Print summary
print_summary() {
    print_header "Backup Complete"

    echo "Backup created successfully:"
    echo "  Location: ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
    echo "  Size:     $(du -h "${BACKUP_DIR}/${BACKUP_NAME}.tar.gz" | awk '{print $1}')"
    echo "  SHA256:   ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz.sha256"
    echo ""
    echo "To restore this backup:"
    echo "  ./scripts/restore.sh ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
    echo ""
    print_success "Backup complete!"
}

# Main execution
main() {
    print_header "C2 Framework Backup"

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

    # Check if services are running
    if ! docker-compose ps | grep -q "Up"; then
        print_error "Services are not running. Start with: docker-compose up -d"
        exit 1
    fi

    create_backup_dir
    backup_database
    backup_config
    backup_logs
    create_metadata
    compress_backup
    cleanup_old_backups
    print_summary
}

# Run
main
