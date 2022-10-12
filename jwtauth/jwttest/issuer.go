package jwttest

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"net/http"

	"github.com/anz-bank/sysl-go/jwtauth"
	"github.com/anz-bank/sysl-go/jwtauth/jwthttp"
	"github.com/google/uuid"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Issuer is a test issuer to issue jwts with claims on demand
//
// Intended for easy testing. Use Issue to create a new jwt that can be authenticated by the public key
//
// ServeHTTP serves the public key as a jwks payload.
type Issuer struct {
	jose.Signer
	PubKey *jose.JSONWebKey
	Name   string
}

// NewIssuer creates a new jwt token issuer with a RS256 key of given size.
func NewIssuer(name string, keysize int) (Issuer, error) {
	pub, priv, err := GenRSKeys(keysize)
	if err != nil {
		return Issuer{}, err
	}
	sigKey := jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       priv,
	}
	sig, err := jose.NewSigner(sigKey, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			jose.HeaderType: "jwt",
		},
	})
	return Issuer{
		Signer: sig,
		PubKey: pub,
		Name:   name,
	}, err
}

// Issue issues a new jwt with the given claims.
func (i Issuer) Issue(claims jwtauth.Claims) (string, error) {
	return i.IssueFromMap(claims)
}

// IssueFromMap issues a new jwt with the given claims.
func (i Issuer) IssueFromMap(claims map[string]interface{}) (string, error) {
	// If claims contains an issuer name, tester has set it deliberately, don't touch it
	issuerNameWrapper, ok := claims["iss"]
	if !ok {
		claims["iss"] = i.Name
	} else {
		issuerName, ok := issuerNameWrapper.(string)
		if ok && issuerName == "" {
			claims["iss"] = i.Name
		}
	}
	return jwt.Signed(i.Signer).Claims(claims).CompactSerialize()
}

// Verify implements jwtauth.Verifier.
func (i Issuer) Verify(token *jwt.JSONWebToken, claims ...interface{}) error {
	return token.Claims(i.PubKey, claims...)
}

// Authenticate implements jwtauth.Authenticator
//
// Checks the issuer on the inbound jwt matches the name of the issuer.
func (i Issuer) Authenticate(ctx context.Context, token string) (jwtauth.Claims, error) {
	parsed, err := jwt.ParseSigned(token)
	if err != nil {
		return jwtauth.Claims{}, &jwtauth.AuthError{
			Code:  jwtauth.AuthErrCodeInvalidJWT,
			Cause: err,
		}
	}
	var insecureClaims jwt.Claims
	if err := parsed.UnsafeClaimsWithoutVerification(&insecureClaims); err != nil {
		return jwtauth.Claims{}, &jwtauth.AuthError{
			Code:  jwtauth.AuthErrCodeInvalidJWT,
			Cause: err,
		}
	}
	if insecureClaims.Issuer != i.Name {
		return jwtauth.Claims{}, &jwtauth.AuthError{
			Code:  jwtauth.AuthErrCodeUntrustedSource,
			Cause: err,
		}
	}
	var claims jwtauth.Claims
	if err := i.Verify(parsed, &claims); err != nil {
		return jwtauth.Claims{}, &jwtauth.AuthError{
			Code:  jwtauth.AuthErrCodeBadSignature,
			Cause: err,
		}
	}
	return claims, nil
}

// Authenticator produces a standard authenticator with only this issuer as a trusted issuer.
func (i Issuer) Authenticator() jwtauth.Authenticator {
	return &jwtauth.StdAuthenticator{
		Verifiers: map[string]jwtauth.Verifier{
			i.Name: i,
		},
	}
}

func (i Issuer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			*i.PubKey,
		},
	}
	payload, err := json.Marshal(jwks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

// HTTPAuth returns an httpauth middleware struct with the issuer set as the authenticator.
func (i Issuer) HTTPAuth() *jwthttp.Auth {
	return &jwthttp.Auth{
		Headers:       []string{"Authorization"},
		Authenticator: i,
	}
}

// Not sure what we want to do about this, we should be able to stand up a full blown test issuer where we can ask it for tokens and public keys
//
// // ServeToken serves a token with the requested claims
// //
// // Request body should be a json payload with the claims the requestor wants in their jwt
// func (i Issuer) ServeToken(w http.ResponseWriter, r *http.Request) {
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.Write([]byte("Failed to read body"))
// 		return
// 	}
// 	var claims jwtauth.Claims
// 	if err := json.Unmarshal(body, &claims); err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.Write([]byte("Invalid claims payload"))
// 		return
// 	}
// 	token, err := i.Issue(claims)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Failed to issue token with requested claims"))
// 		return
// 	}
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(token))
// }

// GenRSKeys generates a public/private key pair for signing using RS256 algorithm.
func GenRSKeys(keysize int) (*jose.JSONWebKey, *jose.JSONWebKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, keysize) // TODO: set key size?
	if err != nil {
		return nil, nil, err
	}
	pubKey := privKey.Public()

	kidUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, nil, err
	}

	kid := kidUUID.String()
	alg := jose.RS256

	pub := &jose.JSONWebKey{
		KeyID:        kid,
		Certificates: []*x509.Certificate{},
		Key:          pubKey,
		Algorithm:    string(alg),
		Use:          "sig",
	}

	priv := &jose.JSONWebKey{
		KeyID:        kid,
		Certificates: []*x509.Certificate{},
		Key:          privKey,
		Algorithm:    string(alg),
		Use:          "sig",
	}

	return pub, priv, nil
}
