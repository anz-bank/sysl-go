package jwtauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testClient() (url string, client *http.Client) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Context-Type", "application/json")
		_, _ = w.Write([]byte(testJWKS))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	return server.URL, server.Client()
}

func TestRemoteJWKSVerify(t *testing.T) {
	url, client := testClient()
	ctx := testContext()
	v, err := NewRemoteJWKSIssuer(ctx, "test-issuer", url, client, time.Minute, 0)
	require.NoError(t, err)
	require.NotNil(t, v)

	token := issueTestJWT()
	jwtToken, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	require.NoError(t, err)
	require.NotNil(t, jwtToken)
	var claims Claims
	require.NoError(t, v.Verify(jwtToken, &claims))
}

func TestRemoteJWKSVerifyCacheExpired(t *testing.T) {
	url, client := testClient()
	// Handcraft cache that looks like already expired
	v := &RemoteJWKSIssuer{
		url:    url,
		client: client,
		cache: &jwksCache{
			ttl: time.Minute,
		},
	}
	token := issueTestJWT()
	jwtToken, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	require.NoError(t, err)
	require.NotNil(t, jwtToken)
	var claims Claims
	require.NoError(t, v.Verify(jwtToken, &claims))

	// Check the cache has been refreshed
	assert.NotZero(t, v.cache.setTime)
}

func TestRemoteJWKSVerifyCannotRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	v := RemoteJWKSIssuer{
		url:    server.URL,
		client: server.Client(),
		cache: &jwksCache{
			ttl: time.Minute,
		},
	}
	token := issueTestJWT()
	jwtToken, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	require.NoError(t, err)
	require.NotNil(t, jwtToken)

	var claims Claims
	err = v.Verify(jwtToken, &claims)
	require.Error(t, err)

	require.IsType(t, &AuthError{}, err)
	autherr, _ := err.(*AuthError)
	assert.Equal(t, AuthErrCodeUnknown, autherr.Code)
}

func TestRemoteJWKSNotAvailableOnStartup(t *testing.T) {
	ctx := testContext()
	badJWKS := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }
	goodJWKS := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Context-Type", "application/json")
		_, _ = w.Write([]byte(testJWKS))
	}
	activeJWKS := badJWKS
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { activeJWKS(w, r) }))
	v, err := NewRemoteJWKSIssuer(ctx, "test-issuer", server.URL, http.DefaultClient, time.Minute, 0)
	assert.NoError(t, err)

	activeJWKS = goodJWKS
	token := issueTestJWT()
	jwtToken, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	require.NoError(t, err)
	require.NotNil(t, jwtToken)
	var claims Claims
	require.NoError(t, v.Verify(jwtToken, &claims))
}

func TestRemoteJWKSVerifyFailsUntrustedJWT(t *testing.T) {
	ctx := testContext()
	url, client := testClient()
	v, err := NewRemoteJWKSIssuer(ctx, "test-issuer", url, client, time.Minute, 0)
	require.NoError(t, err)
	token := issueUntrustedTestJWT()
	jwtToken, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	require.NoError(t, err)
	require.NotNil(t, jwtToken)

	var claims Claims
	err = v.Verify(jwtToken, &claims)
	require.Error(t, err)

	require.IsType(t, &AuthError{}, err)
	autherr, _ := err.(*AuthError)
	assert.Equal(t, AuthErrCodeUntrustedSource, autherr.Code)
}

func TestRemoteJWKSVerifyFailsMaliciousJWT(t *testing.T) {
	// This test issues a jwt claiming to be from testissuer with a matching key id,
	// but was not actually signed with the correct public key
	ctx := testContext()
	url, client := testClient()
	v, err := NewRemoteJWKSIssuer(ctx, "test-issuer", url, client, time.Minute, 0)
	require.NoError(t, err)
	token := issueMaliciousTestJWT()
	jwtToken, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	require.NoError(t, err)
	require.NotNil(t, jwtToken)

	var claims Claims
	err = v.Verify(jwtToken, &claims)
	require.Error(t, err)

	require.IsType(t, &AuthError{}, err)
	autherr, _ := err.(*AuthError)
	assert.Equal(t, AuthErrCodeBadSignature, autherr.Code)
}

// JWKS cache tests

func TestNewRemoteIssuerBadURL(t *testing.T) {
	ctx := testContext()
	_, client := testClient()
	_, err := NewRemoteJWKSIssuer(ctx, "test-issuer", "*://BADURL", client, time.Millisecond, 0)
	assert.Error(t, err)
}

func TestNewRemoteIssuerZeroTTL(t *testing.T) {
	ctx := testContext()
	_, err := NewRemoteJWKSIssuer(ctx, "test-issuer", "http://localhost:8080/.well-known/jwks.json", nil, 0, time.Minute)
	assert.Error(t, err)
}

func TestNewRemoteIssuerBadRemote(t *testing.T) {
	ctx := testContext()
	client := &http.Client{
		Transport: rtFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("Error") }),
	}
	auth, err := NewRemoteJWKSIssuer(ctx, "test-issuer", "http://localhost:8080", client, time.Millisecond, 0)
	assert.NoError(t, err)
	assert.NotNil(t, auth)
}

func TestRefreshCacheSetsTime(t *testing.T) {
	ctx := testContext()
	url, client := testClient()
	v, err := NewRemoteJWKSIssuer(ctx, "test-issuer", url, client, time.Millisecond, 0)
	require.NoError(t, err)
	_, err = v.refreshCache()
	require.NoError(t, err)
	assert.NotZero(t, v.cache.setTime)
}

type rtFunc func(*http.Request) (*http.Response, error)

func (r rtFunc) RoundTrip(req *http.Request) (*http.Response, error) { return r(req) }
func TestRefreshCacheClientError(t *testing.T) {
	client := &http.Client{
		Transport: rtFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("Error") }),
	}
	v := RemoteJWKSIssuer{
		url:    "http://localhost:8080",
		client: client,
		cache: &jwksCache{
			ttl: time.Millisecond,
		},
	}
	_, err := v.refreshCache()
	assert.Error(t, err)
	assert.Zero(t, v.cache.setTime)
}

func TestRefreshCacheNotOK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	server := httptest.NewServer(handler)
	v := RemoteJWKSIssuer{
		url:    server.URL,
		client: server.Client(),
		cache: &jwksCache{
			ttl: time.Millisecond,
		},
	}
	_, err := v.refreshCache()
	assert.Error(t, err)
	assert.Zero(t, v.cache.setTime)
}

func TestRefreshCacheBadResponseBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`NOT JSON`)) })
	server := httptest.NewServer(handler)
	v := RemoteJWKSIssuer{
		url:    server.URL,
		client: server.Client(),
		cache: &jwksCache{
			ttl: time.Millisecond,
		},
	}
	_, err := v.refreshCache()
	assert.Error(t, err)
	assert.Zero(t, v.cache.setTime)
}

func TestJWKSCacheClearAfterTTL(t *testing.T) {
	ctx := testContext()
	ttl := 10 * time.Millisecond
	url, client := testClient()
	v, err := NewRemoteJWKSIssuer(ctx, "test-issuer", url, client, ttl, 0)
	require.NoError(t, err)
	require.NotNil(t, v)
	_, err = v.cache.getKey("test")
	require.NoError(t, err)

	time.Sleep(2 * ttl)

	// This should now return an error, cache should be cleared
	_, err = v.cache.getKey("test")
	assert.Error(t, err)
}

// This test is unstable due to how the refresh loop works. This will probably be fixed with the cache uplift.
func TestJWKSCacheRefreshOK(t *testing.T) {
	t.Skip("todo fixme unstable test needs rework")
	ctx := testContext()
	refreshDelay := 10 * time.Millisecond
	url, client := testClient()
	v, err := NewRemoteJWKSIssuer(ctx, "test-issuer", url, client, time.Minute, refreshDelay)
	require.NoError(t, err)
	require.NotNil(t, v)
	k, err := v.cache.getKey("test")
	require.NoError(t, err)
	require.NotNil(t, k)

	// Not locking here produces a race condition caught by -race
	// this is NOT a race condition in the code, but in this test
	v.cache.Lock()
	v.cache.cache.Keys = nil
	v.cache.Unlock()
	time.Sleep(refreshDelay + 5*time.Millisecond)

	// This should now return the refreshed key
	k, err = v.cache.getKey("test")
	require.NoError(t, err)
	require.NotNil(t, k)
}
