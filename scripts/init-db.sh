#!/bin/bash

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found!"
    exit 1
fi

# Load environment variables
set -a
source .env
set +a

# Verify environment variables are loaded
echo "🔍 Verifying environment variables..."
if [ -z "${POSTGRES_DB}" ]; then
    echo "❌ Error: POSTGRES_DB is not set in .env file"
    exit 1
fi
if [ -z "${POSTGRES_USER}" ]; then
    echo "❌ Error: POSTGRES_USER is not set in .env file"
    exit 1
fi
if [ -z "${POSTGRES_PASSWORD}" ]; then
    echo "❌ Error: POSTGRES_PASSWORD is not set in .env file"
    exit 1
fi

echo "✅ Environment variables loaded:"
echo "Database: ${POSTGRES_DB}"
echo "User: ${POSTGRES_USER}"

echo "🚀 Initializing database..."

# Start PostgreSQL service
echo "🚀 Starting PostgreSQL..."
docker-compose up -d postgres

# Wait for PostgreSQL to be ready (with timeout)
echo "⏳ Waiting for PostgreSQL to be ready..."
timeout=60
counter=0
until docker-compose exec -T postgres pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" > /dev/null 2>&1; do
    if [ $counter -gt $timeout ]; then
        echo "❌ Timeout waiting for PostgreSQL to be ready"
        exit 1
    fi
    echo "⏳ Waiting for PostgreSQL... ($counter seconds)"
    sleep 1
    counter=$((counter + 1))
done

echo "✅ PostgreSQL is ready!"

# Run migrations
echo "Running migrations..."
docker-compose run --rm migrate

echo "✅ Database initialized!"

# Verify setup
echo "📋 Verifying setup..."

# Check users
docker-compose exec postgres psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "\du"

# Check tables
docker-compose exec postgres psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "\dt"

# Check publications
docker-compose exec postgres psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "SELECT * FROM pg_publication;"

# Check replication slots
docker-compose exec postgres psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "SELECT * FROM pg_replication_slots;"