package internal

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/sirupsen/logrus/hooks/test"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/mock"
)

type mockRountTripper struct {
	mock.Mock
}

func (m *mockRountTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	rsp := args.Get(0).(*http.Response)
	rsp.Request = req
	rsp.Body = ioutil.NopCloser(bytes.NewBufferString("test"))
	return rsp, args.Error(1)
}

func TestLoggingRoundtripper(t *testing.T) {
	logger, _ := test.NewNullLogger()
	base := mockRountTripper{}
	base.On("RoundTrip", mock.Anything).Return(&http.Response{}, nil)

	tt := NewLoggingRoundTripper(logger.WithContext(context.Background()), &base)

	body := bytes.NewBufferString("test")
	req, err := http.NewRequest("GET", "localhost/", body)
	require.NoError(t, err)
	req.RemoteAddr = "127.0.0.1"
	req.URL, err = url.Parse("http://www.example.com")
	require.NoError(t, err)

	res, err := tt.RoundTrip(req)
	require.NoError(t, err)
	res.Body.Close()

	base.AssertCalled(t, "RoundTrip", req)
}

type testRoundtripper struct {
	wasCalled  bool
	statusCode int
}

func (r *testRoundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.wasCalled = true

	rsp := http.Response{
		StatusCode: r.statusCode,
		Request:    req,
		Body:       ioutil.NopCloser(bytes.NewBufferString("resp body")),
	}

	return &rsp, nil
}
func TestLoggingTransport_RoundTrip400Code(t *testing.T) {
	tr := testRoundtripper{false, 400}
	logger, hook := test.NewNullLogger()
	logger.Level = logrus.DebugLevel

	transport := NewLoggingRoundTripper(logger.WithContext(context.Background()), &tr)
	body := bytes.NewBufferString("test")
	req, err := http.NewRequest("POST", "http://localhost:1234/", body)
	require.NoError(t, err)

	res, err := transport.RoundTrip(req)
	require.NoError(t, err)
	res.Body.Close()
	require.True(t, tr.wasCalled)

	debugCount := 0
	reqFound := false
	respFound := false
	statusFound := false
	for _, entry := range hook.Entries {
		if entry.Level == logrus.DebugLevel {
			debugCount++
			if entry.Message == "Request Body: test" {
				reqFound = true
			}
			if entry.Message == "Response Body: resp body" {
				respFound = true
			}
		}
		if entry.Message == "Backend request completed" {
			statusFound = true
		}
	}
	require.Equal(t, 4, debugCount)
	require.True(t, reqFound)
	require.True(t, respFound)
	require.True(t, statusFound)
}

func TestLoggingTransport_RoundTripLogFields(t *testing.T) {
	tr := testRoundtripper{false, 400}
	logger, hook := test.NewNullLogger()
	logger.Level = logrus.DebugLevel

	transport := NewLoggingRoundTripper(logger.WithContext(context.Background()), &tr)
	body := bytes.NewBufferString("test")
	req, err := http.NewRequest("POST", "http://localhost:1234/", body)
	req.Header.Add(distributedTraceIDName, "this is trace id")
	req.Header.Add(distributedSpanIDName, "this is span id")
	require.NoError(t, err)

	res, err := transport.RoundTrip(req)
	require.NoError(t, err)
	res.Body.Close()
	require.True(t, tr.wasCalled)

	debugCount := 0
	reqFound := false
	respFound := false
	statusFound := false
	for _, entry := range hook.Entries {
		require.Equal(t, "this is trace id", entry.Data[distributedTraceIDName])
		require.Equal(t, "this is span id", entry.Data[distributedSpanIDName])
		require.Nil(t, entry.Data[distributedParentSpanIDName])
		if entry.Level == logrus.DebugLevel {
			debugCount++
			if entry.Message == "Request Body: test" {
				reqFound = true
			}
			if entry.Message == "Response Body: resp body" {
				respFound = true
			}
		}
		if entry.Message == "Backend request completed" {
			statusFound = true
		}
	}
	require.Equal(t, 4, debugCount)
	require.True(t, reqFound)
	require.True(t, respFound)
	require.True(t, statusFound)
}
