package internal

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anz-bank/sysl-go/logconfig"
	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
)

func TestLogger_FlushLog(t *testing.T) {
	ctx, hook := testutil.NewTestContextWithLoggerHook()

	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	l, _ := NewRequestLogger(ctx, req)

	resp := http.Response{
		Status:           "",
		StatusCode:       200,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           http.Header{},
		Body:             ioutil.NopCloser(&bytes.Buffer{}),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}
	l.LogResponse(&resp)

	require.Equal(t, 2, len(hook.Entries))

	l.FlushLog()
	require.Equal(t, 3, len(hook.Entries))
	require.Equal(t, "Already flushed the request", hook.LastEntry().Message)
}

func TestRequestLogger_NilBody(t *testing.T) {
	ctx, _ := testutil.NewTestContextWithLoggerHook()

	req, err := http.NewRequest("DELETE", "http://example.com/foo", nil)
	require.NoError(t, err)

	require.NotPanics(t, func() {
		NewRequestLogger(ctx, req)
	})
}

func TestRequestLogger_ResponseWriter(t *testing.T) {
	ctx, hook := testutil.NewTestContextWithLoggerHook()

	ctx = logconfig.SetVerboseLogging(ctx, false)
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	l, _ := NewRequestLogger(ctx, req)
	rw := l.ResponseWriter(httptest.NewRecorder())

	_, _ = rw.Write([]byte("hello"))
	l.FlushLog()
	require.Empty(t, hook.Entries)
}
