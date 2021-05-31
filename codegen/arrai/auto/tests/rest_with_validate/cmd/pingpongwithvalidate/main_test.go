package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"

	"github.com/anz-bank/sysl-go/core"
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

func doPongPongRequestResponse(ctx context.Context, identifier int64, value int64) (int, int, error) {
	url := fmt.Sprintf("http://localhost:9021/pong-pong")

	return doPingPongRequestResponseimpl(ctx, url, identifier, value)
}

func doPingPongRequestResponseimpl(ctx context.Context, url string, identifier int64, value int64) (int, int, error) {
	type payload struct {
		Identifier *int64 `json:"identifier,omitempty"`
		Value      *int64 `json:"value,omitempty"`
	}

	requestObj := payload{
		Identifier: &identifier,
		Value:      &value,
	}
	if identifier == -1 {
		requestObj.Identifier = nil
	}
	if value == -1 {
		requestObj.Value = nil
	}

	requestData, err := json.Marshal(&requestObj)
	if err != nil {
		return -1, -1, err
	}

	return doPingRequestResponseImpl(ctx, "POST", url, bytes.NewReader(requestData))
}

func doPingRequestResponseImpl(ctx context.Context, method string, url string, body io.Reader) (int, int, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
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
		_, _, err := doPongPongRequestResponse(ctx, 0, 0)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test various combinations of request data for the pong-pong endpoint.
	// The request is not validated for missing parameters

	// Test a successful request
	identifier, value, err := doPongPongRequestResponse(ctx, 1, 1)
	require.Nil(t, err)
	require.Equal(t, 1, identifier)
	require.Equal(t, 1, value)

	// Test a request that fails due to a missing request parameter (identifier)
	identifier, value, err = doPongPongRequestResponse(ctx, -1, 1)
	require.Equal(t, 400, err.(*ResponseError).StatusCode)

	// Test a request that fails due to a missing request parameter (value)
	identifier, value, err = doPongPongRequestResponse(ctx, 1, -1)
	require.Equal(t, 400, err.(*ResponseError).StatusCode)
}
