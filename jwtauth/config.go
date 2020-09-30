package jwtauth

import (
	"context"
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go/jsontime"
	"github.com/pkg/errors"
)

// Config defines configuration for the standard authenticator.
type Config struct {
	Issuers []IssuerConfig `json:"issuers"         yaml:"issuers"         mapstructure:"issuers"`
}

// AuthFromConfig constructs a standard authenticator from config.
//
// The client func allows the application to configure clients on a per-issuer
// basis. This is in case various remote issuers have different requirements about
// how to call them.
func AuthFromConfig(ctx context.Context, c *Config, client func(string) *http.Client) (*StdAuthenticator, error) {
	if c == nil {
		return nil, errors.New("AuthConfig: Config must not be nil")
	}
	verifiers := map[string]Verifier{}
	for _, ic := range c.Issuers {
		if ic.Name == "" {
			return nil, errors.New("AuthConfig: Issuer must have a name")
		}
		if _, ok := verifiers[ic.Name]; ok {
			return nil, errors.New("AuthConfig: Issuer names are not unique")
		}
		v, err := VerifierFromIssuerConfig(ctx, ic, client(ic.Name))
		if err != nil {
			return nil, errors.Wrapf(err, "AuthConfig: Error creating verifier for issuer %s", ic.Name)
		}
		verifiers[ic.Name] = v
	}
	return &StdAuthenticator{
		Verifiers: verifiers,
	}, nil
}

// IssuerConfig defines config for issuers for the std authenticator.
type IssuerConfig struct {
	Name         string            `json:"name"                       yaml:"name"                       mapstructure:"name"`
	JWKSURL      string            `json:"jwksUrl,omitempty"          yaml:"jwksUrl,omitempty"          mapstructure:"jwksUrl"`
	CacheTTL     jsontime.Duration `json:"cacheTTL"                   yaml:"cacheTTL"                   mapstructure:"cacheTTL"`
	CacheRefresh jsontime.Duration `json:"cacheRefresh"               yaml:"cacheRefresh"               mapstructure:"cacheRefresh"`
}

// VerifierFromIssuerConfig creates a token verifier from issuer config.
func VerifierFromIssuerConfig(ctx context.Context, i IssuerConfig, client *http.Client) (Verifier, error) {
	if i.JWKSURL != "" {
		return NewRemoteJWKSIssuer(ctx, i.Name, i.JWKSURL, client, time.Duration(i.CacheTTL), time.Duration(i.CacheRefresh))
	}
	return nil, errors.New("jwtauth.Config: Can only have one of SharedSecret, PublicKey or JWKSURL set")
}
