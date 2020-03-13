// +build go1.13

package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServerError_Unwrap(t *testing.T) {
	// Given
	innerWrappedError := fmt.Errorf("inner wrapped: %w", err)
	e := CreateError(context.Background(), BadRequestError, "bad request", innerWrappedError)

	// When
	unwrappedErr1 := errors.Unwrap(e)
	unwrappedErr2 := errors.Unwrap(unwrappedErr1)

	// Then
	require.NotNil(t, unwrappedErr1)
	require.Error(t, unwrappedErr1)
	require.Equal(t, innerWrappedError, unwrappedErr1)
	require.NotNil(t, unwrappedErr2)
	require.Error(t, unwrappedErr2)
	require.Equal(t, err, unwrappedErr2)
}

func TestDownstreamError_Unwrap(t *testing.T) {
	// Given
	b := `{"status": {"code": "1234", description: "unknown error"}}`
	r := httptest.NewRecorder()
	r.Header().Set("Content-Type", "application/json")
	r.Header().Set("Content-Length", strconv.Itoa(len(b)))
	r.WriteHeader(http.StatusInternalServerError)

	resp := r.Result()
	resp.Request = &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "https",
			Host:   "www.test.com",
			Path:   "hello",
		},
	}

	innerWrappedError := fmt.Errorf("inner wrapped: %w", err)
	e := CreateDownstreamError(context.Background(), DownstreamUnexpectedResponseError, resp, []byte(b), innerWrappedError)

	// When
	unwrappedErr1 := errors.Unwrap(e)
	unwrappedErr2 := errors.Unwrap(unwrappedErr1)

	// Then
	require.NotNil(t, unwrappedErr1)
	require.Error(t, unwrappedErr1)
	require.Equal(t, innerWrappedError, unwrappedErr1)
	require.NotNil(t, unwrappedErr2)
	require.Error(t, unwrappedErr2)
	require.Equal(t, err, unwrappedErr2)
}

func TestServerError_Is_FmtErrorf(t *testing.T) {
	// Given
	e1 := errors.New("inner most error")
	e2 := fmt.Errorf("inner: %w", e1)
	e := CreateError(context.Background(), BadRequestError, "bad request", e2)

	// When
	isE2 := errors.Is(e, e2)
	isE1 := errors.Is(e, e1)

	// Then
	require.True(t, isE1)
	require.True(t, isE2)
}

type innerError struct {
	message string
	err     error
}

func (e *innerError) Error() string {
	return fmt.Sprintf("inner error: %s", e.message)
}

func (e *innerError) Is(err error) bool {
	_, ok := err.(*innerError)
	return ok
}

func (e *innerError) Unwrap() error {
	return e.err
}

func TestServerError_Is_CustomizedInner(t *testing.T) {
	// Given
	e1 := errors.New("inner most error")
	e2 := &innerError{
		message: "test",
		err:     e1,
	}
	e := CreateError(context.Background(), BadRequestError, "bad request", e2)

	// When
	isE2 := errors.Is(e, e2)
	isE1 := errors.Is(e, e1)

	// Then
	require.True(t, isE1)
	require.True(t, isE2)
}
