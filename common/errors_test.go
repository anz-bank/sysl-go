package common

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testMsg     = "test"
	httpsScheme = "https"
	testHost    = "www.test.com"
	testPath    = "hello"
)

var ctx, _ = context.WithTimeout(context.Background(), 0*time.Nanosecond)
var err = fmt.Errorf("nothing")
var serverErrors = []struct {
	ctx  context.Context
	kind Kind
	msg  string
	in   error
	out  Kind
}{
	{ctx, UnknownError, testMsg, err, DownstreamTimeoutError},
	{context.Background(), UnknownError, testMsg, err, UnknownError},
	{context.Background(), BadRequestError, testMsg, err, BadRequestError},
	{context.Background(), InternalError, testMsg, err, InternalError},
	{context.Background(), UnauthorizedError, testMsg, err, UnauthorizedError},
	{context.Background(), DownstreamUnavailableError, testMsg, err, DownstreamUnavailableError},
	{context.Background(), DownstreamTimeoutError, testMsg, err, DownstreamTimeoutError},
	{context.Background(), BadRequestError, testMsg, &ServerError{Kind: DownstreamTimeoutError, Message: testMsg, Cause: err}, DownstreamTimeoutError},
	{context.Background(), DownstreamUnauthorizedError, testMsg, &ServerError{Kind: DownstreamUnauthorizedError, Message: testMsg, Cause: err}, DownstreamUnauthorizedError},
	{context.Background(), DownstreamUnexpectedResponseError, testMsg, &ServerError{Kind: DownstreamUnexpectedResponseError, Message: testMsg, Cause: err}, DownstreamUnexpectedResponseError},
}

func TestServerErrorCreateError(t *testing.T) {
	req := require.New(t)
	for _, t := range serverErrors {
		req.EqualError(&ServerError{Kind: t.out, Message: t.msg, Cause: err}, CreateError(t.ctx, t.kind, testMsg, t.in).Error())
	}
}

func TestServerError_ErrorClass(t *testing.T) {
	e := CreateError(context.Background(), DownstreamUnavailableError, "message", nil)
	require.NotPanics(t, func() {
		_ = e.(ErrorKinder)
	})
}

func TestDownstreamError_CreateDownstreamError_Timeout(t *testing.T) {
	// Given
	r := httptest.NewRecorder()
	r.WriteHeader(http.StatusConflict)

	resp := r.Result()
	resp.Request = &http.Request{
		Method: "PUT",
		URL: &url.URL{
			Scheme: httpsScheme,
			Host:   testHost,
			Path:   testPath,
		},
	}

	// When
	e := CreateDownstreamError(ctx, DownstreamUnexpectedResponseError, resp, nil, err)

	// Then
	require.IsType(t, &ServerError{}, e)
	require.Implements(t, (*ErrorKinder)(nil), e)
	require.Equal(t, e.(ErrorKinder).ErrorKind(), DownstreamTimeoutError)
	require.EqualError(t, e, "ServerError(Kind=Time out from down stream services, Message=PUT https://www.test.com/hello, Cause=nothing)")
	defer resp.Body.Close()
}

func TestDownstreamError_CreateDownstreamError_UnexpectedResponse(t *testing.T) {
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
			Scheme: httpsScheme,
			Host:   testHost,
			Path:   testPath,
		},
	}

	// When
	e := CreateDownstreamError(context.Background(), DownstreamUnexpectedResponseError, resp, []byte(b), err)

	// Then
	require.IsType(t, &DownstreamError{}, e)
	require.Implements(t, (*ErrorKinder)(nil), e)
	require.Equal(t, e.(ErrorKinder).ErrorKind(), DownstreamUnexpectedResponseError)
	require.EqualError(t, e, "DownstreamError(Kind=Unexpected response from downstream services, Method=POST, URL=https://www.test.com/hello, StatusCode=500, ContentType=application/json, ContentLength=58, Snippet={\"status\": {\"code\": \"1234\", description: \"unknown error\"}}, Cause=nothing)")
	defer resp.Body.Close()
}

func TestDownstreamError_CreateDownstreamError_Unauthorized(t *testing.T) {
	// Given
	b := `This is a very very long response body.
This is a very very long response body.
This is a very very long response body.
This is a very very long response body.`
	r := httptest.NewRecorder()
	r.Header().Set("Content-Type", "text/plain")
	r.Header().Set("Content-Length", strconv.Itoa(len(b)))
	r.WriteHeader(http.StatusUnauthorized)

	resp := r.Result()
	resp.Request = &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: httpsScheme,
			Host:   testHost,
			Path:   testPath,
		},
	}

	// When
	e := CreateDownstreamError(context.Background(), DownstreamUnauthorizedError, resp, []byte(b), err)

	// Then
	require.IsType(t, &DownstreamError{}, e)
	require.Implements(t, (*ErrorKinder)(nil), e)
	require.Equal(t, e.(ErrorKinder).ErrorKind(), DownstreamUnauthorizedError)
	require.EqualError(t, e, "DownstreamError(Kind=Unauthorized error from downstream services, Method=GET, URL=https://www.test.com/hello, StatusCode=401, ContentType=text/plain, ContentLength=159, Snippet=This is a very very long response body.\nThis is a very very long response body.\nThis is a very very long response body.\nThis is , Cause=nothing)")
	defer resp.Body.Close()
}
func TestCheckContextTimeout_ContextDeadlineExceeded(t *testing.T) {
	// Given a context with deadline exceeded
	ctx, cancel := context.WithTimeout(context.Background(), 0*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure context timeout fires

	testErr := fmt.Errorf("some error")

	// When
	err := CheckContextTimeout(ctx, "test message", testErr)

	// Then
	require.NotNil(t, err)
	serverErr, ok := err.(*ServerError)
	require.True(t, ok, "expected ServerError")
	require.Equal(t, DownstreamTimeoutError, serverErr.Kind)
	require.Equal(t, "test message", serverErr.Message)
	require.Equal(t, testErr, serverErr.Cause)
}

func TestCheckContextTimeout_HTTPClientTimeout(t *testing.T) {
	// Given a valid context but an HTTP client timeout error
	ctx := context.Background()

	// Create a mock timeout error (simulates Client.Timeout exceeded)
	timeoutErr := &mockTimeoutError{msg: "Client.Timeout exceeded"}

	// When
	err := CheckContextTimeout(ctx, "test message", timeoutErr)

	// Then
	require.NotNil(t, err)
	serverErr, ok := err.(*ServerError)
	require.True(t, ok, "expected ServerError")
	require.Equal(t, DownstreamTimeoutError, serverErr.Kind)
	require.Equal(t, "test message", serverErr.Message)
	require.Equal(t, timeoutErr, serverErr.Cause)
}

func TestCheckContextTimeout_NoTimeout(t *testing.T) {
	// Given a valid context and a non-timeout error
	ctx := context.Background()
	testErr := fmt.Errorf("regular error")

	// When
	err := CheckContextTimeout(ctx, "test message", testErr)

	// Then
	require.Nil(t, err, "expected nil for non-timeout errors")
}

func TestCheckContextTimeout_NilError(t *testing.T) {
	// Given a valid context and nil error
	ctx := context.Background()

	// When
	err := CheckContextTimeout(ctx, "test message", nil)

	// Then
	require.Nil(t, err, "expected nil for nil error")
}

// mockTimeoutError simulates an HTTP client timeout error.
type mockTimeoutError struct {
	msg string
}

func (e *mockTimeoutError) Error() string {
	return e.msg
}

func (e *mockTimeoutError) Timeout() bool {
	return true
}

func (e *mockTimeoutError) Temporary() bool {
	return true
}
