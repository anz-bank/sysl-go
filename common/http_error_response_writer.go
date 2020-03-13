package common

import (
	"context"
	"encoding/json"
	"net/http"
)

type HTTPError struct {
	HTTPCode    int    `json:"-"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

type httpErrorResponse struct {
	Status *HTTPError `json:"status"`
}

func (httpError *HTTPError) WriteError(ctx context.Context, w http.ResponseWriter) {
	logEntry := GetLogEntryFromContext(ctx)

	b, err := json.Marshal(httpErrorResponse{httpError})
	if err != nil {
		logEntry.Error(err)
		b = []byte(`{"status":{"code": "1234", "description": "Unknown Error"}}`)
		httpError.HTTPCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(httpError.HTTPCode)

	// Ignore write error, if any, as it is probably a client issue.
	_, _ = w.Write(b)
}
