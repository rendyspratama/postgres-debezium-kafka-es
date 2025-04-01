package middleware

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
)

// MiddlewareDoc represents documentation for a middleware
type MiddlewareDoc struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Config      interface{}       `json:"config,omitempty"`
	Examples    []Example         `json:"examples,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// Example represents an example usage of middleware
type Example struct {
	Description string      `json:"description"`
	Code        string      `json:"code"`
	Response    interface{} `json:"response,omitempty"`
}

// MiddlewareDocs manages documentation for all middleware
type MiddlewareDocs struct {
	docs map[string]MiddlewareDoc
}

// NewMiddlewareDocs creates a new middleware documentation manager
func NewMiddlewareDocs() *MiddlewareDocs {
	md := &MiddlewareDocs{
		docs: make(map[string]MiddlewareDoc),
	}
	md.initDefaultDocs()
	return md
}

// initDefaultDocs initializes documentation for built-in middleware
func (md *MiddlewareDocs) initDefaultDocs() {
	// CORS middleware documentation
	md.AddDoc(MiddlewareDoc{
		Name:        "CORS",
		Description: "Handles Cross-Origin Resource Sharing (CORS) headers and preflight requests",
		Config: struct {
			AllowedOrigins []string `json:"allowedOrigins"`
			AllowedMethods []string `json:"allowedMethods"`
			AllowedHeaders []string `json:"allowedHeaders"`
			MaxAge         int      `json:"maxAge"`
		}{},
		Examples: []Example{
			{
				Description: "Basic CORS setup",
				Code: `router.Use(middleware.CORS(middleware.CORSConfig{
    AllowedOrigins: []string{"*"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
}))`,
			},
		},
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "Allowed origins",
			"Access-Control-Allow-Methods": "Allowed HTTP methods",
			"Access-Control-Allow-Headers": "Allowed headers",
		},
	})

	// Recovery middleware documentation
	md.AddDoc(MiddlewareDoc{
		Name:        "Recovery",
		Description: "Recovers from panics and converts them to errors",
		Config:      RecoveryConfig{},
		Examples: []Example{
			{
				Description: "Basic recovery setup",
				Code:        `router.Use(middleware.Recovery(middleware.DefaultRecoveryConfig()))`,
			},
		},
	})

	// Logger middleware documentation
	md.AddDoc(MiddlewareDoc{
		Name:        "Logger",
		Description: "Logs HTTP request and response details",
		Config: struct {
			Format     string `json:"format"`
			TimeFormat string `json:"timeFormat"`
		}{},
		Examples: []Example{
			{
				Description: "Basic logger setup",
				Code:        `router.Use(middleware.Logger)`,
			},
		},
	})

	// Validator middleware documentation
	md.AddDoc(MiddlewareDoc{
		Name:        "Validator",
		Description: "Validates request bodies against defined rules",
		Config: struct {
			Rules map[string]interface{} `json:"rules"`
		}{},
		Examples: []Example{
			{
				Description: "Validation setup for category",
				Code: `validator := middleware.NewValidationMiddleware(config)
router.Use(validator.Validate)`,
			},
		},
	})
}

// AddDoc adds documentation for a middleware
func (md *MiddlewareDocs) AddDoc(doc MiddlewareDoc) {
	md.docs[doc.Name] = doc
}

// GetDoc returns documentation for a middleware
func (md *MiddlewareDocs) GetDoc(name string) (MiddlewareDoc, bool) {
	doc, ok := md.docs[name]
	return doc, ok
}

// GenerateDocsHandler returns a handler that serves middleware documentation
func (md *MiddlewareDocs) GenerateDocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(md.docs)
	}
}

// GenerateMarkdown generates markdown documentation
func (md *MiddlewareDocs) GenerateMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# Middleware Documentation\n\n")

	for _, doc := range md.docs {
		// Write middleware name and description
		sb.WriteString("## " + doc.Name + "\n\n")
		sb.WriteString(doc.Description + "\n\n")

		// Write configuration
		if doc.Config != nil {
			sb.WriteString("### Configuration\n\n")
			sb.WriteString("```go\n")
			sb.WriteString(formatConfig(doc.Config))
			sb.WriteString("```\n\n")
		}

		// Write examples
		if len(doc.Examples) > 0 {
			sb.WriteString("### Examples\n\n")
			for _, example := range doc.Examples {
				sb.WriteString("#### " + example.Description + "\n\n")
				sb.WriteString("```go\n")
				sb.WriteString(example.Code + "\n")
				sb.WriteString("```\n\n")
			}
		}

		// Write headers
		if len(doc.Headers) > 0 {
			sb.WriteString("### Headers\n\n")
			sb.WriteString("| Header | Description |\n")
			sb.WriteString("|--------|-------------|\n")
			for header, desc := range doc.Headers {
				sb.WriteString("| " + header + " | " + desc + " |\n")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// formatConfig formats a config object for documentation
func formatConfig(config interface{}) string {
	t := reflect.TypeOf(config)
	var sb strings.Builder

	sb.WriteString("type Config struct {\n")
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		sb.WriteString("    " + field.Name + " " + field.Type.String() + "\n")
	}
	sb.WriteString("}\n")

	return sb.String()
}
