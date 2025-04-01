#!/bin/bash

# Load environment variables
set -a
source .env
set +a

echo "🔍 Verifying setup..."

echo "Checking PostgreSQL..."
docker-compose exec -T postgres pg_isready -U "${POSTGRES_USER}" || exit 1

echo "Checking database objects..."
docker-compose exec -T postgres psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" << EOF
\echo 'Users:'
\du
\echo '\nTables:'
\dt
\echo '\nPublications:'
SELECT * FROM pg_publication;
\echo '\nReplication slots:'
SELECT * FROM pg_replication_slots;
EOF

echo "✅ Verification complete!"