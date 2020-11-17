package main

import (
	"context"
	"encoding/json"
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
    backend:
      serviceURL: http://localhost:9022
      clientTimeout: 10s
`

type Payload struct {
	Content string `json:"content"`
}

func doGatewayRequestResponse(ctx context.Context) (string, error) {
	// Naive hand-written http client that attempts to call the Gateway service's endpoint.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:9021/api/doop", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 500 {
		return string(data), nil // hack
	}
	var obj Payload
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return "", err
	}
	return obj.Content, nil
}

func startDummyBackendServer(addr string) (stopServer func() error) {
	// Starts a hand-written implementation of the Backend service running on given TCP Address.
	// Returns a function that can be used to stop the server.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.

	// define /doop endpoint handler
	h := func(w http.ResponseWriter, req *http.Request) {
		complain := func(err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		type ErrorResponse struct {
			Error string `json:"Error"`
		}
		var obj ErrorResponse
		obj.Error = "teapots cannot doop"
		responseData, err := json.Marshal(&obj)
		if err != nil {
			complain(err)
			return
		}
		w.WriteHeader(http.StatusTeapot)
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(responseData)
	}
	// define and start http server
	mux := http.NewServeMux()
	mux.HandleFunc("/doop", h)
	server := &http.Server{Addr: addr, Handler: mux}

	c := make(chan error, 1)

	go func() {
		c <- server.ListenAndServe()
	}()

	stopServer = func() error {
		// If the server stopped with some error before the caller
		// tried to stop it, return that error instead.
		select {
		case err := <-c:
			return err
		default:
		}
		return server.Close()
	}
	return stopServer
}

func TestRestErrorDownstream(t *testing.T) {

	// Initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).Onto(context.Background())

	// Add in a fake filesystem to pass in config
	// Override sysl-go app command line interface to directly pass in app config
	ctx = core.WithConfigFile(ctx, []byte(applicationConfig))

	// Start the dummy backend service running
	stopBackendServer := startDummyBackendServer("localhost:9022")
	defer func() {
		err := stopBackendServer()
		require.NoError(t, err)
	}()

	// Start gateway application running as server
	go application(ctx)

	// Wait for application to come up
	backoff, err := retry.NewFibonacci(20 * time.Millisecond)
	require.Nil(t, err)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := doGatewayRequestResponse(ctx)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	expected := "backend sent us an ErrorResponse: teapots cannot doop"
	actual, err := doGatewayRequestResponse(ctx)
	require.Nil(t, err)
	require.Equal(t, expected, actual)

	// FIXME how do we stop the application server?
}
