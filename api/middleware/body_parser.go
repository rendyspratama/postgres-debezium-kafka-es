package middleware

import (
	"context"
	"io"
	"net/http"
)

func BodyParser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only parse body for POST and PUT requests
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// Read the body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Store the body in context
		ctx := context.WithValue(r.Context(), "requestBody", body)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
