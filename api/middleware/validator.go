package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/rendyspratama/digital-discovery/api/config"
	"github.com/rendyspratama/digital-discovery/api/utils"

	"github.com/go-playground/validator/v10"
)

type ValidationMiddleware struct {
	config    config.MiddlewareConfig
	validator *validator.Validate
}

func NewValidationMiddleware(cfg config.MiddlewareConfig) *ValidationMiddleware {
	return &ValidationMiddleware{
		config:    cfg,
		validator: validator.New(),
	}
}

func (v *ValidationMiddleware) Validate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only validate POST and PUT requests
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// Check content length
		if r.ContentLength > v.config.Validation.MaxBodySize {
			utils.WriteError(w, http.StatusRequestEntityTooLarge, "Request body too large")
			return
		}

		// Get the request body from context
		body, ok := r.Context().Value("requestBody").([]byte)
		if !ok {
			utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Try to unmarshal into a map first to check if it's valid JSON
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			utils.WriteError(w, http.StatusBadRequest, "Invalid JSON format")
			return
		}

		// Get the resource type from the URL path
		path := strings.TrimPrefix(r.URL.Path, "/api/")
		resourceType := strings.Split(path, "/")[0]

		// Get validation rules for the resource
		rules, ok := v.config.Validation.Rules[resourceType]
		if !ok {
			utils.WriteError(w, http.StatusBadRequest, "Unknown resource type")
			return
		}

		// Validate the data against rules
		if err := v.validateData(data, rules); err != nil {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (v *ValidationMiddleware) validateData(data map[string]interface{}, rules config.ValidationRule) error {
	// Check if required
	if rules.Required && len(data) == 0 {
		return fmt.Errorf("data is required")
	}

	// Validate type
	if rules.Type != "" {
		switch rules.Type {
		case "string":
			for key, value := range data {
				if _, ok := value.(string); !ok {
					return fmt.Errorf("field %s must be a string", key)
				}
				strValue := value.(string)
				if rules.Min != nil {
					min, _ := strconv.Atoi(rules.Min.(string))
					if len(strValue) < min {
						return fmt.Errorf("field %s must be at least %d characters", key, min)
					}
				}
				if rules.Max != nil {
					max, _ := strconv.Atoi(rules.Max.(string))
					if len(strValue) > max {
						return fmt.Errorf("field %s must be at most %d characters", key, max)
					}
				}
				if rules.Pattern != "" {
					matched, _ := regexp.MatchString(rules.Pattern, strValue)
					if !matched {
						return fmt.Errorf("field %s does not match pattern", key)
					}
				}
			}
		case "integer":
			for key, value := range data {
				if _, ok := value.(float64); !ok {
					return fmt.Errorf("field %s must be an integer", key)
				}
				intValue := int(value.(float64))
				if rules.Min != nil {
					min := int(rules.Min.(float64))
					if intValue < min {
						return fmt.Errorf("field %s must be at least %d", key, min)
					}
				}
				if rules.Max != nil {
					max := int(rules.Max.(float64))
					if intValue > max {
						return fmt.Errorf("field %s must be at most %d", key, max)
					}
				}
				if len(rules.Enum) > 0 {
					valid := false
					for _, enumValue := range rules.Enum {
						if intValue == int(enumValue.(float64)) {
							valid = true
							break
						}
					}
					if !valid {
						return fmt.Errorf("field %s must be one of %v", key, rules.Enum)
					}
				}
			}
		case "object":
			if rules.Rules != nil {
				for key, fieldRules := range rules.Rules {
					if value, exists := data[key]; exists {
						if err := v.validateField(value, fieldRules); err != nil {
							return fmt.Errorf("field %s: %v", key, err)
						}
					} else if fieldRules.Required {
						return fmt.Errorf("field %s is required", key)
					}
				}
			}
		}
	}

	return nil
}

func (v *ValidationMiddleware) validateField(value interface{}, rules config.ValidationRule) error {
	if rules.Required && value == nil {
		return fmt.Errorf("value is required")
	}

	switch rules.Type {
	case "string":
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("must be a string")
		}
		if rules.Min != nil {
			min, _ := strconv.Atoi(rules.Min.(string))
			if len(strValue) < min {
				return fmt.Errorf("must be at least %d characters", min)
			}
		}
		if rules.Max != nil {
			max, _ := strconv.Atoi(rules.Max.(string))
			if len(strValue) > max {
				return fmt.Errorf("must be at most %d characters", max)
			}
		}
		if rules.Pattern != "" {
			matched, _ := regexp.MatchString(rules.Pattern, strValue)
			if !matched {
				return fmt.Errorf("does not match pattern")
			}
		}
	case "integer":
		floatValue, ok := value.(float64)
		if !ok {
			return fmt.Errorf("must be an integer")
		}
		intValue := int(floatValue)
		if rules.Min != nil {
			min := int(rules.Min.(float64))
			if intValue < min {
				return fmt.Errorf("must be at least %d", min)
			}
		}
		if rules.Max != nil {
			max := int(rules.Max.(float64))
			if intValue > max {
				return fmt.Errorf("must be at most %d", max)
			}
		}
		if len(rules.Enum) > 0 {
			valid := false
			for _, enumValue := range rules.Enum {
				if intValue == int(enumValue.(float64)) {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("must be one of %v", rules.Enum)
			}
		}
	}

	return nil
}
