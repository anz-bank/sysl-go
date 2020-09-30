package main

import (
	"context"
	"os"

	pb "grpc_jwt_authorisation/gen/pb/gateway"
	gateway "grpc_jwt_authorisation/gen/pkg/servers/Gateway"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/core"
)

type AppConfig struct {
}

func Hello(ctx context.Context, req *pb.HelloRequest, client gateway.HelloClient) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		Content: "why hello there",
	}, nil
}

func application(ctx context.Context) error {
	return gateway.Serve(ctx,
		func(ctx context.Context, cfg AppConfig) (*gateway.GrpcServiceInterface, *core.Hooks, error) {
			return &gateway.GrpcServiceInterface{
					Hello: Hello,
				},
				&core.Hooks{},
				nil
		},
	)
}

func main() {
	// initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).WithConfigs(log.SetVerboseMode(true)).Onto(context.Background())

	err := application(ctx)

	if err != nil {
		log.Error(ctx, err)
		os.Exit(1)
	}
}
