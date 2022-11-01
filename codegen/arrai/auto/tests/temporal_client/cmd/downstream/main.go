package main

import (
	"context"
	"fmt"

	"github.com/anz-bank/sysl-go/core"

	somedownstream "temporal_client/internal/gen/pkg/servers/downstream"
)

type AppConfig struct {
	// Define app-level config fields here.
}

func main() {
	somedownstream.Serve(context.Background(),
		func(ctx context.Context, config AppConfig) (*somedownstream.ServiceInterface, *core.Hooks, error) {
			// Perform one-time setup based on config here.
			return &somedownstream.ServiceInterface{
				Post: func(ctx context.Context, req *somedownstream.PostRequest) (*somedownstream.SomeResp, error) {
					return &somedownstream.SomeResp{
						Msg: fmt.Sprintf("got request in somedownstream: %s", req.Request.Msg),
					}, nil
				},
			}, nil, nil
		},
	)
}
