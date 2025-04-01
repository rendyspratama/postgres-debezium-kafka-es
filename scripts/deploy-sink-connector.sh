#!/bin/bash

echo "Waiting for Kafka Connect to be ready..."
until curl -s -f -o /dev/null http://localhost:8083/connectors; do
    echo "Waiting for Kafka Connect..."
    sleep 5
done

echo "Deleting existing sink connector (if any)..."
curl -X DELETE http://localhost:8083/connectors/elasticsearch-sink 2>/dev/null || true
sleep 2

echo "Deploying Elasticsearch sink connector..."
curl -X POST http://localhost:8083/connectors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "elasticsearch-sink",
    "config": {
        "connector.class": "io.confluent.connect.elasticsearch.ElasticsearchSinkConnector",
        "tasks.max": "1",
        "topics": "development-digital-discovery-categories-2025-04",
        "connection.url": "http://elasticsearch:9200",
        "key.ignore": "false",
        "schema.ignore": "true",
        "type.name": "_doc",
        "behavior.on.null.values": "delete",
        "key.converter": "org.apache.kafka.connect.storage.StringConverter",
        "key.converter.schemas.enable": "false",
        "value.converter": "org.apache.kafka.connect.json.JsonConverter",
        "value.converter.schemas.enable": "false",
        "write.method": "upsert",
        "batch.size": "1",
        "max.retries": "5",
        "retry.backoff.ms": "1000",
        "errors.tolerance": "all",
        "errors.log.enable": true,
        "errors.log.include.messages": true,
        "auto.create.indices.at.start": "true",
        "drop.invalid.message": "false",
        "behavior.on.malformed.documents": "warn"
    }
}'

echo "Verifying sink connector..."
curl -s "http://localhost:8083/connectors/elasticsearch-sink/status" | jq . 

echo -e "\nAvailable sink connector plugins:"
curl -s "http://localhost:8083/connector-plugins" | jq . 