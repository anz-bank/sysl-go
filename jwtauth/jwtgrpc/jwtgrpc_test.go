package jwtgrpc_test

import (
	"context"
	"testing"

	"github.com/anz-bank/sysl-go/jwtauth/jwtgrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestGetBearer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		md := metadata.New(map[string]string{
			"Authorization": "Bearer my-token",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		token, err := jwtgrpc.GetBearerFromIncomingContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, "my-token", token)
	})

	t.Run("NoMetadata", func(t *testing.T) {
		ctx := context.Background()
		_, err := jwtgrpc.GetBearerFromIncomingContext(ctx)
		require.Error(t, err)
		assert.Equal(t, jwtgrpc.ErrNoAuthHeader, err)
	})

	t.Run("NoAuthHeader", func(t *testing.T) {
		md := metadata.New(map[string]string{})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err := jwtgrpc.GetBearerFromIncomingContext(ctx)
		require.Error(t, err)
		assert.Equal(t, jwtgrpc.ErrNoAuthHeader, err)
	})

	t.Run("EmptyAuthHeader", func(t *testing.T) {
		md := metadata.New(map[string]string{
			"Authorization": "",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err := jwtgrpc.GetBearerFromIncomingContext(ctx)
		require.Error(t, err)
		assert.Equal(t, jwtgrpc.ErrParseJWT, err)
	})

	t.Run("NotBearer", func(t *testing.T) {
		md := metadata.New(map[string]string{
			"Authorization": "Basic OIONFSJBONLVDS",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err := jwtgrpc.GetBearerFromIncomingContext(ctx)
		require.Error(t, err)
		assert.Equal(t, jwtgrpc.ErrParseJWT, err)
	})
}
