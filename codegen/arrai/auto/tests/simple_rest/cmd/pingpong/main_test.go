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
		return -1, nil,err
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

func doPingRequestResponse(ctx context.Context, identifier int) (int, int, error) {
	return doRequestResponse(ctx, "ping", identifier)
}

func doPingTimeoutRequestResponse(ctx context.Context, identifier int) (int, int, error) {
	return doRequestResponse(ctx, "pingTimeout", identifier)
}

func doOneOfRequest(ctx context.Context, identifier int) (*int64, *string, int, error) {
	status, data, err := doRequest(ctx, "getoneof", identifier)
	if err != nil {
		return nil, nil, -1, err
	}
	var (
		i *int64
		s *string
	)
	if identifier == 1 {
		var obj struct {
			IdentifierInt int64 `json:"identifierInt"`
		}
		err = json.Unmarshal(data, &obj)
		i = &obj.IdentifierInt
	} else {
		var obj struct {
			IdentifierString string `json:"identifierString"`
		}
		err = json.Unmarshal(data, &obj)
		s = &obj.IdentifierString
	}
	if err != nil {
		return nil, nil, -1, err
	}
	return i, s, status, nil
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
		_, _, err := doPingRequestResponse(ctx, 0)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test to see if the ping endpoint of our pingpong application server works
	expected := 12345
	actual, status, err := doPingRequestResponse(ctx, 12345)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, expected, actual)

	actual, status, err = doPingTimeoutRequestResponse(ctx, 12345)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, status)

	// Test oneOf endpoint
	i, s, status, err := doOneOfRequest(ctx, 1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.NotNil(t, i)
	require.Equal(t, int64(1), *i)
	require.Nil(t, s)
	i, s, status, err = doOneOfRequest(ctx, 2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Nil(t, i)
	require.NotNil(t, s)
	require.Equal(t, "Two", *s)
}
