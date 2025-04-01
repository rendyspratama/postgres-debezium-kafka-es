#!/bin/bash

echo "Waiting for Debezium to be ready..."
until curl -s -f -o /dev/null http://localhost:8084/connectors; do
    echo "Waiting for Debezium..."
    sleep 5
done

echo "Deleting existing source connector (if any)..."
curl -X DELETE http://localhost:8084/connectors/postgres-source 2>/dev/null || true
sleep 2

echo "Deploying PostgreSQL source connector..."
curl -X POST http://localhost:8084/connectors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "postgres-source",
    "config": {
        "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
        "tasks.max": "1",
        "database.hostname": "postgres",
        "database.port": "5432",
        "database.user": "user",
        "database.password": "password",
        "database.dbname": "digital_discovery",
        "database.server.name": "development",
        "topic.prefix": "digital-discovery",
        "schema.include.list": "public",
        "table.include.list": "public.categories",
        "plugin.name": "pgoutput",
        "slot.name": "debezium",
        "publication.name": "dbz_publication",
        "key.converter": "org.apache.kafka.connect.storage.StringConverter",
        "key.converter.schemas.enable": "false",
        "value.converter": "org.apache.kafka.connect.json.JsonConverter",
        "value.converter.schemas.enable": "false",
        "transforms": "unwrap,route,extractId,toString",
        "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
        "transforms.unwrap.drop.tombstones": "false",
        "transforms.unwrap.delete.handling.mode": "rewrite",
        "transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
        "transforms.route.regex": ".*",
        "transforms.route.replacement": "development-digital-discovery-categories-2025-04",
        "transforms.extractId.type": "org.apache.kafka.connect.transforms.ValueToKey",
        "transforms.extractId.fields": "id",
        "transforms.toString.type": "org.apache.kafka.connect.transforms.ExtractField$Key",
        "transforms.toString.field": "id",
        "tombstones.on.delete": "false",
        "after.state.only": "false",
        "provide.transaction.metadata": "false"
    }
}'

echo "Verifying source connector..."
curl -s "http://localhost:8084/connectors/postgres-source/status" | jq . 

echo -e "\nAvailable source connector plugins:"
curl -s "http://localhost:8084/connector-plugins" | jq . 