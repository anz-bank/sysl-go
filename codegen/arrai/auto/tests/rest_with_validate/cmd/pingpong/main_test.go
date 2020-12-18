package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"
)

const applicationConfig = `---
genCode:
  upstream:
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`

type ResponseError struct {
	StatusCode int
	Body       []byte
}

func (r *ResponseError) Error() string {
	return fmt.Sprintf("code: %d, body: %s", r.StatusCode, r.Body)
}

func doPingRequestResponse(ctx context.Context, identifier int, value int) (int, int, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:9021/ping/%d/%d", identifier, value), nil)
	if err != nil {
		return -1, -1, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return -1, -1, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, -1, err
	}
	if resp.StatusCode != 200 {
		return -1, -1, &ResponseError{resp.StatusCode, data}
	}
	var obj struct {
		Identifier int `json:"identifier"`
		Value      int `json:"value"`
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return -1, -1, err
	}
	return obj.Identifier, obj.Value, nil
}

func doPingIgnoreRequestResponse(ctx context.Context, identifier int, value int) (int, int, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:9021/ping-ignore/%d/%d", identifier, value), nil)
	if err != nil {
		return -1, -1, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return -1, -1, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, -1, err
	}
	if resp.StatusCode != 200 {
		return -1, -1, &ResponseError{resp.StatusCode, data}
	}
	var obj struct {
		Identifier int `json:"identifier"`
		Value      int `json:"value"`
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return -1, -1, err
	}
	return obj.Identifier, obj.Value, nil
}

func TestApplicationSmokeTest(t *testing.T) {

	// Initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).Onto(context.Background())

	// Override sysl-go app command line interface to directly pass in app config
	ctx = core.WithConfigFile(ctx, []byte(applicationConfig))

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
		_, _, err := doPingRequestResponse(ctx, 0, 0)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test various combinations of request data for the ping endpoint.
	// The request/response parameters have the following validation constraints:
	// Identity:    request: min=0,max=20       response: max=10
	// Value:       request: min=0              response: max=10

	// Test a successful request
	identifier, value, err := doPingRequestResponse(ctx, 0, 0)
	require.Nil(t, err)
	require.Equal(t, 0, identifier)
	require.Equal(t, 0, value)

	// Test a request that fails due to an invalid request parameter (identity)
	identifier, value, err = doPingRequestResponse(ctx, -1, 0)
	require.Equal(t, 400, err.(*ResponseError).StatusCode)

	// Test a request that fails due to an invalid request parameter (value)
	identifier, value, err = doPingRequestResponse(ctx, 0, -1)
	require.Equal(t, 400, err.(*ResponseError).StatusCode)

	// Test a request that fails due to an invalid response parameter (identity)
	identifier, value, err = doPingRequestResponse(ctx, 20, 0)
	require.Equal(t, 500, err.(*ResponseError).StatusCode)

	// Test a request that fails due to an invalid response parameter (value)
	identifier, value, err = doPingRequestResponse(ctx, 0, 20)
	require.Equal(t, 500, err.(*ResponseError).StatusCode)

	// Test various combinations of request data for the ping-ignore endpoint.
	// The request/response parameters have the following validation constraints:
	// Identity:    request: min=0,max=20       response: max=10
	// Value:       request: oneof=0 1 20       response: max=10

	// Test a successful request
	identifier, value, err = doPingIgnoreRequestResponse(ctx, 0, 0)
	require.Nil(t, err)
	require.Equal(t, 0, identifier)
	require.Equal(t, 0, value)

	// Test a request that fails due to an invalid request parameter (identity)
	identifier, value, err = doPingIgnoreRequestResponse(ctx, -1, 0)
	require.Equal(t, 400, err.(*ResponseError).StatusCode)

	// Test a request that fails due to an invalid request parameter (value)
	identifier, value, err = doPingIgnoreRequestResponse(ctx, 0, -1)
	require.Equal(t, 400, err.(*ResponseError).StatusCode)

	// Test a request is successful because an invalid response parameter is ignored (identity)
	identifier, value, err = doPingIgnoreRequestResponse(ctx, 20, 0)
	require.Nil(t, err)
	require.Equal(t, 20, identifier)
	require.Equal(t, 0, value)

	// Test a request is successful because an invalid response parameter is ignored (value)
	identifier, value, err = doPingIgnoreRequestResponse(ctx, 0, 20)
	require.Nil(t, err)
	require.Equal(t, 0, identifier)
	require.Equal(t, 20, value)
}
