#!/bin/bash

echo "Checking Debezium (Source Connector)..."
curl -s "http://localhost:8084/connector-plugins" | jq '.[].class'

echo -e "\nChecking Kafka Connect (Sink Connector)..."
curl -s "http://localhost:8083/connector-plugins" | jq '.[].class'

echo -e "\nChecking Source Connector Status..."
curl -s "http://localhost:8084/connectors/postgres-source/status" | jq .

echo -e "\nChecking Sink Connector Status..."
curl -s "http://localhost:8083/connectors/elasticsearch-sink/status" | jq .

echo -e "\nChecking Kafka Topics..."
docker-compose exec kafka kafka-topics --bootstrap-server kafka:29092 --list 