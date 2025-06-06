package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"rest_custom_middleware/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"
)

const serverPort = 9021 // no guarantee this port is free

const applicationConfig = `---
envPrefix: ASDF
genCode:
  upstream:
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021
  downstream:
    contextTimeout: "30s"
`

func doPingRequestResponse(ctx context.Context) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:%d/ping", serverPort), nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var obj struct {
		Data string `json:"data"`
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return "", err
	}
	return obj.Data, nil
}

func TestApplicationSmokeTest(t *testing.T) {

	// Override sysl-go app command line interface to directly pass in app config
	ctx := core.WithConfigFile(context.Background(), []byte(applicationConfig))

	appServer, err := newAppServer(ctx)
	require.NoError(t, err)
	defer func() {
		err := appServer.Stop()
		if err != nil {
			panic(err)
		}
	}()

	// Start application server
	go func() {
		err := appServer.Start()
		if err != nil {
			panic(err)
		}
	}()

	// Wait for application to come up
	backoff := retry.NewFibonacci(10 * time.Millisecond)
	backoff = retry.WithMaxDuration(10*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := doPingRequestResponse(ctx)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test to see if the ping endpoint of our pingpong application server works
	// and has picked up some information injected from our custom middleware
	expected := "once upon a time there was a rambutan"
	actual, err := doPingRequestResponse(ctx)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}

func TestRestCustomMiddleware_Success(t *testing.T) {
	t.Parallel()
	pingpongTester := pingpong.NewTestServer(t, context.Background(), createService, "")
	defer pingpongTester.Close()

	pingpongTester.GetPingList().
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(pingpong.Pong{Data: "once upon a time there was a rambutan"}).
		Send()
}

func TestRestCustomMiddleware_WriteErrorCallback(t *testing.T) {
	t.Parallel()
	pingpongTester := pingpong.NewTestServer(t, context.Background(), createService, "")
	defer pingpongTester.Close()

	pingpongTester.GetWriteerrorcallbackList().
		ExpectResponseCode(http.StatusTeapot).
		ExpectResponseBody(`{"err": "once upon a time there was a rambutan"}`).
		Send()
}

func TestRestCustomMiddleware_ErrorWriter(t *testing.T) {
	t.Parallel()
	pingpongTester := pingpong.NewTestServer(t, context.Background(), createService, "")
	defer pingpongTester.Close()

	pingpongTester.GetErrorwriterList().
		ExpectResponseCode(http.StatusPaymentRequired).
		ExpectResponseBody(`{"err": "once upon a time there was a rambutan"}`).
		Send()
}
