#!/bin/bash
set -euo pipefail  # Add this for better error handling

# Configuration
ALERT_EMAIL="admin@example.com"
ALERT_THRESHOLD=300  # seconds
LOG_FILE="./scripts/migrations/logs/migration_monitor.log"
ALERT_FILE="./scripts/migrations/logs/alerts.log"

# Create log directory if it doesn't exist
mkdir -p "$(dirname "$LOG_FILE")"
mkdir -p "$(dirname "$ALERT_FILE")"

# Source environment variables
source .env 2>/dev/null || true

# Add timeout for commands
TIMEOUT=10

# Function to log messages
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Function to send alerts
alert() {
    local message="$1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ALERT: $message" | tee -a "$ALERT_FILE"
    echo "$message" | mail -s "Database Migration Alert" "$ALERT_EMAIL"
}

# Function to check migration status
check_migration_status() {
    local version=$(docker-compose run --rm migrate version 2>/dev/null)
    local status=$(docker-compose run --rm migrate status 2>/dev/null)
    
    log "Current migration version: $version"
    log "Migration status: $status"
    
    # Check for failed migrations
    if echo "$status" | grep -q "error"; then
        alert "Migration error detected: $status"
    fi
}

# Function to monitor migration performance
monitor_migration_performance() {
    local start_time=$(date +%s)
    
    # Run migration with timing
    { time docker-compose run --rm migrate up 1; } 2>&1 | tee -a "$LOG_FILE"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $duration -gt $ALERT_THRESHOLD ]; then
        alert "Migration took longer than expected: ${duration}s"
    fi
}

# Function to check database health
check_database_health() {
    local health=$(docker-compose exec postgres psql -U user -d digital_discovery -c "SELECT version();" 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        alert "Database health check failed"
    else
        log "Database health check passed"
    fi
}

# Function to check disk space
check_disk_space() {
    local backup_dir="./scripts/migrations/backups"
    local space=$(df -h "$backup_dir" | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$space" -gt 90 ]; then
        alert "Backup directory running low on space: ${space}%"
    fi
}

# Add connection check
check_connection() {
    if ! psql -h localhost -U "${POSTGRES_USER:-user}" -d "${POSTGRES_DB:-digital_discovery}" -c "SELECT 1" &>/dev/null; then
        log "ERROR: Cannot connect to database"
        exit 1
    fi
}

# Main monitoring loop
main() {
    log "Starting migration monitoring"
    
    while true; do
        check_migration_status
        check_database_health
        check_disk_space
        
        # Wait for 5 minutes before next check
        sleep 300
    done
}

# Run main function
main 