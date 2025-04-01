# Digital Discovery Platform

This repository contains two main services:
1. API Service - Handles HTTP requests and manages category data
2. Sync Service - Synchronizes category data with Elasticsearch

Please refer to the specific documentation for each service:
- [API Service Documentation](./api/README.md)
- [Sync Service Documentation](./sync/README.md)

## Prerequisites

- Go 1.21 or later
- Docker 24.0 or later
- Docker Compose v2.0 or later
- Make (for using Makefile commands)
- PostgreSQL client (for database operations)

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

## Setup and Running

### 1. Clone the Repository

```bash
git clone <repository-url>
cd digital-discovery
```

### 2. Environment Setup

Create a `.env` file in the root directory:

```env
# Database
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=digital_discovery

# API
API_PORT=8081
```

### 3. Start Services

Using Docker Compose:

```bash
# Start all services
docker-compose up -d

# Check service status
docker-compose ps
```

### 4. Run Database Migrations

```bash
# Apply migrations
make migrate-up

# Verify migrations
make migrate-verify

# If needed, rollback migrations
make migrate-down
```

### 5. Seed Initial Data

```bash
# Apply seed data
make seed-apply
```

### 6. Build and Run API Service

```bash
# Build the API
make build-api

# Run the API
make run-api
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run specific test
make test-categories
```

### Database Management

```bash
# Access database shell
make db-shell

# Backup database
make db-backup

# Restore database
make db-restore
```

### Monitoring

```bash
# Monitor migrations
make migrate-monitor

# View metrics
curl http://localhost:8081/metrics
```

## API Endpoints

### Health Check
```bash
curl -X GET http://localhost:8081/health
```

### Categories API (v1)
```bash
# List categories
curl -X GET http://localhost:8081/api/categories

# Get single category
curl -X GET "http://localhost:8081/api/categories?id=1"

# Create category
curl -X POST http://localhost:8081/api/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "Pulsa", "status": 1}'

# Update category
curl -X PUT "http://localhost:8081/api/categories?id=1" \
  -H "Content-Type: application/json" \
  -d '{"name": "Pulsa Updated", "status": 1}'

# Delete category
curl -X DELETE "http://localhost:8081/api/categories?id=1"
```

### Categories API (v2)
```bash
# List categories with pagination
curl -X GET "http://localhost:8081/api/categories?page=1&per_page=10"
```

### Documentation
```bash
# View middleware documentation
curl -X GET http://localhost:8081/docs/middleware
```

## Monitoring and Management

- API Metrics: http://localhost:8081/metrics
- Kafka UI: http://localhost:8080
- Elasticsearch: http://localhost:9200

## Troubleshooting

### Common Issues

1. Database Connection Issues
   ```bash
   # Check database logs
   docker-compose logs postgres
   
   # Verify database connection
   make db-shell
   ```

2. API Service Issues
   ```bash
   # Check API logs
   docker-compose logs api
   
   # Restart API service
   docker-compose restart api
   ```

3. Migration Issues
   ```bash
   # Check migration status
   make migrate-status
   
   # Fix migration issues
   make migrate-fix
   ```
### Logs

```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f api
docker-compose logs -f postgres
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

