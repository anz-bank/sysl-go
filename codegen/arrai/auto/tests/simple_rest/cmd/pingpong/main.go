package main

import (
	"context"

	pingpong "simple_rest/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
)

type AppConfig struct{}

func GetPing(ctx context.Context, req *pingpong.GetPingRequest, client pingpong.GetPingClient) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Identifier,
	}, nil
}

func application(ctx context.Context) {
	pingpong.Serve(ctx,
		func(ctx context.Context, config AppConfig) (*pingpong.ServiceInterface, *core.Hooks, error) {

			// FIXME auto codegen and common.MapError don't align.
			mapError := func(ctx context.Context, err error) *common.HTTPError {
				httpErr := common.MapError(ctx, err)
				return &httpErr
			}

			return &pingpong.ServiceInterface{
					GetPing: GetPing,
				}, &core.Hooks{
					MapError: mapError,
				},
				nil
		},
	)
}

func main() {
	// initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).Onto(context.Background())

	application(ctx)
}
