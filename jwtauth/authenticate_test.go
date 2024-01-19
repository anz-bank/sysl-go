package jwtauth

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testContext() context.Context {
	return context.Background()
}

func TestStdAuthenticator(t *testing.T) {
	ctx := testContext()
	auth := &StdAuthenticator{
		Verifiers: map[string]Verifier{
			"test": testVerifier{},
		},
	}
	token := issueTestJWT()
	claims, err := auth.Authenticate(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, claims)
}

func TestAuthenticateExpiredJWT(t *testing.T) {
	ctx := testContext()
	auth := &StdAuthenticator{
		Verifiers: map[string]Verifier{
			"test": testVerifier{},
		},
	}
	token := issueExpiredJWT()
	claims, err := auth.Authenticate(ctx, token)
	require.Error(t, err)
	require.Equal(t, map[string]interface{}{}, claims)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestStdAuthenticatorBadToken(t *testing.T) {
	ctx := testContext()
	auth := &StdAuthenticator{
		Verifiers: map[string]Verifier{
			"test": testVerifier{},
		},
	}
	token := "NOT A JWT"
	claims, err := auth.Authenticate(ctx, token)
	require.Error(t, err)
	require.Equal(t, map[string]interface{}{}, claims)
}

func TestStdAuthenticatorUntrustedSource(t *testing.T) {
	ctx := testContext()
	auth := &StdAuthenticator{
		Verifiers: map[string]Verifier{
			"test": testVerifier{},
		},
	}
	token := issueUntrustedTestJWT()
	claims, err := auth.Authenticate(ctx, token)
	require.Error(t, err)
	require.Equal(t, map[string]interface{}{}, claims)
}

func TestStdAuthenticatorMaliciousSource(t *testing.T) {
	ctx := testContext()
	auth := &StdAuthenticator{
		Verifiers: map[string]Verifier{
			"test": testVerifier{},
		},
	}
	token := issueMaliciousTestJWT()
	claims, err := auth.Authenticate(ctx, token)
	require.Error(t, err)
	require.Equal(t, map[string]interface{}{}, claims)
}

func TestStdAuthenticatorWithActorClaim(t *testing.T) {
	ctx := testContext()
	auth := &StdAuthenticator{
		Verifiers: map[string]Verifier{
			"test": testVerifier{},
		},
	}
	token := issueTestJWTWithActor("foo")
	claims, err := auth.Authenticate(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.Equal(t, "foo", claims["act"].(map[string]interface{})["sub"].(string))
}

func TestInsecureAuthenticatorRejectsNonJWT(t *testing.T) {
	ctx := testContext()
	_, err := InsecureAuthenticator{}.Authenticate(ctx, "NOT A JWT")
	require.Error(t, err)
}

func TestInsecureAuthenticator(t *testing.T) {
	ctx := testContext()
	token := issueUntrustedTestJWT()
	claims, err := InsecureAuthenticator{}.Authenticate(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, claims)
}

type errorVerifier struct{ err error }

func (e errorVerifier) Verify(*jwt.JSONWebToken, ...interface{}) error {
	return e.err
}

func TestAuthenticateVerifierError(t *testing.T) {
	ctx := testContext()
	auth := &StdAuthenticator{
		Verifiers: map[string]Verifier{
			"testissuer": errorVerifier{err: &AuthError{Code: AuthErrCodeUntrustedSource}},
		},
	}
	token := issueTestJWT()
	_, err := auth.Authenticate(ctx, token)
	assert.Error(t, err)

	// Check the error is propagated from the verifier correctly
	require.IsType(t, &AuthError{}, err)
	autherr, _ := err.(*AuthError)
	assert.Equal(t, AuthErrCodeUntrustedSource, autherr.Code)
}

func asJSONBytes(value interface{}) []byte {
	data, err := json.Marshal(&value)
	if err != nil {
		panic(err)
	}
	return data
}

func TestAuthenticateCustomClaims(t *testing.T) {
	ctx := testContext()
	auth := (&StdAuthenticator{
		Verifiers: map[string]Verifier{
			"test": testVerifier{},
		},
	})

	claims := map[string]interface{}{
		"iss":                 "test",
		"_some_private_claim": []string{"1234"},
	}
	token, _ := jwt.Signed(testSigner).Claims(claims).CompactSerialize()

	actual, err := auth.Authenticate(ctx, token)
	require.NoError(t, err)

	expectedPrivateClaimValue := asJSONBytes([]string{"1234"})
	assert.Equal(t, expectedPrivateClaimValue, asJSONBytes(actual["_some_private_claim"]))
}
