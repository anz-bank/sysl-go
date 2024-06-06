package main

import (
	"context"
	"net/http"
	"testing"

	"rest_without_return_type/internal/gen/pkg/servers/pingpong"
)

const applicationConfig = `---
genCode:
  upstream:
    contextTimeout: "0.5s"
`

func TestRestWithoutReturnType(t *testing.T) {
	t.Parallel()
	pingpongTester := pingpong.NewTestServer(t, context.Background(), createService, applicationConfig)
	defer pingpongTester.Close()

	pingpongTester.GetPing0(12345).
		ExpectResponseCode(http.StatusOK).
		Send()

	pingpongTester.GetPing1(12345).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(pingpong.Pong{12345}).
		Send()

	pingpongTester.GetPing2(12345).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(pingpong.Pong{12345}).
		Send()

	pingpongTester.GetPing3(12345).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(pingpong.Pong{12345}).
		Send()

	pingpongTester.GetPing4(12345).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(pingpong.Pong{12345}).
		Send()

	pingpongTester.GetPing5(0).
		ExpectResponseCode(http.StatusCreated).
		ExpectResponseBody(pingpong.Pong{0}).
		Send()

	pingpongTester.GetPing5(12345).
		ExpectResponseCode(http.StatusAccepted).
		ExpectResponseBody(pingpong.Pong2{12345}).
		Send()

	pingpongTester.GetPing6(0).
		ExpectResponseCode(http.StatusCreated).
		ExpectResponseBody(pingpong.Pong{0}).
		Send()

	pingpongTester.GetPing6(12345).
		ExpectResponseCode(http.StatusAccepted).
		ExpectResponseBody(pingpong.Pong2{12345}).
		Send()

	pingpongTester.GetPingtimeout(12345).
		ExpectResponseCode(http.StatusInternalServerError).
		Send()
}
