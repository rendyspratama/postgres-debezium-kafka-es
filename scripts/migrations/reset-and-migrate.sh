#!/bin/bash

echo "🔄 Starting database reset and migration process..."

# Run cleanup first
./scripts/cleanup.sh

echo "⏳ Starting migration..."

# Run the migration container
docker-compose run --rm migrate

echo "✅ Migration completed successfully!"

# Optional: Show current tables
echo "📋 Current database schema:"
docker-compose exec postgres psql -U user -d digital_discovery -c "\dt"

# Optional: Show replication settings
echo "📋 Replication settings:"
docker-compose exec postgres psql -U user -d digital_discovery -c "SELECT * FROM pg_publication;"
docker-compose exec postgres psql -U user -d digital_discovery -c "SELECT rolname, rolreplication FROM pg_roles WHERE rolname = 'debezium';"