package main

import (
    "context"

    "github.com/anz-bank/sysl-go/core"

    "subdir_tests/rest/internal/gen/pkg/servers/rest"
)

type AppConfig struct {
    // Define app-level config fields here.
}

func main() {
    rest_subdir.Serve(context.Background(),
        func(ctx context.Context, config AppConfig) (*rest_subdir.ServiceInterface, *core.Hooks, error) {
            // Perform one-time setup based on config here.
            return &rest_subdir.ServiceInterface{
                // Add handlers here.
            }, nil, nil
        },
    )
}
