// Package mock contains testing utilities and mocks shared between
// packages.
package restlib

import (
	"net/http"
)

// ResponseWriter returns a mock ResponseWriter implementing http.ResponseWriter.
func ResponseWriter() *responseWriter { //nolint:golint
	return &responseWriter{}
}

// responseWriter implements http.ResponseWriter.
type responseWriter struct {
	LastBody   string
	LastStatus int

	header http.Header
	err    error
}

// Err sets the error value returned by the Write method.
// Chain it with construction: mock.ResponseWriter().Err(someErr).
func (m *responseWriter) Err(err error) *responseWriter {
	m.err = err
	return m
}

// Header returns the responseWriters HTTP header.
func (m *responseWriter) Header() http.Header {
	if m.header == nil {
		m.header = http.Header{}
	}
	return m.header
}

// WriteHeader writes HTTP Status code to struct field.
func (m *responseWriter) WriteHeader(status int) {
	m.LastStatus = status
}

// Write writes response body to struct field.
func (m *responseWriter) Write(b []byte) (int, error) {
	m.LastBody = string(b)
	return len(b), m.err
}

type readCloser struct{ err error }

func (r *readCloser) Read(_ []byte) (int, error) {
	return 0, r.err
}
func (r *readCloser) Close() error {
	return r.err
}
func (r *readCloser) Err(err error) *readCloser {
	r.err = err
	return r
}

// ReadCloser returns a mock io.ReadCloser that can be set to return
// an err when Read is called on it and do nothing when Close is called
// on it.
func ReadCloser() *readCloser { //nolint:golint
	return &readCloser{}
}
