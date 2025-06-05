package jwttest

import (
	"encoding/json"
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anz-bank/sysl-go/jwtauth"
)

func asJSONBytes(value interface{}) []byte {
	data, err := json.Marshal(&value)
	if err != nil {
		panic(err)
	}
	return data
}

func TestNewIssuer(t *testing.T) {
	issuer, err := NewIssuer("test", 2048)
	require.NoError(t, err)
	require.NotNil(t, issuer.Signer)
}

func TestIssue(t *testing.T) {
	claims := jwtauth.Claims{
		"scope": "MY.SCOPE",
	}
	issuer, _ := NewIssuer("test", 2048)
	token, err := issuer.Issue(claims)
	require.NoError(t, err)

	// Verify the jwt can be verified by the issuers public key
	parsed, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	require.NoError(t, err)
	var claimsFromToken jwtauth.Claims
	require.NoError(t, parsed.Claims(issuer.PubKey, &claimsFromToken))
	assert.Equal(t, asJSONBytes(claims["scope"]), asJSONBytes(claimsFromToken["scope"]))
}

func TestIssueFromMap(t *testing.T) {
	claims := map[string]interface{}{
		"scope": "MY.SCOPE",
	}
	issuer, _ := NewIssuer("test", 2048)
	token, err := issuer.IssueFromMap(claims)
	require.NoError(t, err)

	// Verify the jwt can be verified by the issuers public key
	parsed, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	require.NoError(t, err)
	var claimsFromToken jwtauth.Claims
	require.NoError(t, parsed.Claims(issuer.PubKey, &claimsFromToken))
	assert.Equal(t, asJSONBytes(claims["scope"]), asJSONBytes(claimsFromToken["scope"]))
}

func TestGenRSKeys(t *testing.T) {
	pub, priv, err := GenRSKeys(2048)
	require.NoError(t, err)
	require.NotNil(t, pub)
	require.NotNil(t, priv)

	assert.Equal(t, priv.KeyID, pub.KeyID)

	// Check we can create a signer with this key
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: priv}, &jose.SignerOptions{})
	require.NoError(t, err)
	require.NotNil(t, signer)

	// Check the signer can sign and public key can verify a signature
	sig, err := signer.Sign([]byte("Hello World"))
	require.NoError(t, err)
	_, err = sig.Verify(pub.Key)
	require.NoError(t, err)
}
