package internal

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/anz-bank/pkg/log"

	"github.com/anz-bank/sysl-go/testutil"
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
	ctx, _ := testutil.NewTestContextWithLoggerHook()
	base := mockRountTripper{}
	base.On("RoundTrip", mock.Anything).Return(&http.Response{}, nil)

	tt := NewLoggingRoundTripper(ctx, &base)

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
	ctx, hook := testutil.NewTestContextWithLoggerHook()
	transport := NewLoggingRoundTripper(ctx, &tr)
	body := bytes.NewBufferString("test")
	req, err := http.NewRequest("POST", "http://localhost:1234/", body)
	require.NoError(t, err)

	response, err := transport.RoundTrip(req)
	require.NoError(t, err)
	require.True(t, tr.wasCalled)
	defer response.Body.Close()

	debugCount := 0
	reqFound := false
	respFound := false
	statusFound := false
	for _, entry := range hook.Entries {
		if entry.Verbose {
			debugCount++
			if entry.Message == "Response: header - map[]\nbody[len:9]: - resp body" {
				reqFound = true
			}
			if entry.Message == "Request: header - map[]\nbody[len:4]: - test" {
				respFound = true
			}
		}
		if entry.Message == "Backend request completed" {
			statusFound = true
		}
	}
	require.Equal(t, 2, debugCount)
	require.True(t, reqFound)
	require.True(t, respFound)
	require.True(t, statusFound)
}

func TestLoggingTransport_RoundTripLogFields(t *testing.T) {
	tr := testRoundtripper{false, 400}
	ctx, hook := testutil.NewTestContextWithLoggerHook()
	ctx = log.WithConfigs(log.SetVerboseMode(true)).Onto(ctx)
	transport := NewLoggingRoundTripper(ctx, &tr)
	body := bytes.NewBufferString("test")
	req, err := http.NewRequest("POST", "http://localhost:1234/", body)
	req.Header.Add(distributedTraceIDName, "this is trace id")
	req.Header.Add(distributedSpanIDName, "this is span id")
	require.NoError(t, err)

	response, err := transport.RoundTrip(req)
	require.NoError(t, err)
	require.True(t, tr.wasCalled)
	defer response.Body.Close()

	debugCount := 0
	reqFound := false
	respFound := false
	statusFound := false
	for _, entry := range hook.Entries {
		entryTraceIDName, _ := entry.Data.Get(distributedTraceIDName)
		entrySpanIDName, _ := entry.Data.Get(distributedSpanIDName)
		entryDistributedParentSpanIDName, _ := entry.Data.Get(distributedParentSpanIDName)
		require.Equal(t, "this is trace id", entryTraceIDName)
		require.Equal(t, "this is span id", entrySpanIDName)
		require.Nil(t, entryDistributedParentSpanIDName)
		if entry.Verbose {
			debugCount++
			if entry.Message == "Response: header - map[]\nbody[len:9]: - resp body" {
				reqFound = true
			}
			if entry.Message == "Request: header - map[X-B3-Spanid:[this is span id] X-B3-Traceid:[this is trace id]]\nbody[len:4]: - test" {
				respFound = true
			}
		}
		if entry.Message == "Backend request completed" {
			statusFound = true
		}
	}
	require.Equal(t, 2, debugCount)
	require.True(t, reqFound)
	require.True(t, respFound)
	require.True(t, statusFound)
}
