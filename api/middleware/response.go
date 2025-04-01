package middleware

import "net/http"

// responseWriter is a wrapper around http.ResponseWriter that captures status and body
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        []byte
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	rw.body = b
	return rw.ResponseWriter.Write(b)
}
