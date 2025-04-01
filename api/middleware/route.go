package middleware

import (
	"net/http"
	"strings"

	"github.com/rendyspratama/digital-discovery/api/versioning"
)

// RouteMiddleware wraps a handler with specific middleware
type RouteMiddleware struct {
	handler    http.Handler
	middleware []func(http.Handler) http.Handler
}

// NewRouteMiddleware creates a new route middleware wrapper
func NewRouteMiddleware(handler http.Handler) *RouteMiddleware {
	return &RouteMiddleware{
		handler:    handler,
		middleware: make([]func(http.Handler) http.Handler, 0),
	}
}

// Use adds middleware to the route
func (rm *RouteMiddleware) Use(middleware func(http.Handler) http.Handler) *RouteMiddleware {
	rm.middleware = append(rm.middleware, middleware)
	return rm
}

// Handler returns the final handler with all middleware applied
func (rm *RouteMiddleware) Handler() http.Handler {
	handler := rm.handler
	// Apply middleware in reverse order (last added is innermost)
	for i := len(rm.middleware) - 1; i >= 0; i-- {
		handler = rm.middleware[i](handler)
	}
	return handler
}

// Route represents a route with its handler and middleware
type Route struct {
	Path       string
	Methods    []string
	Handler    http.HandlerFunc
	Middleware []func(http.Handler) http.Handler
	Version    versioning.Version
}

// NewRoute creates a new route
func NewRoute(path string, methods []string, handler http.HandlerFunc) *Route {
	return &Route{
		Path:       path,
		Methods:    methods,
		Handler:    handler,
		Middleware: make([]func(http.Handler) http.Handler, 0),
		Version:    versioning.Version{Major: 1, Minor: 0}, // Default to v1.0
	}
}

// WithVersion sets the version for the route
func (r *Route) WithVersion(version versioning.Version) *Route {
	r.Version = version
	return r
}

// Use adds middleware to the route
func (r *Route) Use(middleware func(http.Handler) http.Handler) *Route {
	r.Middleware = append(r.Middleware, middleware)
	return r
}

// Router handles route registration and middleware
type Router struct {
	routes          map[string]*Route
	versionedRoutes *versioning.VersionedRoutes
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes:          make(map[string]*Route),
		versionedRoutes: versioning.NewVersionedRoutes(),
	}
}

// Register registers a route
func (r *Router) Register(route *Route) *Router {
	// Add to versioned routes
	r.versionedRoutes.AddRoute(route.Path, route.Version, route.Handler)

	// Add to regular routes for backward compatibility
	r.routes[route.Path] = route
	return r
}

// Handler returns the final handler with all routes and middleware
func (r *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	// Handle versioned routes
	mux.HandleFunc("/api/", func(w http.ResponseWriter, req *http.Request) {
		// Get version from request
		version, err := versioning.VersionFromRequest(req)
		if err != nil {
			http.Error(w, "Invalid API version", http.StatusBadRequest)
			return
		}

		// Get the appropriate handler
		handler, err := r.versionedRoutes.GetHandler(req.URL.Path, version)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Create route-specific middleware wrapper
		routeHandler := NewRouteMiddleware(http.HandlerFunc(handler))
		for _, middleware := range r.routes[req.URL.Path].Middleware {
			routeHandler.Use(middleware)
		}

		// Check if method is allowed
		methodAllowed := false
		for _, method := range r.routes[req.URL.Path].Methods {
			if method == req.Method {
				methodAllowed = true
				break
			}
		}

		if !methodAllowed {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		routeHandler.Handler().ServeHTTP(w, req)
	})

	// Handle non-versioned routes
	for path, route := range r.routes {
		if !strings.HasPrefix(path, "/api/") {
			// Create route-specific middleware wrapper
			routeHandler := NewRouteMiddleware(http.HandlerFunc(route.Handler))
			for _, middleware := range route.Middleware {
				routeHandler.Use(middleware)
			}

			// Register the route with its middleware
			mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// Check if method is allowed
				methodAllowed := false
				for _, method := range route.Methods {
					if method == req.Method {
						methodAllowed = true
						break
					}
				}

				if !methodAllowed {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				routeHandler.Handler().ServeHTTP(w, req)
			}))
		}
	}

	return mux
}
