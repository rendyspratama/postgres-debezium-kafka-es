#!/bin/bash
set -euo pipefail

# Source environment variables
source .env 2>/dev/null || true

# Create backups directory if it doesn't exist
BACKUP_DIR="./scripts/migrations/backups"
mkdir -p "$BACKUP_DIR"

# Generate timestamp for backup file
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/backup_$TIMESTAMP.sql"

# Create backup
echo "Creating database backup..."
docker-compose exec -T postgres pg_dump -U user digital_discovery > "$BACKUP_FILE"

if [ $? -eq 0 ]; then
    echo "Backup created successfully: $BACKUP_FILE"
    
    # Create metadata file
    META_FILE="$BACKUP_FILE.meta"
    echo "Timestamp: $TIMESTAMP" > "$META_FILE"
    echo "Database: digital_discovery" >> "$META_FILE"
    echo "Version: $(docker-compose run --rm migrate version 2>/dev/null)" >> "$META_FILE"
    
    # Cleanup old backups (keep last 5)
    cd "$BACKUP_DIR" && ls -t *.sql | tail -n +6 | xargs -r rm
    cd "$BACKUP_DIR" && ls -t *.meta | tail -n +6 | xargs -r rm
else
    echo "Error creating backup"
    rm -f "$BACKUP_FILE"
    exit 1
fi

# Add disk space check
check_disk_space() {
    local required_space=1000000  # 1GB in KB
    local available_space=$(df -k . | awk 'NR==2 {print $4}')
    if (( available_space < required_space )); then
        echo "ERROR: Not enough disk space for backup"
        exit 1
    fi
}

# Add backup rotation
rotate_backups() {
    local max_backups=5
    local backup_dir="./scripts/migrations/backups"
    local count=$(ls -1 "$backup_dir"/*.sql 2>/dev/null | wc -l)
    if (( count > max_backups )); then
        ls -1t "$backup_dir"/*.sql | tail -n +$(( max_backups + 1 )) | xargs rm -f
    fi
} 