package jwtauth

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/jsontime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthFromConfig(t *testing.T) {
	ctx := testContext()
	url, client := testClient()
	ac := &Config{
		Issuers: []IssuerConfig{
			{
				Name:     "test",
				JWKSURL:  url,
				CacheTTL: jsontime.Duration(time.Minute),
			},
		},
	}
	auth, err := AuthFromConfig(ctx, ac, func(string) *http.Client { return client })
	assert.NoError(t, err)
	assert.NotNil(t, auth)
}

func TestAuthFromConfigDuplicateIssuerNames(t *testing.T) {
	ctx := testContext()
	url, client := testClient()
	ac := &Config{
		Issuers: []IssuerConfig{
			{
				Name:     "test",
				JWKSURL:  url,
				CacheTTL: jsontime.Duration(time.Minute),
			},
			{
				Name:     "test",
				JWKSURL:  url,
				CacheTTL: jsontime.Duration(time.Minute),
			},
		},
	}
	_, err := AuthFromConfig(ctx, ac, func(string) *http.Client { return client })
	assert.Error(t, err)
}

func TestAuthFromConfigIssuerMissingName(t *testing.T) {
	ctx := testContext()
	url, client := testClient()
	ac := &Config{
		Issuers: []IssuerConfig{
			{
				JWKSURL: url,
			},
		},
	}
	_, err := AuthFromConfig(ctx, ac, func(string) *http.Client { return client })
	assert.Error(t, err)
}

func TestAuthFromConfigIssuerNoMethod(t *testing.T) {
	ctx := testContext()
	_, client := testClient()
	ac := &Config{
		Issuers: []IssuerConfig{
			{
				Name: "test",
			},
		},
	}
	_, err := AuthFromConfig(ctx, ac, func(string) *http.Client { return client })
	assert.Error(t, err)
}

func TestVerifierFromConfigRemoteJWKS(t *testing.T) {
	ctx := testContext()
	url, client := testClient()
	ic := IssuerConfig{Name: "test", JWKSURL: url, CacheTTL: jsontime.Duration(time.Minute)}
	_, err := VerifierFromIssuerConfig(ctx, ic, client)
	require.NoError(t, err)
}

// JSON

func TestConfigMarshal(t *testing.T) {
	c := &Config{
		Issuers: []IssuerConfig{},
	}
	expected := `{"issuers":[]}`
	actual, err := json.Marshal(c)
	require.NoError(t, err)
	assert.JSONEq(t, expected, string(actual))
}

func TestConfigUnmarshal(t *testing.T) {
	raw := `{"allowUnsafeJwks":true,"issuers":[{"name":"test","jwksUrl":"https://localhost:8080","cacheTTL":"1m"}]}`
	var c Config
	require.NoError(t, json.Unmarshal([]byte(raw), &c))
	expected := Config{
		Issuers: []IssuerConfig{
			{
				Name:     "test",
				JWKSURL:  "https://localhost:8080",
				CacheTTL: jsontime.Duration(time.Minute),
			},
		},
	}
	assert.Equal(t, expected, c)
}
