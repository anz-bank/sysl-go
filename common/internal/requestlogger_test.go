package internal

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anz-bank/sysl-go/log"

	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
)

func TestLogger_FlushLog(t *testing.T) {
	ctx, logger := testutil.NewTestContextWithLogger(
		testutil.WithLogLevel(log.DebugLevel),
		testutil.WithLogPayloadContents(true))

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

	require.Equal(t, 2, logger.EntryCount())

	l.FlushLog()
	require.Equal(t, 3, logger.EntryCount())
	require.Equal(t, "Already flushed the request", logger.LastEntry().Message)
}

func TestRequestLogger_NilBody(t *testing.T) {
	ctx, _ := testutil.NewTestContextWithLogger()

	req, err := http.NewRequest("DELETE", "http://example.com/foo", nil)
	require.NoError(t, err)

	require.NotPanics(t, func() {
		NewRequestLogger(ctx, req)
	})
}

func TestRequestLogger_ResponseWriter(t *testing.T) {
	ctx, logger := testutil.NewTestContextWithLogger()

	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	require.NoError(t, err)

	l, _ := NewRequestLogger(ctx, req)
	rw := l.ResponseWriter(httptest.NewRecorder())

	_, _ = rw.Write([]byte("hello"))
	l.FlushLog()
	require.Zero(t, logger.EntryCount())
}
