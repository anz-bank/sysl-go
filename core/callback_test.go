package core

import (
	"testing"

	"github.com/anz-bank/sysl-go/config"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestResolveGrpcServerOptionsDefaultCase(t *testing.T) {
	hooks := &Hooks{}

	opts, err := ResolveGrpcServerOptions(hooks, nil)
	require.NoError(t, err)
	expectedOpts, err := DefaultGrpcServerOptions(nil)
	require.NoError(t, err)

	// we want to check these things are the same, but they are things
	// with different memory addresses wrapping function pointers, so let's just check the length.
	require.Equal(t, len(expectedOpts), len(opts))
}

func TestResolveGrpcServerOptionsCannotOverrideAndAddServerOptionsSimultaneously(t *testing.T) {
	// inconsistent config
	hooks := &Hooks{
		AdditionalGrpcServerOptions: []grpc.ServerOption{grpc.MaxRecvMsgSize(123)},
		OverrideGrpcServerOptions: func(_ *config.CommonServerConfig) ([]grpc.ServerOption, error) {
			return []grpc.ServerOption{grpc.MaxRecvMsgSize(123456)}, nil
		},
	}

	_, err := ResolveGrpcServerOptions(hooks, nil)
	require.Equal(t, "Hooks.AdditionalGrpcServerOptions and Hooks.OverrideGrpcServerOptions cannot both be set", err.Error())
}

func TestResolveGrpcServerOptionsCanAddServerOptions(t *testing.T) {
	hooks := &Hooks{
		AdditionalGrpcServerOptions: []grpc.ServerOption{grpc.MaxRecvMsgSize(123)},
	}
	expectedOpts, _ := DefaultGrpcServerOptions(nil)
	expectedOpts = append(expectedOpts, hooks.AdditionalGrpcServerOptions...)

	opts, err := ResolveGrpcServerOptions(hooks, nil)
	require.NoError(t, err)

	require.Equal(t, len(expectedOpts), len(opts))
	require.Equal(t, hooks.AdditionalGrpcServerOptions[0], opts[len(opts)-1])
}

func TestResolveGrpcServerOptionsCanOverrideServerOptions(t *testing.T) {
	myCustomOptions := []grpc.ServerOption{grpc.MaxRecvMsgSize(123456), grpc.ReadBufferSize(1)}
	hooks := &Hooks{
		OverrideGrpcServerOptions: func(_ *config.CommonServerConfig) ([]grpc.ServerOption, error) {
			return myCustomOptions, nil
		},
	}
	expectedOpts := myCustomOptions

	opts, err := ResolveGrpcServerOptions(hooks, nil)
	require.NoError(t, err)
	require.Equal(t, expectedOpts, opts)
}
