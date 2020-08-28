package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anz-bank/pkg/log"
)

type HTTPError struct {
	HTTPCode    int    `json:"-"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`

	extraFields map[string]interface{}
}

func (httpError *HTTPError) AddField(key string, val interface{}) {
	if httpError.extraFields == nil {
		httpError.extraFields = map[string]interface{}{}
	}
	httpError.extraFields[key] = val
}

type KV struct {
	K string
	V interface{}
}

type wrappedError struct {
	e      error
	fields []KV
}

func (w wrappedError) Error() string {
	return fmt.Sprintf("%s --- %+v", w.e.Error(), w.fields)
}

func WrappedError(err error, fields ...KV) error {
	return wrappedError{
		e:      err,
		fields: fields,
	}
}

type httpErrorResponse struct {
	Status interface{} `json:"status"`
}

func (httpError *HTTPError) WriteError(ctx context.Context, w http.ResponseWriter) {
	var marshalTarget interface{}

	marshalTarget = httpError
	if len(httpError.extraFields) > 0 {
		if httpError.Code != "" {
			httpError.extraFields["code"] = httpError.Code
		}
		if httpError.Description != "" {
			httpError.extraFields["description"] = httpError.Description
		}
		marshalTarget = httpError.extraFields
	}

	b, err := json.Marshal(httpErrorResponse{marshalTarget})
	if err != nil {
		log.Error(ctx, err)
		b = []byte(`{"status":{"code": "1234", "description": "Unknown Error"}}`)
		httpError.HTTPCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(httpError.HTTPCode)

	// Ignore write error, if any, as it is probably a client issue.
	_, _ = w.Write(b)
}
