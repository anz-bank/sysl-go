package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
    bananastand:
      serviceURL: http://localhost:9022
      clientTimeout: 10s
`

const standardTestBanana = `TASTY-RIPE-BANANA`

type Payload struct {
	Content string `json:"content"`
}

func doGatewayRequestResponse(ctx context.Context, content string) (string, error) {
	// Naive hand-written http client that attempts to call the Gateway service's banana endpoint.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.
	client := &http.Client{}

	requestObj := Payload{
		Content: content,
	}
	requestData, err := json.Marshal(&requestObj)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:9021/banana", bytes.NewReader(requestData))
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

func startDummyBananaStandServer(addr string) (stopServer func() error) {
	// Starts a hand-written implementation of the BananaStand service running on given TCP Address.
	// Returns a function that can be used to stop the server.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.

	// define POST /banana endpoint handler
	h := func(w http.ResponseWriter, req *http.Request) {
		complain := func(err error) {
			// align with banana_stand.yaml error response schema
			type ErrorResponse struct {
				Details string `json:"details"`
			}
			response := &ErrorResponse{
				Details: err.Error(),
			}
			responseData, err := json.Marshal(response)
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(responseData)
		}
		err := req.ParseForm()
		if err != nil {
			complain(err)
			return
		}
		isAuthorized := req.Form.Get("client_id") == "joke_admin" && req.Form.Get("client_secret") == "changeit"
		if !isAuthorized {
			complain(errors.New("access denied"))
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"banana":"` + standardTestBanana + `"}`))
	}
	// define and start http server
	mux := http.NewServeMux()
	mux.HandleFunc("/banana", h)
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

func TestRestPostURLEncodedFormSmokeTest(t *testing.T) {
	// Override sysl-go app command line interface to directly pass in app config
	ctx := core.WithConfigFile(context.Background(), []byte(applicationConfig))

	// Start the dummy banana stand backend service running
	stopBananaStandServer := startDummyBananaStandServer("localhost:9022")
	defer func() {
		err := stopBananaStandServer()
		require.NoError(t, err)
	}()

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
	backoff, err := retry.NewFibonacci(20 * time.Millisecond)
	require.Nil(t, err)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := doGatewayRequestResponse(ctx, "testing; one two, one two; is this thing on?")
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test if the endpoint of our gateway application server works
	expected := standardTestBanana
	actual, err := doGatewayRequestResponse(ctx, "joke_admin:changeit")
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}
