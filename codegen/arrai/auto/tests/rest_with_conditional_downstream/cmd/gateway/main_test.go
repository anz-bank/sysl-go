package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/core"
	"github.com/go-chi/chi"
	"github.com/sethvargo/go-retry"
	"github.com/spf13/afero"
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

func doGatewayRequestResponse(ctx context.Context, n int64) (string, error) {
	// Naive hand-written http client that attempts to call the Gateway service's fizzbuzz endpoint.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost:9021/fizzbuzz/%d", n), nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("got response with http status %d >= 400", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
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

	handlerFactory := func(tag string) func(w http.ResponseWriter, req *http.Request) {
		h := func(w http.ResponseWriter, req *http.Request) {
			complain := func(err error) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			n, err := strconv.Atoi(chi.URLParam(req, "n"))
			if err != nil {
				complain(err)
				return
			}
			var obj Payload
			obj.Content = fmt.Sprintf("%s(%d)", tag, n) // n.b. we depart from standard fizz-buzz spec
			responseData, err := json.Marshal(&obj)
			if err != nil {
				complain(err)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write(responseData)
		}
		return h
	}

	// define and start http server
	r := chi.NewRouter()
	r.Post("/Fizz/{n}", handlerFactory("FIZZ"))
	r.Post("/Buzz/{n}", handlerFactory("BUZZ"))
	r.Post("/FizzBuzz/{n}", handlerFactory("FIZZBUZZ"))

	server := &http.Server{Addr: addr, Handler: r}

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

func TestRestWithConditionalDownstreamAppSmokeTest(t *testing.T) {
	// Initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).Onto(context.Background())

	// Add in a fake filesystem to pass in config
	memFs := afero.NewMemMapFs()
	err := afero.Afero{Fs: memFs}.WriteFile("config.yaml", []byte(applicationConfig), 0777)
	require.NoError(t, err)
	ctx = core.ConfigFileSystemOnto(ctx, memFs)

	// FIXME patch core.Serve to allow it to optionally load app config path from ctx
	args := os.Args
	defer func() { os.Args = args }()
	os.Args = []string{"./gateway.out", "config.yaml"}

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
		_, err := doGatewayRequestResponse(ctx, 1)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Test if the endpoint of our gateway application server works
	expected := "FIZZ(3)\nBUZZ(5)\nFIZZ(6)\nFIZZ(9)\nBUZZ(10)\nFIZZ(12)\nFIZZBUZZ(15)\n"
	actual, err := doGatewayRequestResponse(ctx, 15)
	require.Nil(t, err)
	require.Equal(t, expected, actual)

	// FIXME how do we stop the application server?
}
