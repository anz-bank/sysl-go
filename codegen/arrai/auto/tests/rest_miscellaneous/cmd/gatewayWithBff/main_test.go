package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
    contextTimeout: "1s"
    http:
      %s
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "1s"
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

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:9021"+basePath+"/ping/binary", bytes.NewReader(requestData))
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
	backoff, err := retry.NewFibonacci(20 * time.Millisecond)
	require.Nil(t, err)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := doGatewayRequestResponse(ctx, basePath, "")
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test if the endpoint of our gateway application server works
	inputbytes := make([]byte, 256)
	for i := range inputbytes {
		inputbytes[i] = byte(i)
	}
	input := base64.StdEncoding.EncodeToString(inputbytes)
	expected := input
	actual, err := doGatewayRequestResponse(ctx, basePath, input)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}

func TestMiscellaniousSmokeTest(t *testing.T) {
	startAndTestServer(t, fmt.Sprintf(applicationConfig, ""), "/bff")
	startAndTestServer(t, fmt.Sprintf(applicationConfig, `basePath: ""`), "/bff")
	startAndTestServer(t, fmt.Sprintf(applicationConfig, `basePath: "/"`), "")
	startAndTestServer(t, fmt.Sprintf(applicationConfig, `basePath: "/foo"`), "/foo")
}
