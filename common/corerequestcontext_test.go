package common

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anz-bank/sysl-go/log"

	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
)

func TestCoreRequestContextMiddleware(t *testing.T) {
	ctx := testutil.NewTestContext()
	mware := CoreRequestContextMiddleware
	body := bytes.NewBufferString("test")
	req, err := http.NewRequest("GET", "localhost/", body)
	require.Nil(t, err)
	req = req.WithContext(ctx)

	mware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		require.NotNil(t, log.GetLogger(r.Context()))
	})).ServeHTTP(nil, req)
}

func TestCoreRequestContextMiddleWare_VerboseLogging_LogRequestHeaderAndResponseHeader(t *testing.T) {
	ctx, logger := testutil.NewTestContextWithLogger(
		testutil.WithLogLevel(log.DebugLevel),
		testutil.WithLogPayloadContents(true))
	mware := CoreRequestContextMiddleware
	body := bytes.NewBufferString("test")
	req, err := http.NewRequest("GET", "localhost/", body)
	require.Nil(t, err)
	req = req.WithContext(ctx)
	fn := mware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {}))
	w := httptest.NewRecorder()
	defer func() {
		// log entry include req header and resp header
		require.Equal(t, 3, logger.EntryCount())
		require.True(t, strings.Contains(logger.Entries()[1].Message, "Request: header"))
		require.True(t, strings.Contains(logger.Entries()[2].Message, "Response: header"))
	}()
	fn.ServeHTTP(w, req)
}

func TestCoreRequestContextMiddleWare_NoVerboseLogging_NotLogRequestHeaderAndResponseHeader(t *testing.T) {
	ctx, logger := testutil.NewTestContextWithLogger(
		testutil.WithLogPayloadContents(true))
	mware := CoreRequestContextMiddleware
	body := bytes.NewBufferString("test")
	req, err := http.NewRequest("GET", "localhost/", body)
	require.Nil(t, err)
	req = req.WithContext(ctx)
	fn := mware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {}))
	w := httptest.NewRecorder()
	defer func() {
		// log entry does not include req header and resp header
		require.Equal(t, 1, logger.EntryCount())
	}()
	fn.ServeHTTP(w, req)
}

func TestSetAndGetNilRequestHeader(t *testing.T) {
	out := RequestHeaderFromContext(RequestHeaderToContext(context.Background(), nil))
	require.Equal(t, http.Header(nil), out)
}

func TestSetAndGetValidRequesetHeader(t *testing.T) {
	in := make(http.Header)
	in["Accept-Header"] = []string{"b"}

	out := RequestHeaderFromContext(RequestHeaderToContext(context.Background(), in))
	require.Equal(t, "b", out["Accept-Header"][0])
}

func TestGetUnsetRequestHeader(t *testing.T) {
	out := RequestHeaderFromContext(context.Background())
	require.Equal(t, http.Header(nil), out)
}

func TestGetUnsetResponseHeaderAndStatus(t *testing.T) {
	header, status := RespHeaderAndStatusFromContext(context.Background())
	require.Equal(t, http.Header(nil), header)
	require.Equal(t, http.StatusOK, status)
}

func TestSetAndGetNilResponseHeader(t *testing.T) {
	header, status := RespHeaderAndStatusFromContext(RespHeaderAndStatusToContext(context.Background(), nil, 0))
	require.Equal(t, http.Header(nil), header)
	require.Equal(t, 0, status)
}

func TestSetAndGetValidResponseHeader(t *testing.T) {
	inHeader := make(http.Header)
	inHeader["Accept-Header"] = []string{"b"}
	inStatus := http.StatusOK

	outHeader, outStatus := RespHeaderAndStatusFromContext(RespHeaderAndStatusToContext(context.Background(), inHeader, inStatus))
	require.Equal(t, "b", outHeader["Accept-Header"][0])
	require.Equal(t, http.StatusOK, outStatus)
}

func TestUpdateResponseStatus(t *testing.T) {
	inHeader := make(http.Header)
	inHeader["Accept-Header"] = []string{"b"}
	inStatus := http.StatusOK
	ctx := RespHeaderAndStatusToContext(context.Background(), inHeader, inStatus)

	outHeader, outStatus := RespHeaderAndStatusFromContext(ctx)
	require.Equal(t, "b", outHeader["Accept-Header"][0])
	require.Equal(t, http.StatusOK, outStatus)

	// create a new context and update the status using the new context
	ctxNew := RequestHeaderToContext(ctx, make(http.Header))

	err := UpdateResponseStatus(ctxNew, http.StatusAccepted)
	require.NoError(t, err)

	// check the status has been updated in the old context
	_, outStatus = RespHeaderAndStatusFromContext(ctx)
	require.Equal(t, http.StatusAccepted, outStatus)
}
