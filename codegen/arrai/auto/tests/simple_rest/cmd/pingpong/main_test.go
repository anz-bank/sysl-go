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

func doPingRequestResponse(ctx context.Context, identifier int) (int, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:9021/ping/%d", identifier), nil)
	if err != nil {
		return -1, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	var obj struct {
		Identifier int `json:"identifier"`
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return -1, err
	}
	return obj.Identifier, nil
}

func TestApplicationSmokeTest(t *testing.T) {

	// Initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).Onto(context.Background())

	// Override sysl-go app command line interface to directly pass in app config
	ctx = core.WithConfigFile(ctx, []byte(applicationConfig))

	// Start pingpong application running as server
	go application(ctx)

	// Wait for application to come up
	backoff, err := retry.NewFibonacci(10 * time.Millisecond)
	require.Nil(t, err)
	backoff = retry.WithMaxDuration(10*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := doPingRequestResponse(ctx, 0)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test to see if the ping endpoint of our pingpong application server works
	expected := 12345
	actual, err := doPingRequestResponse(ctx, 12345)
	require.Nil(t, err)
	require.Equal(t, expected, actual)

	// FIXME how do we stop the application server?

}
