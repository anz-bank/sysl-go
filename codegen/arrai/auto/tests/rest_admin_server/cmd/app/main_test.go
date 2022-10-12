package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"
)

const adminServerPort = 9033 // no guarantee this port is free

const applicationConfig = `---
admin:
  contextTimeout: 10s
  http:
    basePath: /admin
    readTimeout: 10s
    writeTimeout: 10s
    common:
      hostName: "localhost"
      port: 9033
genCode:
  upstream:
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: "9021"
  downstream:
    contextTimeout: "30s"
`

func doHTTPGet(ctx context.Context, endpoint string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:%d/%s", adminServerPort, endpoint), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got HTTP status code %d", resp.StatusCode)
	}
	return data, nil
}

func TestAppAdminServerSmokeTest(t *testing.T) {
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
	backoff, err := retry.NewFibonacci(10 * time.Millisecond)
	require.Nil(t, err)
	backoff = retry.WithMaxDuration(10*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err = doHTTPGet(ctx, "admin/-/status")
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test to see if the admin endpoints of our application server work
	// and contain some suggestive test fragments
	statusData, err := doHTTPGet(ctx, "admin/-/status")
	require.Nil(t, err)
	require.Contains(t, string(statusData), "status")

	metricsData, err := doHTTPGet(ctx, "admin/-/metrics")
	require.Nil(t, err)
	require.Contains(t, string(metricsData), "go_gc_duration_seconds")
}
