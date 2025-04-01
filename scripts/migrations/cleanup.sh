#!/bin/bash

echo "🗑️  Cleaning up database and migrations..."

# Stop containers if running
docker-compose down

# Remove PostgreSQL volume
docker volume rm digital-discovery_postgres_data

# Remove any existing migration version records
docker-compose up -d postgres

echo "⏳ Waiting for PostgreSQL to start..."
sleep 5

# Connect to PostgreSQL and drop everything
docker-compose exec postgres psql -U user -d digital_discovery -c "
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO public;
"

echo "✨ Database cleaned successfully!"