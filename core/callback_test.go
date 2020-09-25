package core

import (
	"fmt"
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

func TestResolveGrpcDialOptionsDefaultCase(t *testing.T) {
	hooks := &Hooks{}

	opts, err := ResolveGrpcDialOptions("dummy_target", hooks, nil)
	require.NoError(t, err)
	expectedOpts, err := config.DefaultGrpcDialOptions(nil)
	require.NoError(t, err)

	// we want to check these things are the same, but they are things
	// with different memory addresses wrapping function pointers, so let's just check the length.
	require.Equal(t, len(expectedOpts), len(opts))
}

func TestResolveGrpcDialOptionsCannotOverrideAndAddDialOptionsSimultaneously(t *testing.T) {
	// inconsistent config
	hooks := &Hooks{
		AdditionalGrpcDialOptions: []grpc.DialOption{grpc.WithWriteBufferSize(123)},
		OverrideGrpcDialOptions: func(_ string, _ *config.CommonGRPCDownstreamData) ([]grpc.DialOption, error) {
			return []grpc.DialOption{grpc.WithWriteBufferSize(123456)}, nil
		},
	}

	_, err := ResolveGrpcDialOptions("dummy_target", hooks, nil)
	require.Equal(t, "Hooks.AdditionalGrpcDialOptions and Hooks.OverrideGrpcDialOptions cannot both be set", err.Error())
}

func TestResolveGrpcDialOptionsCanAddDialOptions(t *testing.T) {
	hooks := &Hooks{
		AdditionalGrpcDialOptions: []grpc.DialOption{grpc.WithWriteBufferSize(123)},
	}
	expectedOpts, _ := config.DefaultGrpcDialOptions(nil)
	expectedOpts = append(expectedOpts, hooks.AdditionalGrpcDialOptions...)

	opts, err := ResolveGrpcDialOptions("dummy_target", hooks, nil)
	require.NoError(t, err)

	require.Equal(t, len(expectedOpts), len(opts))
	require.Equal(t, hooks.AdditionalGrpcDialOptions[0], opts[len(opts)-1])
}

func TestResolveGrpcDialOptionsCanOverrideDialOptions(t *testing.T) {
	myCustomOptions := []grpc.DialOption{grpc.WithWriteBufferSize(123456), grpc.WithReadBufferSize(1)}
	hooks := &Hooks{
		OverrideGrpcDialOptions: func(_ string, _ *config.CommonGRPCDownstreamData) ([]grpc.DialOption, error) {
			return myCustomOptions, nil
		},
	}
	expectedOpts := myCustomOptions

	opts, err := ResolveGrpcDialOptions("dummy_target", hooks, nil)
	require.NoError(t, err)
	require.Equal(t, expectedOpts, opts)
}

func TestResolveGrpcDialOptionsCanOverrideDialOptionsPerTargetService(t *testing.T) {
	customOptionsForFoo := []grpc.DialOption{grpc.WithWriteBufferSize(123456), grpc.WithReadBufferSize(1)}
	customOptionsForBarr := []grpc.DialOption{grpc.WithReadBufferSize(999)}

	hooks := &Hooks{
		OverrideGrpcDialOptions: func(serviceName string, _ *config.CommonGRPCDownstreamData) ([]grpc.DialOption, error) {
			switch serviceName {
			case "foo":
				return customOptionsForFoo, nil
			case "barr":
				return customOptionsForBarr, nil
			default:
				return nil, fmt.Errorf("wat?")
			}
		},
	}

	_, err := ResolveGrpcDialOptions("unexpected_target", hooks, nil)
	require.Error(t, err)

	actualFooOpts, err := ResolveGrpcDialOptions("foo", hooks, nil)
	require.NoError(t, err)
	require.Equal(t, customOptionsForFoo, actualFooOpts)

	actualBarrOpts, err := ResolveGrpcDialOptions("barr", hooks, nil)
	require.NoError(t, err)
	require.Equal(t, customOptionsForBarr, actualBarrOpts)
}
