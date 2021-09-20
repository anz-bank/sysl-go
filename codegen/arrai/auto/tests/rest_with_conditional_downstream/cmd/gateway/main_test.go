package main

import (
	"context"
	"net/http"
	"testing"

	"rest_with_conditional_downstream/internal/gen/pkg/servers/gateway"
	"rest_with_conditional_downstream/internal/gen/pkg/servers/gateway/backend"
)

func TestRestWithConditionalDownstream(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Backend.PostFizzWithArg.
		ExpectURLParamN(3).
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `application/json`}, backend.Response{"FIZZ(3)"})
	gatewayTester.Mocks.Backend.PostFizzWithArg.
		ExpectURLParamN(6).
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `application/json`}, backend.Response{"FIZZ(6)"})
	gatewayTester.Mocks.Backend.PostFizzWithArg.
		ExpectURLParamN(9).
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `application/json`}, backend.Response{"FIZZ(9)"})
	gatewayTester.Mocks.Backend.PostFizzWithArg.
		ExpectURLParamN(12).
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `application/json`}, backend.Response{"FIZZ(12)"})

	gatewayTester.Mocks.Backend.PostBuzzWithArg.
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `application/json`}, backend.Response{"BUZZ(5)"})
	gatewayTester.Mocks.Backend.PostBuzzWithArg.
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `application/json`}, backend.Response{"BUZZ(10)"})

	gatewayTester.Mocks.Backend.PostFizzbuzzWithArg.
		MockResponse(http.StatusOK, map[string]string{"Content-Type": `application/json`}, backend.Response{"FIZZBUZZ(15)"})

	gatewayTester.GetFizzbuzz(15).
		WithHeaders(map[string]string{}).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(gateway.GatewayResponse{Content: "FIZZ(3)\nBUZZ(5)\nFIZZ(6)\nFIZZ(9)\nBUZZ(10)\nFIZZ(12)\nFIZZBUZZ(15)\n"}).
		Send()
}
