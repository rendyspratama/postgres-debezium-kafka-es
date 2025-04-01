#!/bin/bash

echo "🚀 Setting up Digital Discovery..."

# Create necessary directories
mkdir -p postgres/conf
mkdir -p debezium/connectors
mkdir -p scripts/migrations

# Create PostgreSQL config
cat > postgres/conf/postgresql.conf << 'EOF'
wal_level = logical
max_wal_senders = 4
max_replication_slots = 4
max_connections = 200
shared_buffers = 256MB
effective_cache_size = 768MB
log_min_messages = warning
log_min_error_statement = error
checkpoint_timeout = 300
checkpoint_completion_target = 0.9
ssl = off
EOF

# Start services
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
until docker-compose exec -T postgres pg_isready -U user; do
  sleep 1
done

echo "✅ PostgreSQL is ready!"

# Run migrations
docker-compose run --rm migrate

echo "🎉 Setup completed successfully!"

# Create Debezium source connector config
cat > debezium/connectors/postgres-source.json << 'EOF'
{
    "name": "postgres-source",
    "config": {
        "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
        "database.hostname": "postgres",
        "database.port": "5432",
        "database.user": "debezium",
        "database.password": "debezium",
        "database.dbname": "digital_discovery",
        "database.server.name": "postgres",
        "table.include.list": "public.categories",
        "plugin.name": "pgoutput",
        "slot.name": "debezium_categories",
        "publication.name": "dbz_publication",
        "transforms": "unwrap",
        "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
        "transforms.unwrap.drop.tombstones": "false",
        "transforms.unwrap.delete.handling.mode": "rewrite"
    }
}
EOF

# Create Elasticsearch sink connector config
cat > debezium/connectors/elasticsearch-sink.json << 'EOF'
{
    "name": "elasticsearch-sink",
    "config": {
        "connector.class": "io.confluent.connect.elasticsearch.ElasticsearchSinkConnector",
        "tasks.max": "1",
        "topics": "postgres.public.categories",
        "connection.url": "http://elasticsearch:9200",
        "type.name": "_doc",
        "key.ignore": "false",
        "schema.ignore": "true",
        "behavior.on.null.values": "delete",
        "transforms": "TimestampRouter",
        "transforms.TimestampRouter.type": "org.apache.kafka.connect.transforms.TimestampRouter",
        "transforms.TimestampRouter.topic.format": "development-digital-discovery-categories-${timestamp}",
        "transforms.TimestampRouter.timestamp.format": "yyyy-MM"
    }
}
EOF

echo "Setup completed. Files created:"
echo "- debezium/connectors/postgres-source.json"
echo "- debezium/connectors/elasticsearch-sink.json"

# Check PostgreSQL status
docker-compose exec postgres psql -U user -d digital_discovery -c "\dx"
docker-compose exec postgres psql -U user -d digital_discovery -c "\dt"
docker-compose exec postgres psql -U user -d digital_discovery -c "SELECT * FROM pg_publication;"
docker-compose exec postgres psql -U user -d digital_discovery -c "SELECT * FROM pg_replication_slots;" 

# Check PostgreSQL logs
docker-compose logs postgres

# Check system logs for PostgreSQL
docker-compose exec postgres cat /var/log/postgresql/postgresql-*.log
