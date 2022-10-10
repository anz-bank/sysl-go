package main

import (
    "context"

    "github.com/anz-bank/sysl-go/core"

    "subdir_tests/grpc/internal/gen/pkg/servers/grpc"
)

type AppConfig struct {
    // Define app-level config fields here.
}

func main() {
    subdir.Serve(context.Background(),
        func(ctx context.Context, config AppConfig) (*subdir.GrpcServiceInterface, *core.Hooks, error) {
            // Perform one-time setup based on config here.
            return &subdir.GrpcServiceInterface{
                // Add handlers here.
            }, nil, nil
        },
    )
}
