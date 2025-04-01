#!/bin/bash

echo "🔍 Debugging Elasticsearch Sink Connector"

# 1. Check sink connector status
echo "Checking sink connector status..."
curl -s "http://localhost:8083/connectors/elasticsearch-sink/status" | jq .

# 2. Check sink connector config
echo -e "\nChecking sink connector config..."
curl -s "http://localhost:8083/connectors/elasticsearch-sink/config" | jq .

# 3. Check Kafka Connect logs
echo -e "\nChecking Kafka Connect logs for errors..."
docker-compose logs kafka-connect | grep -i error

# 4. Check Elasticsearch indices
echo -e "\nChecking Elasticsearch indices..."
curl -s "localhost:9200/_cat/indices?v" 