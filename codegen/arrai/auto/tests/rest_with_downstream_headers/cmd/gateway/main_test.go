package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/anz-bank/sysl-go/common"
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
    encoder_backend:
      serviceURL: http://localhost:9022
      clientTimeout: 10s
`

type Payload struct {
	Content string `json:"content"`
}

func doGatewayRequestResponse(ctx context.Context, content string) (string, error) {
	return doGatewayRequestResponse_impl(ctx, content, false)
}

func doGatewayRequestResponseMissingHeader(ctx context.Context, content string) (string, error) {
	return doGatewayRequestResponse_impl(ctx, content, true)
}

func doGatewayRequestResponse_impl(ctx context.Context, content string, missingRequiredHeader bool) (string, error) {
	// Naive hand-written http client that attempts to call the Gateway service's encode endpoint.
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

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:9021/encode/rot13", bytes.NewReader(requestData))
	if err != nil {
		return "", err
	}
	if !missingRequiredHeader {
		req.Header.Add("y", `blahblah`)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("got response with http status %d >= 400", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
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

func startDummyEncoderBackendServer(addr string) (stopServer func() error) {
	// Starts a hand-written implementation of the EncoderBackend service running on given TCP Address.
	// Returns a function that can be used to stop the server.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.

	// valuable business logic as used in our dummy implementation of EncoderBackend service
	toRot13 := make(map[rune]rune)
	az := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
	for i, r := range az {
		toRot13[r] = az[(i+13)%len(az)]
	}
	rot13 := func(s string) string {
		var b strings.Builder
		for _, r := range s {
			s, ok := toRot13[unicode.ToLower(r)]
			if ok {
				b.WriteRune(s)
			} else {
				b.WriteRune(r)
			}
		}
		return b.String()
	}

	// define /rot13 endpoint handler
	h := func(w http.ResponseWriter, req *http.Request) {
		complain := func(err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		data, err := io.ReadAll(req.Body)
		if err != nil {
			complain(err)
			return
		}
		var obj Payload
		err = json.Unmarshal(data, &obj)
		if err != nil {
			complain(err)
			return
		}
		obj.Content = rot13(obj.Content)
		responseData, err := json.Marshal(&obj)
		if err != nil {
			complain(err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(responseData)
	}
	// define and start http server
	mux := http.NewServeMux()
	mux.HandleFunc("/rot13", h)
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

func confirmErrorType(confirm chan bool) func(ctx context.Context, err error) *common.HTTPError {
	return func(ctx context.Context, err error) *common.HTTPError {
		var zeroHeaderLengthError *common.ZeroHeaderLengthError
		if errors.As(err, &zeroHeaderLengthError) && zeroHeaderLengthError.CausedByParam("y") {
			confirm <- true
		} else {
			confirm <- false
		}

		return nil
	}
}

func TestRestWithDownstreamHeadersAppSmokeTest(t *testing.T) {
	// Override sysl-go app command line interface to directly pass in app config
	ctx := core.WithConfigFile(context.Background(), []byte(applicationConfig))

	// Start the dummy encoder backend service running
	stopEncoderBackendServer := startDummyEncoderBackendServer("localhost:9022")
	defer func() {
		err := stopEncoderBackendServer()
		require.NoError(t, err)
	}()

	confirmChan := make(chan bool, 1)

	appServer, err := newAppServerWithErrorMapper(ctx, confirmErrorType(confirmChan))
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
	backoff := retry.NewFibonacci(20 * time.Millisecond)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := doGatewayRequestResponse(ctx, "testing; one two, one two; is this thing on?")
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Test if the endpoint of our gateway application server works
	expected := "uryyb jbeyq"
	actual, err := doGatewayRequestResponse(ctx, "hello world")
	require.Nil(t, err)
	require.Equal(t, expected, actual)

	_, err = doGatewayRequestResponseMissingHeader(ctx, "hello world")
	require.Error(t, err)
	require.True(t, <-confirmChan)
}
