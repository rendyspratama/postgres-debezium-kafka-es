# Sync Service

The sync service handles data synchronization between PostgreSQL and Elasticsearch using Debezium and Kafka.

## Structure

```
sync/
├── api/             # API handlers
├── config/          # Configuration
├── consumers/       # Kafka consumers
├── metrics/         # Prometheus

## Features

- Real-time data synchronization from PostgreSQL to Elasticsearch
- Kafka message consumption and processing
- Bulk indexing operations
- Automatic index lifecycle management
- Health monitoring
- Prometheus metrics
- Graceful shutdown handling
- Retry mechanisms for failed operations

## Prerequisites

- Go 1.21 or later
- Docker 24.0 or later
- Docker Compose v2.0 or later
- Elasticsearch 8.x
- Apache Kafka
- Make

## Quick Start

1. **Clone the repository**
```bash
git clone <repository-url>
cd digital-discovery
```

2. **Set up environment variables**
Create a `.env` file in the root directory:
```env
# Elasticsearch
ES_HOSTS=http://localhost:9200
ES_USERNAME=elastic
ES_PASSWORD=changeme

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=category-events
KAFKA_GROUP_ID=sync-service

# Service
SYNC_PORT=8082
```

3. **Start dependencies**
```bash
# Start all required services (Elasticsearch, Kafka, Zookeeper)
make docker-up

# Verify services are running
docker-compose ps
```

4. **Build and run the service**
```bash
# Build the sync service
make build-sync

# Run the sync service
make run-sync
```

## Configuration

```yaml
# config.yaml
sync:
  port: 8082
  service_name: "digital-discovery-sync"
  shutdown_timeout: "30s"

elasticsearch:
  hosts: ["http://localhost:9200"]
  username: "elastic"
  password: "changeme"
  index_prefix: "categories"
  retry_backoff: "5s"
  max_retries: 3
  bulk_size: 1000
  flush_interval: "5s"

kafka:
  brokers: ["localhost:9092"]
  topic: "category-events"
  group_id: "sync-service"
  auto_offset_reset: "earliest"
  enable_auto_commit: true
  session_timeout: "30s"

monitoring:
  metrics_enabled: true
  metrics_port: 9091
  health_check_port: 8082
```

## Health Check Endpoints

```bash
# Health check
curl http://localhost:8082/health

# Readiness check (includes Elasticsearch and Kafka connection status)
curl http://localhost:8082/ready

# Metrics
curl http://localhost:8082/metrics
```

## Monitoring

### Available Metrics
- Kafka message processing stats
- Elasticsearch bulk operation stats
- Sync operation latencies
- Error counts
- System metrics

### Logging
JSON structured logging with fields:
- operation
- status
- latency
- error details
- batch size
- index name

## Troubleshooting

### Common Issues

1. **Elasticsearch Connection Issues**
```bash
# Check Elasticsearch is running
curl http://localhost:9200/_cluster/health

# Verify credentials
curl -u elastic:changeme http://localhost:9200
```

2. **Kafka Connection Issues**
```bash
# Check Kafka logs
docker-compose logs kafka

# Verify topic exists
docker-compose exec kafka kafka-topics.sh --list --bootstrap-server localhost:9092
```

3. **Sync Service Issues**
```bash
# Check service logs
docker-compose logs sync-service

# Verify service is running
curl http://localhost:8082/health
```

## Development

```bash
# Run tests
make test-sync

# Run linter
make lint

# Format code
make fmt

# Clean build artifacts
make clean
```

## Project Structure 

## API Documentation

### Categories API

The sync service provides a REST API for managing categories. Below are the available endpoints and example curl commands for testing.

#### List Categories
```bash
curl 'http://localhost:8082/api/v1/categories'
```

#### Create Category
```bash
curl -X POST 'http://localhost:8082/api/v1/categories' \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-category",
    "name": "Test Category",
    "description": "Test Description"
  }'
```

#### Get Category by ID
```bash
curl 'http://localhost:8082/api/v1/category?id=test-category'
```

#### Update Category
```bash
curl -X PUT 'http://localhost:8082/api/v1/category?id=test-category' \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-category",
    "name": "Updated Category",
    "description": "Updated Description"
  }'
```

#### Delete Category
```bash
curl -X DELETE 'http://localhost:8082/api/v1/category?id=test-category'
```

### Important Notes:
- All URLs should be wrapped in quotes to handle special characters correctly
- The service runs on port 8082 by default
- Content-Type header must be set to "application/json" for POST and PUT requests
- Category ID is passed as a query parameter for GET, PUT, and DELETE operations 

## Data Flow

1. **PostgreSQL Changes**
   - Debezium captures changes from PostgreSQL
   - Changes are published to Kafka topics

2. **Kafka Processing**
   - Sync service consumes messages from Kafka
   - Processes and transforms data

3. **Elasticsearch Sync**
   - Updates Elasticsearch index
   - Maintains data consistency

## Configuration

### Debezium Source Connector

```bash
# Deploy source connector
./scripts/deploy-source-connector.sh
```

Configuration:
```json
{
  "name": "postgres-source",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "user",
    "database.password": "password",
    "database.dbname": "digital_discovery",
    "database.server.name": "development",
    "topic.prefix": "digital-discovery",
    "schema.include.list": "public",
    "table.include.list": "public.categories"
  }
}
```

### Elasticsearch Sink Connector

```bash
# Deploy sink connector
./scripts/deploy-sink-connector.sh
```

Configuration:
```json
{
  "name": "elasticsearch-sink",
  "config": {
    "connector.class": "io.confluent.connect.elasticsearch.ElasticsearchSinkConnector",
    "connection.url": "http://elasticsearch:9200",
    "topics": "development-digital-discovery-categories-2025-04",
    "key.ignore": "false",
    "schema.ignore": "true"
  }
}
```

## Development

### Running Locally

```bash
# Run with hot reload
make run-sync

# Run tests
make test-sync

# Build binary
make build-sync
```

### Connector Management

```bash
# Check connector status
curl http://localhost:8083/connectors/elasticsearch-sink/status
curl http://localhost:8084/connectors/postgres-source/status

# View connector logs
curl http://localhost:8083/connectors/elasticsearch-sink/tasks/0/log
```

### Monitoring

- **Kafka UI**: http://localhost:8080
- **Elasticsearch**: http://localhost:9200
- **Metrics**: http://localhost:8082/metrics

## Troubleshooting

### Common Issues

1. **Kafka Connect Issues**
   ```bash
   # Check connector status
   curl http://localhost:8083/connectors/elasticsearch-sink/status
   
   # View connector logs
   curl http://localhost:8083/connectors/elasticsearch-sink/tasks/0/log
   ```

2. **Elasticsearch Sync Issues**
   ```bash
   # Check index
   curl http://localhost:9200/development-digital-discovery-categories-2025-04/_search
   
   # Check mapping
   curl http://localhost:9200/development-digital-discovery-categories-2025-04/_mapping
   ```

3. **Data Flow Verification**
   ```bash
   # Check Kafka messages
   docker-compose exec kafka kafka-console-consumer \
     --bootstrap-server kafka:29092 \
     --topic development-digital-discovery-categories-2025-04 \
     --from-beginning
   ```

## Metrics

The service exposes Prometheus metrics for:
- Message processing rate
- Error rates
- Processing latency
- Connector status

## Contributing

1. Follow the Go code style guide
2. Write tests for new features
3. Update documentation
4. Create a pull request 