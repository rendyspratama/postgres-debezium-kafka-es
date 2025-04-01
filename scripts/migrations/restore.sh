#!/bin/bash

# Set backup directory
BACKUP_DIR="./scripts/migrations/backups"

# Check if backups exist
if [ ! -d "$BACKUP_DIR" ] || [ -z "$(ls -A $BACKUP_DIR/*.sql 2>/dev/null)" ]; then
    echo "No backups found in $BACKUP_DIR"
    exit 1
fi

# List available backups
echo "Available backups:"
ls -1t "$BACKUP_DIR"/*.sql | while read backup; do
    timestamp=$(basename "$backup" .sql | sed 's/backup_//')
    echo "  $timestamp"
    if [ -f "$backup.meta" ]; then
        echo "    Details:"
        cat "$backup.meta" | sed 's/^/      /'
    fi
done

# Ask which backup to restore
echo -n "Enter backup timestamp to restore (YYYYMMDD_HHMMSS): "
read TIMESTAMP

BACKUP_FILE="$BACKUP_DIR/backup_$TIMESTAMP.sql"

if [ ! -f "$BACKUP_FILE" ]; then
    echo "Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Confirm restoration
echo "WARNING: This will overwrite the current database!"
echo -n "Are you sure you want to restore from $BACKUP_FILE? [y/N]: "
read CONFIRM

if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    echo "Restoration cancelled"
    exit 0
fi

# Create temporary backup before restore
echo "Creating safety backup..."
TMP_BACKUP="$BACKUP_DIR/pre_restore_$(date +%Y%m%d_%H%M%S).sql"
docker-compose exec -T postgres pg_dump -U user digital_discovery > "$TMP_BACKUP"

# Restore database
echo "Restoring database from $BACKUP_FILE..."
cat "$BACKUP_FILE" | docker-compose exec -T postgres psql -U user digital_discovery

if [ $? -eq 0 ]; then
    echo "Database restored successfully"
    echo "Pre-restore backup saved as: $TMP_BACKUP"
else
    echo "Error restoring database"
    echo "Your database may be in an inconsistent state."
    echo "A backup was automatically created at: $TMP_BACKUP"
    exit 1
fi 