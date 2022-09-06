package jwthttp

import "net/http"

// Commonly used handler cases
// To be added to over time as more use cases arise

// DefaultUnauthHandler is used when a middleware is created without a custom
// unauth handler
//
// If error defines its own http status, use that, otherwise respond 403 forbidden.
//
// Does not attach a body to the response.
func DefaultUnauthHandler(w http.ResponseWriter, r *http.Request, err error) {
	// TODO: log the error
	if s, ok := err.(interface{ HTTPStatus() int }); ok {
		w.WriteHeader(s.HTTPStatus())
		return
	}
	w.WriteHeader(http.StatusForbidden)
}

// HiddenEndpoint is an unauth handler that hides the existence of the protected resource
// Does not send any error details.
func HiddenEndpoint(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusNotFound)
}
