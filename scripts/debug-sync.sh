#!/bin/bash

echo "Checking Kafka messages..."
docker-compose exec kafka kafka-console-consumer \
  --bootstrap-server kafka:29092 \
  --topic development-digital-discovery-categories-2025-04 \
  --from-beginning \
  --property print.key=true \
  --property key.separator=: \
  --max-messages 5

echo -e "\nChecking Elasticsearch connector status..."
curl -s http://localhost:8083/connectors/elasticsearch-sink/status | jq .

echo -e "\nChecking Elasticsearch connector tasks logs..."
curl -s "http://localhost:8083/connectors/elasticsearch-sink/tasks/0/log" | tail -n 20

echo -e "\nChecking document in Elasticsearch..."
curl -s "localhost:9200/development-digital-discovery-categories-2025-04/_search" | jq . 