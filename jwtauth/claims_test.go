package jwtauth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddClaimsToContext(t *testing.T) {
	claims := Claims{"scope": "email address phone"} // Ref: [RFC8693, Section 4.2]  https://tools.ietf.org/html/rfc8693
	ctx := AddClaimsToContext(context.Background(), claims)
	require.NotNil(t, ctx)

	inCtx, ok := ctx.Value(claimsKey).(Claims)
	require.True(t, ok)
	assert.Equal(t, claims, inCtx)

	// Check claims inside ctx is a copy of what was added.
	claims["scope"] = "banana"
	assert.Equal(t, "email address phone", inCtx["scope"].(string))
}

func TestGetClaimsFromContext(t *testing.T) {
	claims := Claims{"scope": "a"} // Ref: [RFC8693, Section 4.2]  https://tools.ietf.org/html/rfc8693
	ctx := AddClaimsToContext(context.Background(), claims)
	require.NotNil(t, ctx)

	inCtx, ok := ctx.Value(claimsKey).(Claims)
	require.True(t, ok)
	assert.Equal(t, claims, inCtx)

	retrieved, ok := GetClaimsFromContext(ctx)
	require.True(t, ok)
	require.NotNil(t, retrieved)

	// Check claims obtained from ctx is a copy of what is in ctx.
	retrieved["scope"] = "b"
	assert.Equal(t, "a", inCtx["scope"].(string))
}

func TestGetClaimsFromContextEmpty(t *testing.T) {
	claims, ok := GetClaimsFromContext(context.Background())
	assert.False(t, ok)
	assert.Empty(t, claims)
}
