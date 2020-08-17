package debug

import "github.com/go-chi/chi"

// Enable enables the debug service for the application.
func Enable(serviceName string, router chi.Router) {
	var metadata = &MetadataStore

	router.Use(NewDebugMiddleware(serviceName, metadata))
	router.Get("/-/trace", NewDebugIndexHandler(metadata))
	router.Get("/-/trace/{traceId}", NewDebugHandler(metadata))
}
