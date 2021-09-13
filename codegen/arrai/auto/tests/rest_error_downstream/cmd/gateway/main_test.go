package main

import (
	"context"
	"net/http"
	"testing"

	"rest_error_downstream/internal/gen/pkg/servers/gateway"
	"rest_error_downstream/internal/gen/pkg/servers/gateway/backend"
)

func TestRestErrorDownstreamNew(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Backend.PostDoop.
		MockResponse(http.StatusTeapot, map[string]string{"Content-Type": `text/plain; charset=utf-8`}, backend.ErrorResponse{"teapots cannot doop"})

	gatewayTester.GetApiDoopList().
		ExpectResponseCode(200).
		ExpectResponseBody(gateway.GatewayResponse{"backend sent us an ErrorResponse: teapots cannot doop"}).
		Send()
}
