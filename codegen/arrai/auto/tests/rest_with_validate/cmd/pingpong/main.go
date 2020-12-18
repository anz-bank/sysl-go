package main

import (
	"context"
	"os"

	pingpong "rest_with_validate/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/core"
)

type AppConfig struct{}

func GetPing(ctx context.Context, req *pingpong.GetPingRequest) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Identifier,
		Value:      req.Value,
	}, nil
}

func GetPingIgnore(ctx context.Context, req *pingpong.GetPingIgnoreRequest) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Identifier,
		Value:      req.Value,
	}, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return pingpong.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*pingpong.ServiceInterface, *core.Hooks, error) {
			return &pingpong.ServiceInterface{
					GetPing:       GetPing,
					GetPingIgnore: GetPingIgnore,
				}, &core.Hooks{},
				nil
		},
	)
}

func main() {
	// initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).WithConfigs(log.SetVerboseMode(true)).Onto(context.Background())

	handleError := func(err error) {
		if err != nil {
			log.Error(ctx, err)
			os.Exit(1)
		}
	}

	srv, err := newAppServer(ctx)
	handleError(err)
	err = srv.Start()
	handleError(err)
}
