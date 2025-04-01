package config

import "time"

type MiddlewareConfig struct {
	CORS struct {
		AllowedOrigins []string
		AllowedMethods []string
		AllowedHeaders []string
		MaxAge         int
	}
	Logger struct {
		Format     string
		TimeFormat string
		Level      string
	}
	Validation struct {
		MaxBodySize int64
		Rules       map[string]ValidationRule
	}
}

type ValidationRule struct {
	Required bool
	Type     string
	Min      interface{}
	Max      interface{}
	Pattern  string
	Enum     []interface{}
	Rules    map[string]ValidationRule
}

func LoadMiddlewareConfig() MiddlewareConfig {
	cfg := MiddlewareConfig{}

	// CORS Configuration
	cfg.CORS.AllowedOrigins = []string{"*"}
	cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cfg.CORS.AllowedHeaders = []string{
		"Accept",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"X-Request-ID",
	}
	cfg.CORS.MaxAge = 86400 // 24 hours

	// Logger Configuration
	cfg.Logger.Format = "[%s] %s %s %d %s %s %s"
	cfg.Logger.TimeFormat = time.RFC3339
	cfg.Logger.Level = "info"

	// Validation Configuration
	cfg.Validation.MaxBodySize = 1024 * 1024 // 1MB
	cfg.Validation.Rules = map[string]ValidationRule{
		"category": {
			Required: true,
			Type:     "object",
			Rules: map[string]ValidationRule{
				"name": {
					Required: true,
					Type:     "string",
					Min:      3,
					Max:      100,
				},
				"status": {
					Required: true,
					Type:     "integer",
					Enum:     []interface{}{0, 1},
				},
			},
		},
		"operator": {
			Required: true,
			Type:     "object",
			Rules: map[string]ValidationRule{
				"name": {
					Required: true,
					Type:     "string",
					Min:      2,
					Max:      50,
				},
				"category_id": {
					Required: true,
					Type:     "integer",
					Min:      1,
				},
				"status": {
					Required: true,
					Type:     "integer",
					Enum:     []interface{}{0, 1},
				},
			},
		},
	}

	return cfg
}
