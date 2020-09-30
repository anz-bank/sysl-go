package jwtauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// RemoteJWKSIssuer is a Verifier that retrieves and stores a jwks from a remote issuer.
//
// Assumes the public key is served at GET {url}/.well-known/jwks.json.
type RemoteJWKSIssuer struct {
	url    string
	client *http.Client
	cache  *jwksCache
}

// NewRemoteJWKSIssuer creates a new RemoteJWKSIssuer.
// UNSTABLE: This API should be avoided in favour of `VerifierFromIssuerConfig()`.
//
// keeps a cache of the jwks so it does not have to poll the remote jwks endpoint for
// every verify.
// cacheTTL defines the expiry time of the cache.
// cacheRefresh defines a cycle-time for a pre-emptive refresh background process (where cacheRefresh > 0).
func NewRemoteJWKSIssuer(ctx context.Context, issuer string, issuerURL string, client *http.Client, cacheTTL time.Duration,
	cacheRefresh time.Duration) (*RemoteJWKSIssuer, error) {
	// Verify the issuer url is valid by parsing it
	if _, err := url.Parse(issuerURL); err != nil {
		return nil, err
	}
	if cacheTTL == 0 {
		return nil, errors.New("Must have a non-zero cache ttl")
	}
	r := &RemoteJWKSIssuer{
		url:    issuerURL,
		client: client,
		cache: &jwksCache{
			ttl: cacheTTL,
		},
	}
	if _, err := r.refreshCache(); err != nil {
		pkgLogger.Debug(ctx, "Error initializing jwks cache for remote issuer:", issuer, err)
	}
	if cacheRefresh > 0 {
		go func() {
			for {
				time.Sleep(cacheRefresh)
				pkgLogger.Debug(ctx, "Refreshing JWKS Cache")
				if _, err := r.refreshCache(); err != nil {
					pkgLogger.Debug(ctx, "Error in Refreshing Cache for JWKS API", err)
				}
			}
		}()
	}
	return r, nil
}

// Verify implements the Verify interface for RemoteJWKSIssuer.
func (r *RemoteJWKSIssuer) Verify(token *jwt.JSONWebToken, claims ...interface{}) error {
	headers := token.Headers
	if len(headers) != 1 {
		// This is currently a limitation of this library
		return &AuthError{
			Code:  AuthErrCodeInvalidJWT,
			Cause: errors.New("Token must have one header"),
		}
	}
	kid := headers[0].KeyID

	keys, err := r.cache.getKey(kid)
	if err != nil {
		jwks, err := r.refreshCache()
		if err != nil {
			return &AuthError{
				Code:  AuthErrCodeUnknown,
				Cause: errors.Wrap(err, "Unable to refresh jwks"),
			}
		}
		keys = jwks.Key(kid)
	}
	if len(keys) == 0 {
		return &AuthError{
			// no key with matching key id, cannot trust source
			Code:  AuthErrCodeUntrustedSource,
			Cause: errors.New("No mathing key id for incoming jwt"),
		}
	}
	if err := token.Claims(keys[0], claims...); err != nil {
		return &AuthError{
			Code:  AuthErrCodeBadSignature,
			Cause: errors.Wrap(err, "jwt verify error"),
		}
	}
	return nil
}

func (r *RemoteJWKSIssuer) refreshCache() (*jose.JSONWebKeySet, error) {
	resp, err := r.client.Get(r.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks fetch error: Received status %d from issuer", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading jwks response")
	}
	var newjwks jose.JSONWebKeySet
	if err := json.Unmarshal(body, &newjwks); err != nil {
		return nil, err
	}
	r.cache.Put(&newjwks)
	return &newjwks, nil
}

type jwksCache struct {
	sync.RWMutex
	cache   *jose.JSONWebKeySet
	ttl     time.Duration
	setTime time.Time
}

// Error is returned to distinguish between an empty key and an expired cache.
func (c *jwksCache) getKey(kid string) ([]jose.JSONWebKey, error) {
	c.RLock()
	defer c.RUnlock()
	if c.ttl < time.Since(c.setTime) {
		return nil, errors.New("Cache expired")
	}
	return c.cache.Key(kid), nil
}

func (c *jwksCache) Put(jwks *jose.JSONWebKeySet) {
	c.Lock()
	defer c.Unlock()
	c.cache = jwks
	c.setTime = time.Now()
}
