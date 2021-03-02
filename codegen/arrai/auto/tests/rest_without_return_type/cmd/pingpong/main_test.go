package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"
)

const applicationConfig = `---
genCode:
  upstream:
    contextTimeout: "1s"
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "1s"
`

func doRequest(ctx context.Context, target string, identifier int) (int, []byte, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:9021/%s/%d", target, identifier), nil)
	if err != nil {
		return -1, nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return -1, nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, data, err
}

func doRequestResponse(ctx context.Context, target string, identifier int) (int, int, error) {
	statusCode, data, err := doRequest(ctx, target, identifier)
	if err != nil {
		return -1, -1, err
	}
	var obj struct {
		Identifier int `json:"identifier"`
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return -1, -1, err
	}
	return obj.Identifier, statusCode, nil
}

func doPingTimeoutRequestResponse(ctx context.Context, identifier int) (int, int, error) {
	return doRequestResponse(ctx, "pingTimeout", identifier)
}

func verifyPing(ctx context.Context, t *testing.T, target string, identifier int) {
	actual, status, err := doRequestResponse(ctx, target, identifier)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, identifier, actual)
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
	backoff, err := retry.NewFibonacci(10 * time.Millisecond)
	require.Nil(t, err)
	backoff = retry.WithMaxDuration(10*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, _, err := doRequestResponse(ctx, "ping1", 0)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test to see if the ping endpoint of our pingpong application server works
	verifyPing(ctx, t, "ping1", 12345)
	verifyPing(ctx, t, "ping2", 12345)

	_, status, err := doPingTimeoutRequestResponse(ctx, 12345)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, status)
}
