# Digital Discovery Platform

A microservices-based platform for managing and synchronizing category data between PostgreSQL and Elasticsearch using Debezium and Kafka.

## Architecture

The platform consists of two main services:
1. **API Service** - REST API for managing categories
2. **Sync Service** - Handles data synchronization between PostgreSQL and Elasticsearch

### Technology Stack
- **API Service**: Go, PostgreSQL
- **Sync Service**: Go, Kafka, Debezium, Elasticsearch
- **Infrastructure**: Docker, Docker Compose

## Project Structure

```
.
├── api/                    # API service
│   ├── config/            # Configuration
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/           # Data models
│   ├── repositories/     # Database repositories
│   ├── routes/          # API routes
│   ├── services/        # Business logic
│   └── utils/           # Utility functions
├── scripts/              # Scripts directory
│   └── migrations/      # Database migrations
├── docker-compose.yml    # Docker services configuration
└── Makefile             # Build and utility commands
```

## Services

- API Service (Go)
- PostgreSQL
- Elasticsearch
- Kafka
- Zookeeper
- Debezium
- Kafka UI

## Prerequisites

- Go 1.21 or later
- Docker 24.0 or later
- Docker Compose v2.0 or later
- Make (for using Makefile commands)

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd digital-discovery
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start services**
   ```bash
   docker-compose up -d
   ```

4. **Run migrations**
   ```bash
   make migrate-up
   ```

5. **Deploy connectors**
   ```bash
   ./scripts/deploy-source-connector.sh
   ./scripts/deploy-sink-connector.sh
   ```

## API Endpoints

### Categories API (v1)

```bash
# Create category
curl -X POST http://localhost:8081/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Category",
    "description": "Test Description",
    "status": 1
  }'

# Get category by ID
curl -X GET http://localhost:8081/api/v1/categories/1

# Update category
curl -X PUT http://localhost:8081/api/v1/categories/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Category",
    "description": "Updated Description",
    "status": 1
  }'

# Delete category
curl -X DELETE http://localhost:8081/api/v1/categories/1
```

## Configuration

Create a `.env` file in the root directory:

```env
# API Configuration
API_PORT=8081
API_ENV=development

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=digital_discovery
DB_SSL_MODE=disable
```

## Development

### Running Locally

```bash
# Run with hot reload
make run-api

# Run tests
make test-api

# Build binary
make build-api
```

### Database Migrations

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Check migration status
make migrate-status
```

### Testing

```bash
# Run all tests
make test-api

# Run specific test
go test ./handlers -v
```

## Monitoring

- **Health Check**: http://localhost:8081/health
- **Metrics**: http://localhost:8081/metrics

## Error Handling

The API uses standard HTTP status codes and returns errors in the following format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description"
  }
}
```

## Middleware

- Request ID tracking
- CORS handling
- Request logging
- Error recovery
- Authentication (if configured)

## Contributing

1. Follow the Go code style guide
2. Write tests for new features
3. Update documentation
4. Create a pull request

## License

This project is licensed under the MIT License.

