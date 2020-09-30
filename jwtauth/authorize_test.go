package jwtauth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthorizeFunc(t *testing.T) {
	authFunc := AuthoriseFunc(func(Claims) error {
		return nil
	})
	require.NoError(t, authFunc.Authorise(Claims{}))
}
