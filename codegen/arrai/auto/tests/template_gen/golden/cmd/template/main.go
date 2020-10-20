package main

import (
	"context"

	"github.com/anz-bank/sysl-go/core"

	"template_gen/internal/gen/pkg/servers/template"
)

type AppConfig struct {
	// Define app-level config fields here.
}

func main() {
	template.Serve(context.Background(),
		func(ctx context.Context, config AppConfig) (*template.ServiceInterface, *core.Hooks, error) {
			// Perform one-time setup based on config here.
			return &template.ServiceInterface{
				// Add handlers here.
			}, nil, nil
		},
	)
}
