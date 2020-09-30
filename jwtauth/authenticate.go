package jwtauth

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Authenticator can authenticate raw tokens.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (Claims, error)
}

// Verifier defines an interface that can verify an already parsed jwt token.
//
// Intended use is in the StdAuthenticator, where each verifier corresponds to a single named token issuer.
type Verifier interface {
	Verify(token *jwt.JSONWebToken, claims ...interface{}) error
}

// StdAuthenticator is the standard jwt authenticator.
//
// Keeps track of multiple verifiers. Authenticates jwts using the
// iss and kid fields in the jwt to pick a public key from multiple possible
// issuers and keys.
type StdAuthenticator struct {
	Verifiers map[string]Verifier
}

// Authenticate authenticates a jwt and returns the extracted claims, or an
// error if any occur.
func (a *StdAuthenticator) Authenticate(ctx context.Context, raw string) (Claims, error) {
	token, err := jwt.ParseSigned(raw)
	if err != nil {
		pkgLogger.Debug(ctx, "error parsing jwt:", err)
		return Claims{}, &AuthError{
			Code:  AuthErrCodeInvalidJWT,
			Cause: errors.Wrap(err, "jwt parse error"),
		}
	}

	// Extract the issuer
	var insecureClaims jwt.Claims
	if err := token.UnsafeClaimsWithoutVerification(&insecureClaims); err != nil {
		pkgLogger.Debug(ctx, "error extracting claims:", err)
		return Claims{}, &AuthError{
			Code:  AuthErrCodeUnknown,
			Cause: errors.Wrap(err, "jwt verify error"),
		}
	}
	if err := insecureClaims.ValidateWithLeeway(jwt.Expected{Time: time.Now()}, time.Second); err != nil {
		pkgLogger.Debug(ctx, "jwt expired")
		return Claims{}, &AuthError{
			Code:  AuthErrCodeInvalidJWT,
			Cause: err,
		}
	}
	verifier, ok := a.Verifiers[insecureClaims.Issuer]
	if !ok {
		pkgLogger.Debugf(ctx, "issuer not registered: %s", insecureClaims.Issuer)
		return Claims{}, &AuthError{
			Code:  AuthErrCodeUntrustedSource,
			Cause: fmt.Errorf("issuer not registered: %s", insecureClaims.Issuer),
		}
	}

	// Verify the token and populate claims
	var claims Claims
	if err := verifier.Verify(token, &claims); err != nil {
		pkgLogger.Debug(ctx, err)
		return Claims{}, err // Don't wrap this error
	}
	return claims, nil
}

// InsecureAuthenticator does not attempt to verify the signature of a jwt.
//
// USE ONLY IN TESTING.
type InsecureAuthenticator struct{}

// Authenticate implements the Authenticator interface.
func (i InsecureAuthenticator) Authenticate(ctx context.Context, raw string) (Claims, error) {
	token, err := jwt.ParseSigned(raw)
	if err != nil {
		pkgLogger.Debug(ctx, "jwt parse error:", err)
		return Claims{}, &AuthError{
			Code:  AuthErrCodeInvalidJWT,
			Cause: errors.Wrap(err, "jwt parse error"),
		}
	}
	var insecureClaims Claims
	if err := token.UnsafeClaimsWithoutVerification(&insecureClaims); err != nil {
		pkgLogger.Debug(ctx, "jwt verify error:", err)
		return Claims{}, &AuthError{
			Code:  AuthErrCodeUnknown,
			Cause: errors.Wrap(err, "jwt verify error"),
		}
	}
	return insecureClaims, nil
}
