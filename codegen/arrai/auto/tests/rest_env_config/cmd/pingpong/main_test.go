package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"
)

const serverPort = 9021 // no guarantee this port is free

// BEWARE: the implementation of our config loading library
// (viper), completely ignores environment variables that you
// tell it to read UNLESS the config key is explicitly present
// in the config file AS WELL AS in an env var. That's why
// we need to set a dummy value of the port config key below.
// This seems fairly surprising, but it is the way it is.
// Ref: https://github.com/spf13/viper/issues/584
const applicationConfig = `---
envPrefix: ASDF
genCode:
  upstream:
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: "this-should-be-replaced-by-env-var"
  downstream:
    contextTimeout: "30s"
`

func doPingRequestResponse(ctx context.Context, identifier int) (int, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:%d/ping/%d", serverPort, identifier), nil)
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

	// Set environment variable to configure what port the server should listen on
	os.Setenv("ASDF_GENCODE_UPSTREAM_HTTP_COMMON_PORT", "9021")

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

}
