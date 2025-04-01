#!/bin/bash
# cleanup-all.sh

echo "🧹 Cleaning up everything..."

# Stop all containers
docker-compose down

# Remove volumes
docker volume rm digital-discovery_postgres_data digital-discovery_elasticsearch_data

# Remove any leftover containers
docker rm -f $(docker ps -a | grep 'digital-discovery' | awk '{print $1}') 2>/dev/null || true

# Clean up network
docker network rm digital-discovery-network 2>/dev/null || true

echo "✨ Cleanup completed!"