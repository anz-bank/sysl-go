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

var ctx, _ = context.WithTimeout(context.Background(), 0*time.Nanosecond)
var err = fmt.Errorf("nothing")
var serverErrors = []struct {
	ctx  context.Context
	kind Kind
	msg  string
	in   error
	out  Kind
}{
	{ctx, UnknownError, "test", err, DownstreamTimeoutError},
	{context.Background(), UnknownError, "test", err, UnknownError},
	{context.Background(), BadRequestError, "test", err, BadRequestError},
	{context.Background(), InternalError, "test", err, InternalError},
	{context.Background(), UnauthorizedError, "test", err, UnauthorizedError},
	{context.Background(), DownstreamUnavailableError, "test", err, DownstreamUnavailableError},
	{context.Background(), DownstreamTimeoutError, "test", err, DownstreamTimeoutError},
	{context.Background(), BadRequestError, "test", &ServerError{Kind: DownstreamTimeoutError, Message: "test", Cause: err}, DownstreamTimeoutError},
	{context.Background(), DownstreamUnauthorizedError, "test", &ServerError{Kind: DownstreamUnauthorizedError, Message: "test", Cause: err}, DownstreamUnauthorizedError},
	{context.Background(), DownstreamUnexpectedResponseError, "test", &ServerError{Kind: DownstreamUnexpectedResponseError, Message: "test", Cause: err}, DownstreamUnexpectedResponseError},
}

func TestServerErrorCreateError(t *testing.T) {
	req := require.New(t)
	for _, t := range serverErrors {
		req.EqualError(&ServerError{Kind: t.out, Message: t.msg, Cause: err}, CreateError(t.ctx, t.kind, "test", t.in).Error())
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
	defer resp.Body.Close()
	resp.Request = &http.Request{
		Method: "PUT",
		URL: &url.URL{
			Scheme: "https",
			Host:   "www.test.com",
			Path:   "hello",
		},
	}

	// When
	e := CreateDownstreamError(ctx, DownstreamUnexpectedResponseError, resp, nil, err)

	// Then
	require.IsType(t, &ServerError{}, e)
	require.Implements(t, (*ErrorKinder)(nil), e)
	require.Equal(t, e.(ErrorKinder).ErrorKind(), DownstreamTimeoutError)
	require.EqualError(t, e, "ServerError(Kind=Time out from down stream services, Message=PUT https://www.test.com/hello, Cause=nothing)")
}

func TestDownstreamError_CreateDownstreamError_UnexpectedResponse(t *testing.T) {
	// Given
	b := `{"status": {"code": "1234", description: "unknown error"}}`
	r := httptest.NewRecorder()
	r.Header().Set("Content-Type", "application/json")
	r.Header().Set("Content-Length", strconv.Itoa(len(b)))
	r.WriteHeader(http.StatusInternalServerError)

	resp := r.Result()
	defer resp.Body.Close()
	resp.Request = &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "https",
			Host:   "www.test.com",
			Path:   "hello",
		},
	}

	// When
	e := CreateDownstreamError(context.Background(), DownstreamUnexpectedResponseError, resp, []byte(b), err)

	// Then
	require.IsType(t, &DownstreamError{}, e)
	require.Implements(t, (*ErrorKinder)(nil), e)
	require.Equal(t, e.(ErrorKinder).ErrorKind(), DownstreamUnexpectedResponseError)
	require.EqualError(t, e, "DownstreamError(Kind=Unexpected response from downstream services, Method=POST, URL=https://www.test.com/hello, StatusCode=500, ContentType=application/json, ContentLength=58, Snippet={\"status\": {\"code\": \"1234\", description: \"unknown error\"}}, Cause=nothing)")
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
	defer resp.Body.Close()
	resp.Request = &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "https",
			Host:   "www.test.com",
			Path:   "hello",
		},
	}

	// When
	e := CreateDownstreamError(context.Background(), DownstreamUnauthorizedError, resp, []byte(b), err)

	// Then
	require.IsType(t, &DownstreamError{}, e)
	require.Implements(t, (*ErrorKinder)(nil), e)
	require.Equal(t, e.(ErrorKinder).ErrorKind(), DownstreamUnauthorizedError)
	require.EqualError(t, e, "DownstreamError(Kind=Unauthorized error from downstream services, Method=GET, URL=https://www.test.com/hello, StatusCode=401, ContentType=text/plain, ContentLength=159, Snippet=This is a very very long response body.\nThis is a very very long response body.\nThis is a very very long response body.\nThis is , Cause=nothing)")
}
