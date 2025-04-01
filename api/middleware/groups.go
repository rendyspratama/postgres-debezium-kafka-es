package middleware

import (
	"net/http"
)

// MiddlewareGroup represents a named group of middleware
type MiddlewareGroup struct {
	name       string
	middleware []func(http.Handler) http.Handler
}

// NewMiddlewareGroup creates a new middleware group
func NewMiddlewareGroup(name string) *MiddlewareGroup {
	return &MiddlewareGroup{
		name:       name,
		middleware: make([]func(http.Handler) http.Handler, 0),
	}
}

// Add adds middleware to the group
func (mg *MiddlewareGroup) Add(middleware ...func(http.Handler) http.Handler) *MiddlewareGroup {
	mg.middleware = append(mg.middleware, middleware...)
	return mg
}

// Middleware returns all middleware in the group
func (mg *MiddlewareGroup) Middleware() []func(http.Handler) http.Handler {
	return mg.middleware
}

// MiddlewareGroups manages all middleware groups
type MiddlewareGroups struct {
	groups map[string]*MiddlewareGroup
}

// NewMiddlewareGroups creates a new middleware groups manager
func NewMiddlewareGroups() *MiddlewareGroups {
	mg := &MiddlewareGroups{
		groups: make(map[string]*MiddlewareGroup),
	}

	// Initialize default groups
	mg.InitDefaultGroups()
	return mg
}

// InitDefaultGroups initializes default middleware groups
func (mg *MiddlewareGroups) InitDefaultGroups() {
	// Base group (used by all routes)
	mg.AddGroup("base", NewMiddlewareGroup("base").Add(
		RequestID,
		ResponseMetadata,
	))

	// API group (used by all API routes)
	mg.AddGroup("api", NewMiddlewareGroup("api").Add(
		RequestID,
		ResponseMetadata,
		BodyParser,
	))

	// Public group (used by public routes)
	mg.AddGroup("public", NewMiddlewareGroup("public").Add(
		RequestID,
		ResponseMetadata,
	))

	// Protected group (used by authenticated routes)
	mg.AddGroup("protected", NewMiddlewareGroup("protected").Add(
		RequestID,
		ResponseMetadata,
		BodyParser,
		// Auth middleware will be added later
	))
}

// AddGroup adds a middleware group
func (mg *MiddlewareGroups) AddGroup(name string, group *MiddlewareGroup) {
	mg.groups[name] = group
}

// GetGroup returns a middleware group by name
func (mg *MiddlewareGroups) GetGroup(name string) *MiddlewareGroup {
	return mg.groups[name]
}

// ApplyGroup applies a middleware group to a route
func (mg *MiddlewareGroups) ApplyGroup(route *Route, groupName string) *Route {
	group := mg.GetGroup(groupName)
	if group != nil {
		for _, m := range group.Middleware() {
			route.Use(m)
		}
	}
	return route
}
