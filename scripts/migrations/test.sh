#!/bin/bash
set -euo pipefail

# Source environment variables
source .env 2>/dev/null || true

# Configuration
TEST_DB="digital_discovery_test"
TEST_DIR="./scripts/migrations/tests"
LOG_FILE="./scripts/migrations/logs/test.log"

# Create directories
mkdir -p "$TEST_DIR"
mkdir -p "$(dirname "$LOG_FILE")"

# Function to log messages
log() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local green='\033[0;32m'
    local blue='\033[0;34m'
    local red='\033[0;31m'
    local nc='\033[0m' # No Color

    case "$2" in
        "INFO")
            echo -e "${blue}[${timestamp}]${nc} $1" | tee -a "$LOG_FILE"
            ;;
        "SUCCESS")
            echo -e "${green}[${timestamp}]${nc} ✓ $1" | tee -a "$LOG_FILE"
            ;;
        "ERROR")
            echo -e "${red}[${timestamp}]${nc} ✗ $1" | tee -a "$LOG_FILE"
            ;;
        *)
            echo -e "[${timestamp}] $1" | tee -a "$LOG_FILE"
            ;;
    esac
}

# Function to create test database
setup_test_db() {
    log "Starting test database setup..." "INFO"
    
    # Drop test database if exists
    log "Dropping existing test database..." "INFO"
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-user} -d postgres -c "DROP DATABASE IF EXISTS $TEST_DB;" >/dev/null 2>&1
    
    # Create fresh test database
    log "Creating fresh test database..." "INFO"
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-user} -d postgres -c "CREATE DATABASE $TEST_DB;" >/dev/null 2>&1
    
    # Run migrations on test database
    log "Running migrations..." "INFO"
    docker-compose run --rm migrate -path=/migrations -database "postgres://${POSTGRES_USER:-user}:${POSTGRES_PASSWORD:-password}@postgres:5432/$TEST_DB?sslmode=disable" up
    
    log "Test database setup complete" "SUCCESS"
}

# Function to run test cases
run_test_cases() {
    local test_file="$1"
    local test_name=$(basename "$test_file" .sql)
    
    log "Running test: ${test_name}" "INFO"
    
    if docker-compose exec -T postgres psql -U ${POSTGRES_USER:-user} -d "$TEST_DB" -f "$test_file" >/dev/null 2>&1; then
        log "Test passed: ${test_name}" "SUCCESS"
        return 0
    else
        log "Test failed: ${test_name}" "ERROR"
        return 1
    fi
}

# Function to verify data integrity
verify_data_integrity() {
    local test_file="$1"
    local test_name=$(basename "$test_file" .sql)
    
    log "Verifying data integrity for: $test_name"
    
    # Run verification queries
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-user} -d "$TEST_DB" -f "$test_file"
    
    if [ $? -eq 0 ]; then
        log "Data integrity check passed: $test_name"
    else
        log "Data integrity check failed: $test_name"
        return 1
    fi
}

# Function to run rollback tests
test_rollback() {
    local test_file="$1"
    local test_name=$(basename "$test_file" .sql)
    
    log "Testing rollback for: $test_name"
    
    # Run migration
    docker-compose run --rm migrate -path=/migrations -database "postgres://${POSTGRES_USER:-user}:${POSTGRES_PASSWORD:-password}@postgres:5432/$TEST_DB?sslmode=disable" up 1
    
    # Verify data
    verify_data_integrity "$test_file"
    
    # Run rollback
    docker-compose run --rm migrate -path=/migrations -database "postgres://${POSTGRES_USER:-user}:${POSTGRES_PASSWORD:-password}@postgres:5432/$TEST_DB?sslmode=disable" down 1
    
    # Verify rollback
    verify_data_integrity "$test_file"
    
    if [ $? -eq 0 ]; then
        log "Rollback test passed: $test_name"
    else
        log "Rollback test failed: $test_name"
        return 1
    fi
}

# Add test setup function
setup() {
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-user} -d "$TEST_DB" -c "BEGIN;"
}

# Add test teardown function
teardown() {
    docker-compose exec -T postgres psql -U ${POSTGRES_USER:-user} -d "$TEST_DB" -c "ROLLBACK;"
}

# Add test reporting
report_test() {
    local test_name="$1"
    local status="$2"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] Test '$test_name': $status"
}

# Main test function
main() {
    log "=== Starting Migration Tests ===" "INFO"
    echo
    
    setup_test_db
    echo
    
    local test_count=0
    local pass_count=0
    
    for test_file in "$TEST_DIR"/*.sql; do
        if [ -f "$test_file" ]; then
            ((test_count++))
            setup
            if run_test_cases "$test_file"; then
                ((pass_count++))
            fi
            teardown
        fi
    done
    
    echo
    if [ $pass_count -eq $test_count ]; then
        log "All tests passed! ($pass_count/$test_count)" "SUCCESS"
    else
        log "Some tests failed. ($pass_count/$test_count passed)" "ERROR"
    fi
}

# Run main function
main 