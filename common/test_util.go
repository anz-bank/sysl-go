package common

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/anz-bank/sysl-go/log"

	"github.com/stretchr/testify/mock"
)

func NewString(s string) *string {
	return &s
}

func NewBool(b bool) *bool {
	return &b
}

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*http.Response), args.Error(1)
}

// NewHTTPTestServer returns a new httptest.Server with the given handler suitable for use within
// unit tests. The returned server comes equipped with the following:
// 1. log.Logger.
func NewHTTPTestServer(handler http.Handler) *httptest.Server {
	ts := NewUnstartedHTTPTestServer(handler)
	ts.Start()
	return ts
}

// NewHTTPTestServer returns an unstarted httptest.Server with the given handler suitable for use
// within unit tests. The returned server comes equipped with the following:
// 1. log.Logger.
func NewUnstartedHTTPTestServer(handler http.Handler) *httptest.Server {
	ts := httptest.NewUnstartedServer(handler)
	ts.Config.BaseContext = func(net.Listener) context.Context { return log.PutLogger(context.Background(), log.NewDefaultLogger()) }
	return ts
}
