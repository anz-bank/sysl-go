package debug

import (
	"net/http"
	"time"
)

// CapturingResponseWriter wraps a delegate ResponseWriter and records writes to it.
type CapturingResponseWriter struct {
	w          *http.ResponseWriter
	body       string
	statusCode int
}

// NewCapturingResponseWriter returns a new CapturingResponseWriter.
func NewCapturingResponseWriter(delegate *http.ResponseWriter) CapturingResponseWriter {
	return CapturingResponseWriter{w: delegate}
}

func (w *CapturingResponseWriter) Header() http.Header {
	return (*w.w).Header()
}

func (w *CapturingResponseWriter) Write(b []byte) (int, error) {
	w.body = string(b)
	return (*w.w).Write(b)
}

func (w *CapturingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	(*w.w).WriteHeader(statusCode)
}

// NewDebugMiddleware returns a new middleware to record debug metadata for requests and responses.
func NewDebugMiddleware(metadata *Metadata) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			cw := NewCapturingResponseWriter(&w)
			next.ServeHTTP(&cw, r)
			elapsed := time.Since(start)

			metadata.Record(r, cw.body, elapsed)
		})
	}
}
