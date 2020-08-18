package debug

import (
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"net/http"
)

// NewDebugIndexHandler returns a handler for the debug index endpoint.
func NewDebugIndexHandler(metadata *Metadata) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		err := writeIndex(w, *metadata)
		if err != nil {
			logrus.WithError(err).Error("render index failed")
			w.WriteHeader(500)
		}
	}
}

// NewDebugHandler returns a handler for the debug endpoint.
func NewDebugHandler(metadata *Metadata) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		traceId := chi.URLParam(r, "traceId")

		w.Header().Add("Content-Type", "text/html")
		err := writeTrace(w, traceId, *metadata)
		if err != nil {
			logrus.WithError(err).Error("render trace failed")
			w.WriteHeader(500)
		}
	}
}
