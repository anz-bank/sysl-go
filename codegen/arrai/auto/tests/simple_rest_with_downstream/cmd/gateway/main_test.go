package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/anz-bank/sysl-go/syslgo"
	"github.com/stretchr/testify/require"
	"simple_rest_with_downstream/internal/gen/pkg/servers/gateway"
	"simple_rest_with_downstream/internal/gen/pkg/servers/gateway/encoder_backend"
)

func TestSimpleRestWithDownstream_Success(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Encoder_backend.PostRot13.
		ExpectHeaders(map[string]string{}).
		ExpectBody(encoder_backend.EncodingRequest{Content: "hello world"}).
		MockResponse(200, map[string]string{"Content-Type": `application/json`}, encoder_backend.EncodingResponse{Content: "uryyb jbeyq"})

	gatewayTester.PostEncodeEncoder_id("rot13").
		WithHeaders(map[string]string{}).
		WithBody(gateway.GatewayRequest{Content: "hello world"}).
		ExpectResponseCode(200).
		ExpectResponseBody(gateway.GatewayResponse{Content: "uryyb jbeyq"}).
		Send()
}

func TestSimpleRestWithDownstream_Fail(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Encoder_backend.PostRot13.
		ExpectHeaders(map[string]string{}).
		ExpectBody(encoder_backend.EncodingRequest{Content: "notfound"}).
		MockResponsePlain(404, map[string]string{"Content-Type": `text/plain; charset=utf-8`}, []byte(`Not found`))

	gatewayTester.PostEncodeEncoder_id("rot13").
		WithHeaders(map[string]string{}).
		WithBody(gateway.GatewayRequest{Content: "notfound"}).
		ExpectResponseCode(503).
		ExpectResponseBody(`{"status":{"code":"1013","description":"Downstream system is unavailable"}}`).
		Send()
}

// This test is to test the testing framework capabilities.
func TestSimpleRestWithDownstream_Extras(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	called := 0
	test := func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		called++
		require.Equal(t, r.URL.String(), "/rot13")
	}

	gatewayTester.Mocks.Encoder_backend.PostRot13.
		Expect(test).
		ExpectBody(encoder_backend.EncodingRequest{Content: "timeout"}).
		ExpectHeadersExist([]string{"foo"}).
		ExpectHeadersDoNotExist([]string{"foo3"}).
		ExpectHeadersExistExactly([]string{"foo", "foo2", "Content-Length", "Accept-Encoding", "User-Agent"}).
		Timeout()

	gatewayTester.PostEncodeEncoder_id("rot13").
		WithHeaders(map[string]string{"foo": "bar", "foo2": "bar"}).
		WithBody(gateway.GatewayRequest{Content: "timeout"}).
		TestResponseCode(func(t syslgo.TestingT, actual int) {
			called++
			require.Equal(t, http.StatusGatewayTimeout, actual)
		}).
		TestResponseBody(func(t syslgo.TestingT, actual []byte) {
			called++
			require.Equal(t, `{"status":{"code":"1005","description":"Time out from down stream services"}}`, string(actual))
		}).
		Send()

	require.Equal(t, called, 3)
}
