package main

import (
	"context"
	"os"

	"rest_with_validate/internal/gen/pkg/servers/pingpongwithvalidate"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct{}

func PostPongPong(_ context.Context, req *pingpong.PostPongPongRequest) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Request.Identifier,
		Value:      req.Request.Value,
	}, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return pingpong.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*pingpong.ServiceInterface, *core.Hooks, error) {
			return &pingpong.ServiceInterface{
					PostPongPong: PostPongPong,
				}, &core.Hooks{},
				nil
		},
	)
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
