package main

import (
	"context"
	"os"
	"time"

	pingpong "simple_rest/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/core"
)

type AppConfig struct{}

func GetPing(_ context.Context, req *pingpong.GetPingRequest) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Identifier,
	}, nil
}

func GetPingTimeout(_ context.Context, req *pingpong.GetPingtimeoutRequest) (*pingpong.Pong, error) {
	time.Sleep(10 * time.Second)
	return &pingpong.Pong{
		Identifier: req.Identifier,
	}, nil
}

func GetPingoneof(_ context.Context, req *pingpong.GetGetoneofRequest) (*pingpong.OneOfResponse, error) {
	var ret pingpong.OneOfResponse
	if req.Identifier == 1 {
		ret = pingpong.OneOfResponseOne{1}
	} else {
		ret = pingpong.OneOfResponseTwo{"Two"}
	}

	return &ret, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return pingpong.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*pingpong.ServiceInterface, *core.Hooks, error) {
			return &pingpong.ServiceInterface{
					GetPing:        GetPing,
					GetPingtimeout: GetPingTimeout,
					GetGetoneof:    GetPingoneof,
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
