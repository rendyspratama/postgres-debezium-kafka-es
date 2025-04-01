# Digital Discovery API Service

RESTful API service for managing digital product categories with real-time synchronization capabilities.

## Features

- RESTful API endpoints for category management
- Real-time data synchronization with Elasticsearch via Kafka
- OpenAPI/Swagger documentation
- Structured logging
- Prometheus metrics
- Health checks
- Request validation
- Response caching
- Rate limiting
- API versioning (v1, v2)

## API Endpoints

### Health and Metrics
```bash
# Health check
GET /health

# Readiness check
GET /ready

# Prometheus metrics
GET /metrics
```

### Categories API (v1)
```bash
# List categories
GET /api/v1/categories
Query Parameters:
  - page (int, default: 1)
  - per_page (int, default: 10)
  - sort (string, default: "created_at")
  - order (string, default: "desc")

# Get single category
GET /api/v1/categories/{id}

# Create category
POST /api/v1/categories
Body:
{
    "name": "Pulsa",
    "description": "Pulsa all operator",
    "status": "active"
}

# Update category
PUT /api/v1/categories/{id}
Body:
{
    "name": "Pulsa Updated",
    "description": "Updated description",
    "status": "inactive"
}

# Delete category
DELETE /api/v1/categories/{id}
```

### Categories API (v2)
```bash
# Enhanced list with advanced filtering
GET /api/v2/categories
Query Parameters:
  - page (int)
  - per_page (int)
  - sort (string)
  - order (string)
  - status (string)
  - search (string)
  - created_after (datetime)
  - created_before (datetime)
```

## Configuration

```yaml
# config.yaml
api:
  port: 8081
  env: "development"
  service_name: "digital-discovery-api"
  request_timeout: "30s"
  shutdown_timeout: "10s"
  cors:
    enabled: true
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
  rate_limit:
    enabled: true
    requests_per_second: 100

database:
  host: "localhost"
  port: 5432
  name: "digital_discovery"
  user: "postgres"
  password: "password"
  max_connections: 20
  idle_timeout: "5m"

kafka:
  brokers: ["localhost:9092"]
  topic: "category-events"
  group_id: "api-service"
  auto_offset_reset: "earliest"

monitoring:
  log_format: "json"
  metrics_enabled: true
  metrics_port: 9090
  health_check_port: 8081
```

## Project Structure 