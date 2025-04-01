package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rendyspratama/digital-discovery/api/config"
)

type CORSMiddleware struct {
	config config.MiddlewareConfig
}

func NewCORSMiddleware(cfg config.MiddlewareConfig) *CORSMiddleware {
	return &CORSMiddleware{config: cfg}
}

func (c *CORSMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set allowed origins
		origin := r.Header.Get("Origin")
		if origin != "" {
			allowed := false
			for _, allowedOrigin := range c.config.CORS.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		// Set allowed methods
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.config.CORS.AllowedMethods, ", "))

		// Set allowed headers
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.config.CORS.AllowedHeaders, ", "))

		// Set max age for preflight requests
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", c.config.CORS.MaxAge))

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
