package middleware

import (
	"net/http"
	"time"
)

func ResponseMetadata(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Timestamp", time.Now().Format(time.RFC3339))
		if reqID := r.Context().Value("requestID"); reqID != nil {
			w.Header().Set("X-Request-ID", reqID.(string))
		}
		next.ServeHTTP(w, r)
	})
}
