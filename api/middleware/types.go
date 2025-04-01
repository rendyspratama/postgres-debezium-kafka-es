package middleware

import "net/http"

// ResponseWriter is a wrapper around http.ResponseWriter that captures status and body
type ResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        []byte
}

// NewResponseWriter creates a new response writer
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w}
}

// WriteHeader implements http.ResponseWriter
func (rw *ResponseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// Write implements http.ResponseWriter
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	rw.body = b
	return rw.ResponseWriter.Write(b)
}
