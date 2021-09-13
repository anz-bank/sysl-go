package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"rest_env_config/internal/gen/pkg/servers/pingpong"
)

// BEWARE: the implementation of our config loading library
// (viper), completely ignores environment variables that you
// tell it to read UNLESS the config key is explicitly present
// in the config file AS WELL AS in an env var. That's why
// we need to set a dummy value of the port config key below.
// This seems fairly surprising, but it is the way it is.
// Ref: https://github.com/spf13/viper/issues/584
const applicationConfig = `---
envPrefix: ASDF
app:
    id2: 56789 #"this-should-be-replaced-by-env-var"
`

func TestRestEnvConfig(t *testing.T) {
	t.Parallel()

	const expected1 = 12345
	const expected2 = 9021

	_ = os.Setenv("ASDF_APP_ID2", fmt.Sprint(expected2))

	pingpongTester := pingpong.NewTestServer(t, context.Background(), createService, applicationConfig)
	defer pingpongTester.Close()

	pingpongTester.GetPing(expected1).
		ExpectResponseCode(http.StatusOK).
		ExpectResponseBody(pingpong.Pong{expected1, expected2}).
		Send()
}
