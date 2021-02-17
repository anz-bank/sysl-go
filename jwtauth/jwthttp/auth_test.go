package jwthttp

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/common"

	"github.com/anz-bank/sysl-go/jsontime"
	"github.com/anz-bank/sysl-go/jwtauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContext() context.Context {
	return context.Background()
}

type badAuthenticator struct {
	err error
}

func (b badAuthenticator) Authenticate(ctx context.Context, token string) (jwtauth.Claims, error) {
	return jwtauth.Claims{}, b.err
}

type goodAuthenticator struct {
	claims jwtauth.Claims
}

func (g goodAuthenticator) Authenticate(ctx context.Context, token string) (jwtauth.Claims, error) {
	return g.claims, nil
}

func TestAuthFromConfig(t *testing.T) {
	ctx := testContext()
	c := &Config{
		Config: jwtauth.Config{},
	}
	auth, err := AuthFromConfig(ctx, c, nil)
	assert.NoError(t, err)
	assert.NotNil(t, auth)
}

func TestAuthFromConfigError(t *testing.T) {
	ctx := testContext()
	c := &Config{
		Config: jwtauth.Config{
			Issuers: []jwtauth.IssuerConfig{
				{
					Name:     "BADISSUER",
					JWKSURL:  "*://BADURL", // This will induce an error on startup
					CacheTTL: jsontime.Duration(time.Minute),
				},
			},
		},
	}
	_, err := AuthFromConfig(ctx, c, func(string) *http.Client { return nil })
	assert.Error(t, err)
}

func TestWithUnauthHandler(t *testing.T) {
	auth := &Auth{}
	handler := func(http.ResponseWriter, *http.Request, error) {}
	auth2 := auth.WithUnauthHandler(handler)

	// This proves 2 things
	// the handler is set in the new one
	// It has not overridden the old one
	assert.Nil(t, auth.UnauthHandler)
	assert.NotNil(t, auth2.UnauthHandler)
}

func TestAuthorizePassesAuthorizedRequests(t *testing.T) {
	// Test a authorizor middleware with no authorizors still authenticates (validates jwts)
	auth := &Auth{
		Headers: []string{"Authorization"},
		// returns an error when attempting to authenticate
		Authenticator: goodAuthenticator{},
	}
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := auth.Auth()(endpoint)
	server := common.NewHTTPTestServer(handler)
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthAllowAnonPassthrough(t *testing.T) {
	// Test a authorizor anonymous middleware passes requests with no Authorization header
	auth := &Auth{
		Headers: []string{"Authorization"},
		// returns an error when attempting to authenticate
		Authenticator: badAuthenticator{err: errors.New("bad auth")},
	}
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := auth.AuthAllowAnon()(endpoint)
	server := common.NewHTTPTestServer(handler)

	// with no auth header
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestAuthAllowAnonReject(t *testing.T) {
	// Test a authorizor anonymous middleware rejects a request with invalid JWT
	auth := &Auth{
		Headers: []string{"Authorization"},
		// returns an error when attempting to authenticate
		Authenticator: badAuthenticator{err: errors.New("bad auth")},
	}
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := auth.AuthAllowAnon()(endpoint)
	server := common.NewHTTPTestServer(handler)

	// with bad auth header
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Add("Authorization", "Bearer BAD")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAuthNOAuthenticates(t *testing.T) {
	// Test a authorizor middleware with no authorizors still authenticates (validates jwts)
	auth := &Auth{
		Headers: []string{"Authorization"},
		// returns an error when attempting to authenticate
		Authenticator: badAuthenticator{errors.New("AuthError")},
	}
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := auth.Auth()(endpoint)
	server := common.NewHTTPTestServer(handler)
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAuthNOAuthorizes(t *testing.T) {
	// Test an auth middleware applies authorizors
	auth := &Auth{
		Headers: []string{"Authorization"},
		// returns an error when attempting to authenticate
		Authenticator: badAuthenticator{errors.New("AuthError")},
	}

	authorizor := jwtauth.AuthoriseFunc(func(c jwtauth.Claims) error { return errors.New("Auth error") })
	endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	handler := auth.WithAuthorisers(authorizor).Auth()(endpoint)

	server := common.NewHTTPTestServer(handler)
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetBearer(t *testing.T) {
	auth := &Auth{
		Headers:       []string{"Authorization"},
		Authenticator: goodAuthenticator{jwtauth.Claims{"scope": "a"}},
	}
	token := "Bearer token"
	headers := http.Header{
		"Authorization": []string{token},
	}
	assert.Equal(t, "token", auth.getBearer(headers))
}
