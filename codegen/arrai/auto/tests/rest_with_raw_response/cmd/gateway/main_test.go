package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"
	"rest_with_raw_response/internal/gen/pkg/servers/gateway"
	"rest_with_raw_response/internal/gen/pkg/servers/gateway/encoder_backend"
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

func doReverseBytesRequestResponse(ctx context.Context, b []byte, count int) ([]byte, error) {
	// Naive hand-written http client that attempts to call the Gateway service's encode endpoint.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://localhost:9021/reverse-bytes-n?count=%d", count), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("got response with http status %d >= 400", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func doReverseStringRequestResponse(ctx context.Context, s string, count int) (string, error) {
	// Naive hand-written http client that attempts to call the Gateway service's encode endpoint.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://localhost:9021/reverse-string-n?count=%d", count), bytes.NewReader([]byte(s)))
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("got response with http status %d >= 400", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func startDummyEncoderBackendServer(addr string) (stopServer func() error) {
	// Starts a hand-written implementation of the EncoderBackend service running on given TCP Address.
	// Returns a function that can be used to stop the server.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.

	// define reverse bytes handler
	b := func(w http.ResponseWriter, req *http.Request) {
		complain := func(err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			complain(err)
			return
		}

		for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
			data[i], data[j] = data[j], data[i]
		}

		w.Header().Add("Content-Type", "application/octet-stream")
		_, _ = w.Write(data)
	}

	// define reverse string handler
	s := func(w http.ResponseWriter, req *http.Request) {
		complain := func(err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			complain(err)
			return
		}

		runes := []rune(string(data))
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}

		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(string(runes)))
	}

	// define and start http server
	mux := http.NewServeMux()
	mux.HandleFunc("/reverse-bytes", b)
	mux.HandleFunc("/reverse-string", s)
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

func TestSimpleRestWithDownstreamAppSmokeTest(t *testing.T) {
	// Override sysl-go app command line interface to directly pass in app config
	ctx := core.WithConfigFile(context.Background(), []byte(applicationConfig))

	// Start the dummy encoder backend service running
	stopEncoderBackendServer := startDummyEncoderBackendServer("localhost:9022")
	defer func() {
		err := stopEncoderBackendServer()
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
		_, err := doReverseStringRequestResponse(ctx, "hello world", 0)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test if a reversed byte request works
	actualBytes, err := doReverseBytesRequestResponse(ctx, []byte{65, 55, 67}, 1)
	require.Nil(t, err)
	require.Equal(t, []byte{67, 55, 65}, actualBytes)

	// Test if the reversed string request works
	actualString, err := doReverseStringRequestResponse(ctx, "hello world", 1)
	require.Nil(t, err)
	require.Equal(t, "dlrow olleh", actualString)
}

func TestRestWithRawResponse(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	sendByte := []byte{65, 55, 67}
	returnByte := []byte{67, 55, 65}

	sendString := "hello world"
	returnString := "dlrow olleh"

	gatewayTester.Mocks.Encoder_backend.PostReverseBytes.
		ExpectBody(sendByte).
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `application/octet-stream`}, returnByte)

	gatewayTester.Mocks.Encoder_backend.PostReverseString.
		ExpectBody(sendString).
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `text/html; charset=utf-8`}, returnString)

	gatewayTester.Mocks.Encoder_backend.PostPingStringAlias.
		ExpectBody(encoder_backend.PingStringRequest(sendString)).
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `text/html; charset=utf-8`}, encoder_backend.PingStringRequest(returnString))

	gatewayTester.Mocks.Encoder_backend.PostPingByteAlias.
		ExpectBody(encoder_backend.PingByteRequest(sendByte)).
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `application/octet-stream`}, encoder_backend.PingByteRequest(returnByte))

	gatewayTester.PostReverseBytesN(1).
		WithBody(sendByte).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(returnByte).
		Send()

	gatewayTester.PostReverseStringN(1).
		WithBody(sendString).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(returnString).
		Send()

	gatewayTester.PostPingStringAlias().
		WithBody(gateway.PingStringRequest(sendString)).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(gateway.PingStringResponse(returnString)).
		Send()

	gatewayTester.PostPingByteAlias().
		WithBody(gateway.PingByteRequest(sendByte)).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(gateway.PingByteResponse(returnByte)).
		Send()
}

func TestRestWithRawResponse_ActualDownstream(t *testing.T) {
	stopEncoderBackendServer := startDummyEncoderBackendServer("localhost:9022")
	defer func() {
		err := stopEncoderBackendServer()
		require.NoError(t, err)
	}()

	gatewayTester := gateway.NewTestServerWithActualDownstreams(t, context.Background(), createService, applicationConfig)
	defer gatewayTester.Close()

	sendByte := []byte{65, 55, 67}
	returnByte := []byte{67, 55, 65}

	sendString := "hello world"
	returnString := "dlrow olleh"

	gatewayTester.PostReverseBytesN(1).
		WithBody(sendByte).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(returnByte).
		Send()

	gatewayTester.PostReverseStringN(1).
		WithBody(sendString).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(returnString).
		Send()
}
