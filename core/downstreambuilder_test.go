package core

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripper struct {
}

func (roundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestDownstreamRoundTripper(t *testing.T) {
	client, err := BuildDownstreamHTTPClient(ctx,
		"Name",
		&Hooks{
			DownstreamRoundTripper: func(serviceName string, serviceURL string, original http.RoundTripper) http.RoundTripper {
				return roundTripper{}
			},
		},
		nil)

	require.Nil(t, err)
	require.NotNil(t, client)
	require.IsType(t, roundTripper{}, client.Transport)
}
