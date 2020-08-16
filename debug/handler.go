package debug

import (
	"github.com/go-chi/chi"
	"net/http"
)

// NewDebugIndexHandler returns a handler for the debug index endpoint.
func NewDebugIndexHandler(metadata *Metadata) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Add("Content-Type", "application/json")
	_ = writeIndex(w, metadata)
	}
}

// NewDebugHandler returns a handler for the debug endpoint.
func NewDebugHandler(metadata *Metadata) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		traceId := chi.URLParam(r, "traceId")
		e := metadata.GetEntryByTrace(traceId)
		//w.Header().Add("Content-Type", "application/json")
		_ = writeTrace(w, &e)
	}
}
