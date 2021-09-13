package main

import (
	"context"
	"os"

	"rest_env_config/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct {
	Identifier2 int64 `yaml:"id2" mapstructure:"id2"`
}

func (c *AppConfig) GetPing(_ context.Context, req *pingpong.GetPingRequest) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier:  req.Identifier,
		Identifier2: c.Identifier2,
	}, nil
}

func createService(_ context.Context, config AppConfig) (*pingpong.ServiceInterface, *core.Hooks, error) {
	return &pingpong.ServiceInterface{
			GetPing: config.GetPing,
		}, &core.Hooks{},
		nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return pingpong.NewServer(ctx, createService)
}

func main() {
	ctx := log.PutLogger(context.Background(), log.NewDefaultLogger())

	handleError := func(err error) {
		if err != nil {
			log.Error(ctx, err, "something goes wrong")
			os.Exit(1)
		}
	}

	srv, err := newAppServer(ctx)
	handleError(err)
	err = srv.Start()
	handleError(err)
}
