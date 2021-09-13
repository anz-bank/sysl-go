package main

import (
	"context"
	"net/http"
	"testing"

	"simple_rest/internal/gen/pkg/servers/pingpong"
)

const applicationConfig = `---
genCode:
  upstream:
    contextTimeout: "0.5s"
`

func TestSimpleRest(t *testing.T) {
	t.Parallel()
	gatewayTester := pingpong.NewTestServer(t, context.Background(), createService, applicationConfig)
	defer gatewayTester.Close()

	const expected = 12345
	gatewayTester.GetPing(expected).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(pingpong.Pong{expected}).
		Send()

	gatewayTester.GetPingtimeout(expected).
		ExpectResponseCode(http.StatusInternalServerError).
		Send()

	gatewayTester.GetGetoneof(1).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(pingpong.OneOfResponseOne{1}).
		Send()

	gatewayTester.GetGetoneof(2).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(pingpong.OneOfResponseTwo{"Two"}).
		Send()
}
