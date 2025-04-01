#!/bin/bash

# Function to make API calls
test_api() {
    local name=$1
    local status=$2
    
    # Create JSON with proper quotes
    json="{\"name\":\"$name\",\"status\":$status}"
    
    echo "Testing with payload: $json"
    
    curl -X POST http://localhost:8081/api/v1/categories \
    -H "Content-Type: application/json" \
    -d "$json"
}

# Test cases
echo "Test Case 1: Basic category"
test_api "Test Category 1" 1

echo -e "\nTest Case 2: Another category"
test_api "Test Category 2" 1

# If you need to include description
test_api_with_description() {
    local name=$1
    local description=$2
    local status=$3
    
    json="{\"name\":\"$name\",\"description\":\"$description\",\"status\":$status}"
    
    echo "Testing with payload: $json"
    
    curl -X POST http://localhost:8081/api/v1/categories \
    -H "Content-Type: application/json" \
    -d "$json"
}

echo -e "\nTest Case 3: Category with description"
test_api_with_description "Test Category 3" "This is a test description" 1 