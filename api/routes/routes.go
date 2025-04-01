package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rendyspratama/digital-discovery/api/config"
	"github.com/rendyspratama/digital-discovery/api/handlers"
	"github.com/rendyspratama/digital-discovery/api/middleware"
	"github.com/rendyspratama/digital-discovery/api/repositories"
)

// APIDocumentation contains the documentation for all API endpoints
var APIDocumentation = `
Digital Discovery API Documentation
=================================

Base URL: /api

Authentication
-------------
All API endpoints require authentication via Bearer token in the Authorization header.
Example: Authorization: Bearer <your-token>

Health Check
-----------
GET /health
- Description: Check if the API is running
- Response: 200 OK
  {
    "status": "ok",
    "timestamp": "2024-03-21T15:04:05Z"
  }

Categories API v1
----------------
Base path: /api/v1/categories

1. List Categories
GET /api/v1/categories
- Description: Get all categories
- Query Parameters:
  * limit (optional): Number of items per page (default: 10)
  * offset (optional): Starting position (default: 0)
- Response: 200 OK
  {
    "data": [
      {
        "id": "uuid",
        "name": "string",
        "description": "string",
        "created_at": "timestamp",
        "updated_at": "timestamp"
      }
    ],
    "metadata": {
      "total": integer,
      "limit": integer,
      "offset": integer
    }
  }

2. Create Category
POST /api/v1/categories
- Description: Create a new category
- Request Body:
  {
    "name": "string",
    "description": "string"
  }
- Response: 201 Created
  {
    "data": {
      "id": "uuid",
      "name": "string",
      "description": "string",
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  }

3. Get Category by ID
GET /api/v1/categories/{id}
- Description: Get category details by ID
- Parameters:
  * id: Category UUID
- Response: 200 OK
  {
    "data": {
      "id": "uuid",
      "name": "string",
      "description": "string",
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  }

4. Update Category
PUT /api/v1/categories/{id}
- Description: Update category details
- Parameters:
  * id: Category UUID
- Request Body:
  {
    "name": "string",
    "description": "string"
  }
- Response: 200 OK
  {
    "data": {
      "id": "uuid",
      "name": "string",
      "description": "string",
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  }

5. Delete Category
DELETE /api/v1/categories/{id}
- Description: Delete a category
- Parameters:
  * id: Category UUID
- Response: 204 No Content

Categories API v2
----------------
Base path: /api/v2/categories

1. List Categories (Enhanced)
GET /api/v2/categories
- Description: Get all categories with enhanced features
- Query Parameters:
  * limit (optional): Number of items per page (default: 10)
  * offset (optional): Starting position (default: 0)
  * sort (optional): Sort field (name, created_at)
  * order (optional): Sort order (asc, desc)
- Response: 200 OK
  {
    "data": [
      {
        "id": "uuid",
        "name": "string",
        "description": "string",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "metadata": {
          "items_count": integer
        }
      }
    ],
    "metadata": {
      "total": integer,
      "limit": integer,
      "offset": integer
    }
  }

Metrics
-------
GET /metrics
- Description: Get API performance metrics
- Response: 200 OK
  Content-Type: text/plain
  Shows:
  * API-wide metrics
  * V1 Categories metrics
  * V2 Categories metrics
  * Latency and error rates

Documentation
------------
GET /docs/middleware
- Description: Get detailed API documentation
- Response: 200 OK
  Returns this documentation in HTML format
`

func SetupRouter() http.Handler {
	// Load configurations
	middlewareConfig := config.LoadMiddlewareConfig()

	// Initialize repositories
	categoryRepo := repositories.NewCategoryRepository()

	// Initialize handlers
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)

	// Initialize middleware components
	logger := middleware.NewLoggerMiddleware(middlewareConfig)
	cors := middleware.NewCORSMiddleware(middlewareConfig)
	// validator := middleware.NewValidationMiddleware(middlewareConfig)
	metrics := middleware.NewMiddlewareMetrics()
	// docs := middleware.NewMiddlewareDocs()
	recovery := middleware.Recovery(middleware.DefaultRecoveryConfig())

	// Create router
	r := chi.NewRouter()

	// Add global middleware in correct order
	r.Use(middleware.RequestID)
	r.Use(logger.Logger)
	r.Use(recovery)
	r.Use(cors.CORS)
	r.Use(middleware.ResponseMetadata)

	// Health check route
	r.Get("/health", handlers.HealthCheck)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Track all API requests
		r.Use(func(next http.Handler) http.Handler {
			return metrics.Track("api", next)
		})

		// V1 routes
		r.Route("/v1", func(r chi.Router) {
			// Categories endpoints
			r.Route("/categories", func(r chi.Router) {
				// Base metrics for categories
				r.Use(func(next http.Handler) http.Handler {
					return metrics.Track("v1.categories", next)
				})

				r.Get("/", categoryHandler.GetCategories)
				// r.With(validator.Validate, middleware.BodyParser).
				// 	Post("/", categoryHandler.CreateCategory)
				r.Post("/", categoryHandler.CreateCategory)
				r.Get("/{id}", categoryHandler.GetCategory)
				// r.With(validator.Validate, middleware.BodyParser).
				// 	Put("/{id}", categoryHandler.UpdateCategory)
				r.Put("/{id}", categoryHandler.UpdateCategory)
				r.Delete("/{id}", categoryHandler.DeleteCategory)
			})
		})

		// V2 routes
		r.Route("/v2", func(r chi.Router) {
			r.Route("/categories", func(r chi.Router) {
				// V2 metrics
				r.Use(func(next http.Handler) http.Handler {
					return metrics.Track("v2.categories", next)
				})
				r.Get("/", categoryHandler.GetCategoriesV2)
			})
		})
	})

	// Metrics endpoint
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		// Get metrics data
		apiMetrics := metrics.GetMetrics("api")
		v1Metrics := metrics.GetMetrics("v1.categories")
		v2Metrics := metrics.GetMetrics("v2.categories")

		// Write metrics report
		fmt.Fprintf(w, "=== API Metrics ===\n")
		if apiMetrics != nil {
			fmt.Fprintf(w, "API Latency: %.2fms\n", metrics.GetAverageLatency("api"))
			fmt.Fprintf(w, "API Error Rate: %.2f%%\n\n", metrics.GetErrorRate("api"))
		}

		fmt.Fprintf(w, "=== V1 Categories Metrics ===\n")
		if v1Metrics != nil {
			fmt.Fprintf(w, "Latency: %.2fms\n", metrics.GetAverageLatency("v1.categories"))
			fmt.Fprintf(w, "Error Rate: %.2f%%\n\n", metrics.GetErrorRate("v1.categories"))
		}

		fmt.Fprintf(w, "=== V2 Categories Metrics ===\n")
		if v2Metrics != nil {
			fmt.Fprintf(w, "Latency: %.2fms\n", metrics.GetAverageLatency("v2.categories"))
			fmt.Fprintf(w, "Error Rate: %.2f%%\n", metrics.GetErrorRate("v2.categories"))
		}
	})

	// Documentation endpoint
	r.Get("/docs/middleware", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		htmlDocs := strings.ReplaceAll(APIDocumentation, "\n", "<br>")
		htmlDocs = "<pre>" + htmlDocs + "</pre>"
		fmt.Fprint(w, htmlDocs)
	})

	return r
}
