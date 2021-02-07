package internal

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/anz-bank/sysl-go/testutil"
	"github.com/stretchr/testify/require"
)

func TestCheckForUnclosedResponses(t *testing.T) {
	testContext, hook := testutil.NewTestContextWithLoggerHook()
	// setup the monitor
	ctx := AddResponseBodyMonitorToContext(testContext)

	// create a response which doesnt get read
	body := strings.NewReader("test string")
	req, err := http.NewRequest("GET", "/test", body)
	require.NoError(t, err)
	resp := &http.Response{
		Request: req,
		Body:    ioutil.NopCloser(body),
	}
	AddResponseToMonitor(ctx, resp)

	// test
	require.Panics(t, func() {
		CheckForUnclosedResponses(ctx)
	})

	require.NotEmpty(t, hook.Entries)
}

func TestCheckForUnclosedResponses_AllClosed(t *testing.T) {
	// setup the monitor
	ctx := AddResponseBodyMonitorToContext(context.Background())

	// create a response which doesnt get read
	testData := "test string"
	body := strings.NewReader(testData)
	req, err := http.NewRequest("GET", "/test", body)
	require.NoError(t, err)
	resp := &http.Response{
		Request: req,
		Body:    ioutil.NopCloser(body),
	}
	AddResponseToMonitor(ctx, resp)

	// test
	dst := make([]byte, len(testData))
	_, err = resp.Body.Read(dst)
	require.NoError(t, err)
	require.Equal(t, string(dst), testData)
	resp.Body.Close()

	require.NotPanics(t, func() {
		CheckForUnclosedResponses(ctx)
	})
}
