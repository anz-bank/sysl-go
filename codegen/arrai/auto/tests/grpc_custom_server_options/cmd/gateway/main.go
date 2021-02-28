package main

import (
	"context"
	"fmt"
	"os"

	pb "grpc_custom_server_options/internal/gen/pb/gateway"
	gateway "grpc_custom_server_options/internal/gen/pkg/servers/gateway"

	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
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

func Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("expected to receive super metadata in ctx, but alas")
	}
	return &pb.HelloResponse{
		Content: fmt.Sprintf("%s %v", req.Content, map[string][]string(md)),
	}, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx,
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
	ctx := log.PutLogger(context.Background(), log.NewDefaultLogger())

	handleError := func(err error) {
		if err != nil {
			log.Error(ctx, err, "something goes wrong")
			os.Exit(1)
		}
	}

	srv, err := newAppServer(ctx)
	handleError(err)
	err = srv.Start()
	handleError(err)
}
