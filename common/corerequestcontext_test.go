package common

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreRequestContextMiddleware(t *testing.T) {
	_, _, ctx := NewTestCoreRequestContext()
	mware := CoreRequestContextMiddleware()
	body := bytes.NewBufferString("test")
	req, err := http.NewRequest("GET", "localhost/", body)
	require.Nil(t, err)
	req = req.WithContext(ctx)

	fn := mware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		coreCtx := getCoreContext(r.Context())
		require.NotNil(t, coreCtx)
	}))

	fn.ServeHTTP(nil, req)
}

func TestTestCoreRequestContextMiddleware(t *testing.T) {
	logger, _, ctx := NewTestCoreRequestContext()

	coreCtx := getCoreContext(ctx)
	require.NotNil(t, coreCtx)
	require.Equal(t, logger, coreCtx.logger)
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
