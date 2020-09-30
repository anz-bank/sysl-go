package main

import (
	"context"
	"fmt"
	"os"

	pb "grpc_custom_server_options/gen/pb/gateway"
	gateway "grpc_custom_server_options/gen/pkg/servers/Gateway"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type KVPair struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// AppConfig is an example of using custom application config to control the
// behaviour of custom grpc.ServerOption configuration.
type AppConfig struct {
	CustomMetadata []KVPair `yaml:"customMetadata"`

	// these keys are application-defined, there is nothing special about them.
	SetAdditionalGrpcServerOptions bool `yaml:"setAdditionalGrpcServerOptions"`
	SetOverrideGrpcServerOptions   bool `yaml:"setOverrideGrpcServerOptions"`
}

func makeCustomGrpcMetadataInjector(key, value string) grpc.UnaryServerInterceptor {
	f := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var md metadata.MD
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			md = md.Copy()
		} else {
			md = metadata.New(nil)
		}
		md.Set(key, value)
		ctxPrime := metadata.NewIncomingContext(ctx, md)
		return handler(ctxPrime, req)
	}
	return f
}

func Hello(ctx context.Context, req *pb.HelloRequest, client gateway.HelloClient) (*pb.HelloResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("expected to receive super metadata in ctx, but alas")
	}
	return &pb.HelloResponse{
		Content: fmt.Sprintf("%s %v", req.Content, map[string][]string(md)),
	}, nil
}

func application(ctx context.Context) error {
	return gateway.Serve(ctx,
		func(ctx context.Context, cfg AppConfig) (*gateway.GrpcServiceInterface, *core.Hooks, error) {
			// We can access the AppConfig here to help define our
			// custom grpc.ServerOption configuration
			opts := []grpc.ServerOption{}
			for _, kvpair := range cfg.CustomMetadata {
				f := makeCustomGrpcMetadataInjector(kvpair.Key, kvpair.Value)
				opts = append(opts, grpc.ChainUnaryInterceptor(f))
			}

			// example of using a hook to append gRPC server options
			myHooks := &core.Hooks{}
			if cfg.SetAdditionalGrpcServerOptions {
				myHooks.AdditionalGrpcServerOptions = opts
			}
			// example of using a hook to override gRPC server options
			if cfg.SetOverrideGrpcServerOptions {
				myHooks.OverrideGrpcServerOptions = func(_ context.Context, _ *config.CommonServerConfig) ([]grpc.ServerOption, error) {
					return opts, nil
				}
			}
			return &gateway.GrpcServiceInterface{
					Hello: Hello,
				},
				myHooks,
				nil
		},
	)
}

func main() {
	// initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).Onto(context.Background())

	err := application(ctx)

	if err != nil {
		log.Error(ctx, err)
		os.Exit(1)
	}
}
