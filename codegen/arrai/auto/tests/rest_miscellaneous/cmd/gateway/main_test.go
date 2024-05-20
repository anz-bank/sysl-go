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

	"github.com/anz-bank/sysl-go/validator"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/array_response_backend"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/encoder_backend"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/multi_contenttype_backend"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/oneof_backend"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/types"

	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
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
    encoder_backend:
      clientTimeout: 1s
    oneof_backend:
      clientTimeout: 1s
    multi_contenttype_backend:
      clientTimeout: 1s
    array_response_backend:
      clientTimeout: 1s
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
	backoff, err := retry.NewFibonacci(20 * time.Millisecond)
	require.Nil(t, err)
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

func TestMiscellaneousSmokeTest(t *testing.T) {
	startAndTestServer(t, fmt.Sprintf(applicationConfig, ""), "")
	startAndTestServer(t, fmt.Sprintf(applicationConfig, `basePath: ""`), "")
	startAndTestServer(t, fmt.Sprintf(applicationConfig, `basePath: "/"`), "")
	startAndTestServer(t, fmt.Sprintf(applicationConfig, `basePath: "/foo"`), "/foo")
}

func TestMiscellaneous(t *testing.T) {
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
			gatewayTester := gateway.NewTestServer(t, context.Background(), createService, fmt.Sprintf(applicationConfig, test.basePath))
			defer gatewayTester.Close()

			gatewayTester.PostPingBinary().
				WithBody(gateway.GatewayBinaryRequest{Content: inputBytes}).
				ExpectResponseCode(200).
				ExpectResponseBody(gateway.GatewayBinaryResponse{Content: inputBytes}).
				Send()
		})
	}
}

func TestMiscellaneous_DownstreamQuery(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	const expectId = 24
	const expectString = "Foo"

	gatewayTester.Mocks.Encoder_backend.GetPingList.
		ExpectQueryParams(map[string][]string{"id": {fmt.Sprint(expectId)}}).
		MockResponse(200, map[string]string{"Content-Type": `application/json`}, encoder_backend.Pong{Identifier: expectId})

	gatewayTester.Mocks.Encoder_backend.GetPingStringS.
		ExpectURLParamS(expectString).
		MockResponse(200, map[string]string{"Content-Type": `application/json`}, encoder_backend.PongString{S: expectString})

	gatewayTester.GetPingIdList(expectId).
		ExpectResponseCode(200).
		ExpectResponseBody(gateway.Pong{Identifier: expectId}).
		Send()

	gatewayTester.GetPingStringS(expectString).
		ExpectResponseCode(200).
		ExpectResponseBody(gateway.PongString{S: expectString}).
		Send()

	gatewayTester.Mocks.Encoder_backend.GetPingList.
		ExpectQueryParams(map[string][]string{"id": {fmt.Sprint(expectId)}}).
		MockResponse(200, map[string]string{"Content-Type": `application/json`}, encoder_backend.Pong{Identifier: expectId})

	gatewayTester.Mocks.Multi_contenttype_backend.PostPingMultiColon.
		MockResponse(200, nil, nil)

	gatewayTester.GetPingAsyncdownstreamsList(expectId).
		ExpectResponseCode(200).
		ExpectResponseBody(gateway.Pong{Identifier: expectId}).
		Send()
}

func TestMiscellaneous_DownstreamMultipleContentType(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	const expectString = "Foo"

	gatewayTester.Mocks.Multi_contenttype_backend.PostPingMultiColon.
		ExpectBodyPlain(([]byte)(`{"val":"`+expectString+`"}`)).
		MockResponse(200, map[string]string{"Content-Type": `application/json`}, multi_contenttype_backend.PingMultiRes{Val: common.NewString(expectString)})

	gatewayTester.Mocks.Multi_contenttype_backend.PostPingMultiColon.
		ExpectBodyPlain(([]byte)(`{"val":"`+expectString+`"}`)).
		MockResponse(200, map[string]string{"Content-Type": `application/json; charset = utf-8`}, multi_contenttype_backend.PingMultiRes{Val: common.NewString(expectString)})

	gatewayTester.Mocks.Multi_contenttype_backend.PostPingMultiUrlencoded.
		ExpectBodyPlain(([]byte)(`val=`+expectString)).
		MockResponse(200, map[string]string{"Content-Type": `application/json`}, multi_contenttype_backend.PingMultiRes{Val: common.NewString(expectString)})

	gatewayTester.Mocks.Multi_contenttype_backend.PostPingMultiUrlencoded.
		ExpectBodyPlain(([]byte)(`val=`+expectString)).
		MockResponse(200, map[string]string{"Content-Type": `application/json; charset = utf-8`}, multi_contenttype_backend.PingMultiRes{Val: common.NewString(expectString)})

	gatewayTester.GetPingMultiContentBackendS(expectString).
		ExpectResponseCode(200).
		ExpectResponseBody(gateway.PongString{S: expectString}).
		Send()
}

func TestMiscellaneous_Patch(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	const expectString = "Foo"

	gatewayTester.PatchPing().
		WithBody(gateway.GatewayPatchRequest{Content: expectString}).
		ExpectResponseCode(202).
		ExpectResponseHeaders(map[string]string{"Content-Type": `application/json;charset=UTF-8`}).
		ExpectResponseBody(gateway.GatewayPatchResponse{Content: expectString}).
		Send()
}

func TestMiscellaneous_OneOf(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Oneof_backend.PostRotateOneOf.
		ExpectBody(oneof_backend.OneOfRequest{Values: []oneof_backend.OneOfRequest_values{
			{One: &oneof_backend.One{One: true}},
			{Two: &oneof_backend.Two{Two: "two"}},
			{Three: &oneof_backend.Three{Three: 3}},
			{EmptyType: &oneof_backend.EmptyType{}},
		}}).
		MockResponse(200, nil, oneof_backend.OneOfResponse{Values: []oneof_backend.OneOfResponse_values{
			{Three: &oneof_backend.Three{Three: 3}},
			{One: &oneof_backend.One{One: true}},
			{Two: &oneof_backend.Two{Two: "two"}},
			{EmptyType: &oneof_backend.EmptyType{}},
		}})

	gatewayTester.PostRotateOneOf().
		WithBody(gateway.OneOfRequest{Values: []gateway.OneOfRequest_values{
			{One: &gateway.One{One: true}},
			{Two: &gateway.Two{Two: "two"}},
			{Three: &gateway.Three{Three: 3}},
			{EmptyType: &gateway.EmptyType{}},
		}}).
		ExpectResponseCode(201).
		ExpectResponseHeaders(map[string]string{"Content-Type": `application/json; charset = utf-8`}).
		ExpectResponseBody(gateway.OneOfResponse{Values: []gateway.OneOfResponse_values{
			{Three: &gateway.Three{Three: 3}},
			{One: &gateway.One{One: true}},
			{Two: &gateway.Two{Two: "two"}},
			{EmptyType: &gateway.EmptyType{}},
		}}).
		Send()
}

func TestMiscellaneous_OneOfRaw(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	req := ([]byte)(`{"values":[{"one":true},{"two":"two"},{"three":3}]}`)
	res := ([]byte)(`{"values":[{"three":3},{"one":true},{"two":"two"}]}`)

	gatewayTester.Mocks.Oneof_backend.PostRotateOneOf.
		ExpectBodyPlain(req).
		MockResponse(200, nil, res)

	gatewayTester.PostRotateOneOf().
		WithBodyPlain(req).
		ExpectResponseCode(201).
		ExpectResponseBody(res).
		Send()
}

func TestMiscellaneous_OneOfErrorReturn(t *testing.T) {
	t.Parallel()

	// Just want to confirm that it generates types that satisfy the 'error' and 'Validator' interfaces
	var e error
	e = gateway.Post_rotateOneOf_400_resp_type_body{}
	e = oneof_backend.Post_rotateOneOf_400_resp_type_body{}
	_ = e

	var v validator.Validator
	v = &gateway.Post_rotateOneOf_400_resp_type_body{}
	v = &oneof_backend.Post_rotateOneOf_400_resp_type_body{}
	_ = v
}

func TestMiscellaneous_MultiCode(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.GetPingMultiCode(0).
		ExpectResponseCode(200).
		ExpectResponseHeaders(map[string]string{"Content-Type": `application/json;charset=UTF-8`}).
		ExpectResponseBody(gateway.Pong{Identifier: 0}).
		Send()

	gatewayTester.GetPingMultiCode(1).
		ExpectResponseCode(201).
		ExpectResponseHeaders(map[string]string{"Content-Type": `application/json`}).
		ExpectResponseBody(gateway.PongString{S: "One"}).
		Send()
}

func TestMiscellaneous_CheckExternals(t *testing.T) {
	t.Parallel()
	// Just want to confirm that it generates the type with the correct name
	var _ *gateway.EXTERNAL_MissingType = gateway.UndefinedPropertyType{}.Value
}

func TestMiscellaneous_DoubleUnderscore(t *testing.T) {
	t.Parallel()
	// Just want to confirm that it generates a type that is accessible
	_ = encoder_backend.Double_underscore{S: "accessible"}
}

func TestMiscellaneous_StartWithInteger(t *testing.T) {
	t.Parallel()
	// Just want to confirm that it generates a valid type
	_ = encoder_backend.N1TypeStartsWithInt{N1s: "accessible"}
}

func TestMiscellaneous_TypesSomethingExternal(t *testing.T) {
	t.Parallel()

	_ = types.SomethingExternal{}
}

func TestMiscellaneous_EmptyResponse(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.GetEmptyResponseList().
		ExpectResponseCode(200).
		ExpectResponseHeaders(map[string]string{"Content-Type": `application/json`}).
		ExpectResponseBody("{}").
		Send()

	gatewayTester.GetWithHeaderList().
		ExpectResponseCode(200).
		ExpectResponseHeaders(map[string]string{"Content-Type": `application/json`}).
		ExpectResponseBody("{}").
		Send()
}

func TestMiscellaneous_ArrayResponse(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Array_response_backend.GetArrayResponseList.
		MockResponse(200, map[string]string{"Content-Type": `application/json`}, []*array_response_backend.Res{{Val: common.NewString("Res")}})

	gatewayTester.GetPingArrayResponseList().
		ExpectResponseCode(200).
		ExpectResponseHeaders(map[string]string{"Content-Type": `application/json`}).
		ExpectResponseBody([]gateway.Res{{Val: common.NewString("Res")}}).
		Send()

	gatewayTester.Mocks.Array_response_backend.GetStringResponseList.
		MockResponse(200, map[string]string{"Content-Type": `text/plain`}, "Res")

	gatewayTester.GetPingStringResponseList().
		ExpectResponseCode(200).
		ExpectResponseHeaders(map[string]string{"Content-Type": `text/plain`}).
		ExpectResponseBody("Res").
		Send()

	gatewayTester.Mocks.Array_response_backend.GetBytesResponseList.
		MockResponse(200, map[string]string{"Content-Type": `application/octet-stream`}, []byte("Res"))

	gatewayTester.GetPingBytesResponseList().
		ExpectResponseCode(200).
		ExpectResponseHeaders(map[string]string{"Content-Type": `application/octet-stream`}).
		ExpectResponseBody([]byte("Res")).
		Send()
}
