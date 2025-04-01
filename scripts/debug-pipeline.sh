#!/bin/bash

echo "🔍 Debugging Pipeline"

echo -e "\n1. Checking Source Connector Status:"
curl -s http://localhost:8084/connectors/postgres-source/status | jq .

echo -e "\n2. Checking Sink Connector Status:"
curl -s http://localhost:8083/connectors/elasticsearch-sink/status | jq .

echo -e "\n3. Checking Kafka Topic Messages:"
docker-compose exec kafka kafka-console-consumer \
  --bootstrap-server kafka:29092 \
  --topic development-digital-discovery-categories-2025-04 \
  --from-beginning \
  --max-messages 1

echo -e "\n4. Checking Elasticsearch Indices:"
curl -s "localhost:9200/_cat/indices?v"

echo -e "\n5. Checking Elasticsearch Index Mapping:"
curl -s "localhost:9200/development-digital-discovery-categories-2025-04/_mapping" | jq .

echo -e "\n6. Checking Sink Connector Tasks Logs:"
curl -s "http://localhost:8083/connectors/elasticsearch-sink/tasks/0/log" | tail -n 20

echo -e "\n7. Checking if index exists in Elasticsearch:"
curl -s -I "localhost:9200/development-digital-discovery-categories-2025-04"

echo -e "\n8. Checking Sink Connector Config:"
curl -s http://localhost:8083/connectors/elasticsearch-sink/config | jq . 