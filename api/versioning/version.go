package versioning

import (
	"fmt"
	"net/http"
	"strings"
)

// Version represents an API version
type Version struct {
	Major int
	Minor int
}

// String returns the version string in format "v1.0"
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
}

// ParseVersion parses a version string into a Version struct
func ParseVersion(version string) (Version, error) {
	if !strings.HasPrefix(version, "v") {
		return Version{}, fmt.Errorf("invalid version format: %s", version)
	}

	var major, minor int
	_, err := fmt.Sscanf(version[1:], "%d.%d", &major, &minor)
	if err != nil {
		return Version{}, fmt.Errorf("invalid version format: %s", version)
	}

	return Version{Major: major, Minor: minor}, nil
}

// VersionFromRequest extracts version from request path or header
func VersionFromRequest(r *http.Request) (Version, error) {
	// First try to get version from path
	path := r.URL.Path
	if strings.HasPrefix(path, "/api/") {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			if version, err := ParseVersion(parts[2]); err == nil {
				return version, nil
			}
		}
	}

	// Then try to get version from header
	version := r.Header.Get("X-API-Version")
	if version != "" {
		return ParseVersion(version)
	}

	// Default to latest version
	return Version{Major: 1, Minor: 0}, nil
}

// VersionedHandler wraps a handler with version information
type VersionedHandler struct {
	Version Version
	Handler http.HandlerFunc
}

// VersionedRoutes manages versioned routes
type VersionedRoutes struct {
	routes map[string][]VersionedHandler
}

// NewVersionedRoutes creates a new versioned routes manager
func NewVersionedRoutes() *VersionedRoutes {
	return &VersionedRoutes{
		routes: make(map[string][]VersionedHandler),
	}
}

// AddRoute adds a versioned route
func (vr *VersionedRoutes) AddRoute(path string, version Version, handler http.HandlerFunc) {
	vr.routes[path] = append(vr.routes[path], VersionedHandler{
		Version: version,
		Handler: handler,
	})
}

// GetHandler returns the appropriate handler for the request version
func (vr *VersionedRoutes) GetHandler(path string, version Version) (http.HandlerFunc, error) {
	handlers, exists := vr.routes[path]
	if !exists {
		return nil, fmt.Errorf("no handlers found for path: %s", path)
	}

	// Find the best matching version
	var bestHandler http.HandlerFunc
	var bestVersion Version
	found := false

	for _, h := range handlers {
		if h.Version.Major == version.Major {
			if h.Version.Minor <= version.Minor {
				if !found || h.Version.Minor > bestVersion.Minor {
					bestHandler = h.Handler
					bestVersion = h.Version
					found = true
				}
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("no compatible version found for %s", version.String())
	}

	return bestHandler, nil
}
