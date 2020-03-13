package internal

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestCheckForUnclosedResponses(t *testing.T) {
	logger, hook := test.NewNullLogger()
	// setup the monitor
	ctx := AddResponseBodyMonitorToContext(context.Background())

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
		CheckForUnclosedResponses(ctx, logger.WithField("test", "test"))
	})

	for _, entry := range hook.Entries {
		require.Equal(t, "\"internal.openBodyError{cause:\\\"/test\\\", err:\\\"response body not closed\\\"} \\n\"", entry.Message)
	}
}

func TestCheckForUnclosedResponses_AllClosed(t *testing.T) {
	logger, hook := test.NewNullLogger()
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
		CheckForUnclosedResponses(ctx, logger.WithField("test", "test"))
	})

	require.Empty(t, hook.Entries)
}
