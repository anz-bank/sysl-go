package main

import (
    "context"
    "log"

    "github.com/anz-bank/sysl-go/core"

    "template_gen/gen/pkg/servers/Template"
)

func main() {
    type AppConfig struct {
        // Define app-level config fields here.
    }
    log.Fatal(template.Serve(context.Background(),
        func(ctx context.Context, config AppConfig) (*template.ServiceInterface, *core.Hooks, error) {
            // Perform one-time setup based on config here.
            return &template.ServiceInterface{
                // Add handlers here.
            }, nil, nil
        },
    ))
}
