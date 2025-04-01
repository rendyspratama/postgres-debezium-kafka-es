#!/bin/bash
set -euo pipefail

# Source environment variables
source .env 2>/dev/null || true

# Configuration
SEED_DIR="./scripts/migrations/seeds"
LOG_FILE="./scripts/migrations/logs/seed.log"

# Create directories
mkdir -p "$SEED_DIR"
mkdir -p "$(dirname "$LOG_FILE")"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Add validation function
validate_seed_file() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        echo "ERROR: Seed file '$file' not found"
        exit 1
    fi
    
    if ! psql -h localhost -U "${POSTGRES_USER:-user}" -d "${POSTGRES_DB:-digital_discovery}" -f "$file" --dry-run &>/dev/null; then
        echo "ERROR: Invalid SQL in seed file '$file'"
        exit 1
    fi
}

# Function to apply seed data
apply_seed() {
    local seed_file="$1"
    local seed_name=$(basename "$seed_file" .sql)
    
    log "Applying seed: $seed_name"
    
    # Run seed SQL
    psql -h localhost -U "${POSTGRES_USER:-user}" -d "${POSTGRES_DB:-digital_discovery}" <<EOF
    BEGIN;
    \i $seed_file
    COMMIT;
EOF
    
    if [ $? -eq 0 ]; then
        log "Seed applied successfully: $seed_name"
    else
        log "Failed to apply seed: $seed_name"
        return 1
    fi
}

# Function to remove seed data
remove_seed() {
    local seed_file="$1"
    local seed_name=$(basename "$seed_file" .sql)
    
    log "Removing seed: $seed_name"
    
    # Extract table names from seed file
    local tables=$(grep -o "INSERT INTO [a-zA-Z_]*" "$seed_file" | cut -d' ' -f3 | sort -u)
    
    # Remove data from tables
    for table in $tables; do
        docker-compose exec postgres psql -U user -d digital_discovery -c "DELETE FROM $table;"
        log "Removed data from table: $table"
    done
}

# Function to list available seeds
list_seeds() {
    echo "Available seeds:"
    for seed_file in "$SEED_DIR"/*.sql; do
        if [ -f "$seed_file" ]; then
            echo "  $(basename "$seed_file")"
        fi
    done
}

# Function to create new seed
create_seed() {
    local name="$1"
    local seed_file="$SEED_DIR/${name}.sql"
    
    if [ -f "$seed_file" ]; then
        log "Seed file already exists: $seed_file"
        return 1
    fi
    
    # Create seed file with template
    cat > "$seed_file" << EOF
-- Seed data for $name
-- Created: $(date '+%Y-%m-%d %H:%M:%S')

BEGIN;

-- Add your seed data here
-- Example:
-- INSERT INTO table_name (column1, column2) VALUES ('value1', 'value2');

COMMIT;
EOF
    
    log "Created new seed file: $seed_file"
}

# Main function
main() {
    case "$1" in
        "apply")
            if [ -z "$2" ]; then
                echo "Usage: $0 apply <seed_file>"
                list_seeds
                exit 1
            fi
            apply_seed "$SEED_DIR/$2"
            ;;
        "remove")
            if [ -z "$2" ]; then
                echo "Usage: $0 remove <seed_file>"
                list_seeds
                exit 1
            fi
            remove_seed "$SEED_DIR/$2"
            ;;
        "create")
            if [ -z "$2" ]; then
                echo "Usage: $0 create <seed_name>"
                exit 1
            fi
            create_seed "$2"
            ;;
        "list")
            list_seeds
            ;;
        *)
            echo "Usage: $0 {apply|remove|create|list} [seed_file|seed_name]"
            exit 1
            ;;
    esac
}

# Run main function with arguments
main "$@" 