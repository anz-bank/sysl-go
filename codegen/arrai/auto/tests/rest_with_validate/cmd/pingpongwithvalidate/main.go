package main

import (
	"context"
	"os"

	"rest_with_validate/internal/gen/pkg/servers/pingpongwithvalidate"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct{}

func PostPingWithValidate(_ context.Context, _ *pingpongwithvalidate.PostPingWithValidateRequest) error {
	return nil
}

func PostPongPong(_ context.Context, req *pingpongwithvalidate.PostPongPongRequest) (*pingpongwithvalidate.Pong, error) {
	return &pingpongwithvalidate.Pong{
		Identifier: req.Request.Identifier,
		Value:      req.Request.Value,
	}, nil
}

func createService(_ context.Context, _ AppConfig) (*pingpongwithvalidate.ServiceInterface, *core.Hooks, error) {
	return &pingpongwithvalidate.ServiceInterface{
			PostPingWithValidate: PostPingWithValidate,
			PostPongPong:         PostPongPong,
		}, &core.Hooks{},
		nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return pingpongwithvalidate.NewServer(ctx, createService)
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
