package main

import (
	"context"
	"grpc_complex_app_name/internal/gen/pkg/servers/encoder_backend"

	"github.com/anz-bank/sysl-go/core"
)

type AppConfig struct {
	// Define app-level config fields here.
}

func main() {
	encoder_backend.Serve(context.Background(),
		func(ctx context.Context, config AppConfig) (*encoder_backend.GrpcServiceInterface, *core.Hooks, error) {
			// Perform one-time setup based on config here.
			return &encoder_backend.GrpcServiceInterface{
				// Add handlers here.
			}, nil, nil
		},
	)
}
