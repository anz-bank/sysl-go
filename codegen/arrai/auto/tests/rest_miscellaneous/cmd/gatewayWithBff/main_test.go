package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"rest_miscellaneous/internal/gen/pkg/servers/gatewayWithBff"
	"rest_miscellaneous/internal/gen/pkg/servers/gatewayWithBff/gateway"
	"rest_miscellaneous/internal/gen/pkg/servers/gatewayWithBff/types"

	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"
)

const applicationConfig = `---
genCode:
  upstream:
    contextTimeout: "1s"
    http:
      %s
      common:
        hostName: "localhost"
        port: 9022 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "1s"
    gateway:
        clientTimeout: "1s"
`

type Payload struct {
	Content string `json:"content"`
}

func doGatewayRequestResponse(ctx context.Context, basePath, content string) (string, error) {
	// Naive hand-written http client that attempts to call the Gateway service's /ping/binary endpoint.
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

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:9022"+basePath+"/ping/binary", bytes.NewReader(requestData))
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

func startAndTestServer(t *testing.T, applicationConfig, basePath string) {
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
	backoff := retry.NewFibonacci(20 * time.Millisecond)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	require.NoError(t, retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := doGatewayRequestResponse(ctx, basePath, "")
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	}))

	// Test if the endpoint of our gateway application server works
	inputbytes := make([]byte, 256)
	for i := range inputbytes {
		inputbytes[i] = byte(i)
	}
	input := base64.StdEncoding.EncodeToString(inputbytes)
	expected := input
	actual, err := doGatewayRequestResponse(ctx, basePath, input)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestMiscellaniousSmokeTest(t *testing.T) {
	startAndTestServer(t, fmt.Sprintf(applicationConfig, ""), "/bff")
	startAndTestServer(t, fmt.Sprintf(applicationConfig, `basePath: ""`), "/bff")
	startAndTestServer(t, fmt.Sprintf(applicationConfig, `basePath: "/"`), "")
	startAndTestServer(t, fmt.Sprintf(applicationConfig, `basePath: "/foo"`), "/foo")
}

func TestMiscellaniousWithBff(t *testing.T) {
	inputBytes := make([]byte, 256)
	for i := range inputBytes {
		inputBytes[i] = byte(i)
	}

	for _, test := range []struct {
		name, basePath string
	}{
		{`missing`, ``},
		{`empty`, `basePath: ""`},
		{`slash`, `basePath: "/"`},
		{`foo`, `basePath: "/foo"`},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			gatewayTester := gatewayWithBff.NewTestServer(t, context.Background(), createService, fmt.Sprintf(applicationConfig, test.basePath))
			defer gatewayTester.Close()

			gatewayTester.PostPingBinary().
				WithBody(gatewayWithBff.GatewayBinaryRequest{Content: inputBytes}).
				ExpectResponseCode(200).
				ExpectResponseBody(gatewayWithBff.GatewayBinaryResponse{Content: inputBytes}).
				Send()
		})
	}
}

func TestMultiResponse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		code         int
		mockResp     any
		expectedResp string
	}{
		{200, gateway.Pong{Identifier: 123}, `Pong response: &{Identifier:123}`},
		{201, gateway.PongString{S: `pong string`}, `PongString response: pong string`},
		{202, gateway.PongString{S: `pong string 202`}, `PongString response: pong string 202`},
	} {
		test := test
		t.Run(fmt.Sprintf("code-%d", test.code), func(t *testing.T) {
			t.Parallel()
			gatewayTester := gatewayWithBff.NewTestServer(t, context.Background(), createService, fmt.Sprintf(applicationConfig, ``))
			defer gatewayTester.Close()

			gatewayTester.Mocks.Gateway.GetPingMultiCode.
				ExpectURLParamCode(int64(test.code)).
				MockResponse(test.code, map[string]string{}, test.mockResp)

			gatewayTester.PostMultiResponses().
				WithBody(gatewayWithBff.Code{int64(test.code)}).
				ExpectResponseBody(&types.SomethingExternal{
					Data: test.expectedResp,
				}).
				Send()
		})
	}
}

func TestMultiStatusCodeResponse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		code         int
		mockResp     any
		expectedResp string
	}{
		{200, gateway.Pong{Identifier: 200}, `Pong response: &{Identifier:200}`},
		{201, gateway.Pong{Identifier: 201}, `Pong response: &{Identifier:201}`},
		{202, gateway.PongString{S: `pong string 202`}, `PongString response: pong string 202`},
	} {
		test := test
		t.Run(fmt.Sprintf("code-%d", test.code), func(t *testing.T) {
			t.Parallel()
			gatewayTester := gatewayWithBff.NewTestServer(t, context.Background(), createService, fmt.Sprintf(applicationConfig, ``))
			defer gatewayTester.Close()

			gatewayTester.Mocks.Gateway.GetPingMultiCodeTypesList.
				ExpectURLParamCode(int64(test.code)).
				MockResponse(test.code, map[string]string{}, test.mockResp)

			gatewayTester.PostMultiStatuses().
				WithBody(gatewayWithBff.Code{int64(test.code)}).
				ExpectResponseBody(&types.SomethingExternal{
					Data: test.expectedResp,
				}).
				Send()
		})
	}
}

func TestMultiResponseError(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		code      int
		mockError any
	}{
		{400, gateway.GatewayBinaryRequest{Content: []byte("400 error")}},
		{500, gateway.GatewayBinaryResponse{Content: []byte("500 error")}},
	} {
		test := test
		t.Run(fmt.Sprintf("code-%d", test.code), func(t *testing.T) {
			t.Parallel()
			gatewayTester := gatewayWithBff.NewTestServer(t, context.Background(), createService, fmt.Sprintf(applicationConfig, ``))
			defer gatewayTester.Close()

			gatewayTester.Mocks.Gateway.GetPingMultiCode.
				ExpectURLParamCode(int64(test.code)).
				MockResponse(test.code, map[string]string{}, test.mockError)

			gatewayTester.PostMultiResponses().
				WithBody(gatewayWithBff.Code{int64(test.code)}).
				ExpectResponseBody(test.mockError).
				Send()
		})
	}
}

func TestMultiStatusCodeError(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		code      int
		mockError any
	}{
		{400, gateway.GatewayBinaryRequest{Content: []byte("400 error")}},
		{500, gateway.GatewayBinaryRequest{Content: []byte("500 error")}},
	} {
		test := test
		t.Run(fmt.Sprintf("code-%d", test.code), func(t *testing.T) {
			t.Parallel()
			gatewayTester := gatewayWithBff.NewTestServer(t, context.Background(), createService, fmt.Sprintf(applicationConfig, ``))
			defer gatewayTester.Close()

			gatewayTester.Mocks.Gateway.GetPingMultiCodeTypesList.
				ExpectURLParamCode(int64(test.code)).
				MockResponse(test.code, map[string]string{}, test.mockError)

			gatewayTester.PostMultiStatuses().
				WithBody(gatewayWithBff.Code{int64(test.code)}).
				ExpectResponseBody(test.mockError).
				Send()
		})
	}
}
