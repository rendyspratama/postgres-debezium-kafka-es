#!/bin/bash

echo "Cleaning up connectors..."

echo "Deleting source connector from Debezium..."
curl -X DELETE http://localhost:8084/connectors/postgres-source 2>/dev/null || true

echo "Deleting sink connector from Kafka Connect..."
curl -X DELETE http://localhost:8083/connectors/elasticsearch-sink 2>/dev/null || true

echo "Waiting for connectors to be deleted..."
sleep 2

echo "Verifying cleanup..."
echo "Checking Debezium connectors:"
curl -s "http://localhost:8084/connectors" | jq .

echo "Checking Kafka Connect connectors:"
curl -s "http://localhost:8083/connectors" | jq .

echo "Cleanup complete!" 