package metrics

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
)

// ProxyResponseWriter is a proxy for a http.ResponseWriter allowing you to
// access the various parts of the response at a later date.
type ProxyResponseWriter interface {
	http.ResponseWriter
	Status() int
}

// StatusResponseWriter implements the http.ResponseWriter interface and allows
// us access to the response status through the .Status() function.
type StatusResponseWriter struct {
	http.ResponseWriter
	code        int  // Response status
	wroteHeader bool // Check if header has been assigned
}

// WriteHeader captures the assigned response code for access at a later time,
// and writes the response header accordingly.
func (w *StatusResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.code = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

// Write is a wrapper around the ResponseWriter.Write function call which checks
// to see if the Status has been written, and defaults to http.StatusOK if not.
func (w *StatusResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

// Status returns the status assigned to the outgoing http response.
func (w *StatusResponseWriter) Status() int {
	return w.code
}

// NewStatusResponseWriter returns a ProxyResponseWriter which allows us to
// access the outgoing response status.
func NewStatusResponseWriter(w http.ResponseWriter) ProxyResponseWriter {
	return &StatusResponseWriter{ResponseWriter: w}
}

// GetChiPathPattern will use the chi context to return the Route
// Pattern associated with the provided requests context.
func GetChiPathPattern(ctx context.Context) string {
	return chi.RouteContext(ctx).RoutePattern()
}
