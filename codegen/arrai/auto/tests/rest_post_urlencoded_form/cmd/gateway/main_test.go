package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/syslgo"
	"github.com/stretchr/testify/require"
	"rest_post_urlencoded_form/internal/gen/pkg/servers/gateway"
	"rest_post_urlencoded_form/internal/gen/pkg/servers/gateway/bananastand"
)

const standardTestBanana = `TASTY-RIPE-BANANA`

func TestRestPostURLEncodedForm(t *testing.T) {
	t.Parallel()
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Bananastand.PostBanana.
		Expect(func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			require.NoError(t, err)
			isAuthorized := r.Form.Get("client_id") == "joke_admin" && r.Form.Get("client_secret") == "changeit"
			require.True(t, isAuthorized)
		}).
		MockResponse(
			http.StatusOK,
			map[string]string{"Content-Type": `application/json`},
			bananastand.BananaResponse{common.NewString(standardTestBanana)},
		)

	gatewayTester.PostBanana().
		WithBody(gateway.GatewayRequest{"joke_admin:changeit"}).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(gateway.GatewayResponse{standardTestBanana}).
		Send()
}
