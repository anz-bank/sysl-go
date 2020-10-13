package main

import (
	"context"
	"os"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"

	gateway "rest_jwt_authorization/gen/pkg/servers/Gateway"
)

type AppConfig struct {
}

func Hello(ctx context.Context, req *gateway.PostHelloRequest, client gateway.PostHelloClient) (*gateway.HelloResponse, error) {
	return &gateway.HelloResponse{
		Content: "why hello there",
	}, nil
}

func application(ctx context.Context) error {
	return gateway.Serve(ctx,
		func(ctx context.Context, cfg AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {

			// FIXME auto codegen and common.MapError don't align.
			mapError := func(ctx context.Context, err error) *common.HTTPError {
				httpErr := common.MapError(ctx, err)
				return &httpErr
			}

			return &gateway.ServiceInterface{
					PostHello: Hello,
				},
				&core.Hooks{
					MapError: mapError,
				},
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
