package main

import (
	"context"
	"os"

	"rest_with_validate/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct{}

func GetPing(_ context.Context, req *pingpong.GetPingRequest) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Identifier,
		Value:      req.Value,
	}, nil
}

func GetPingIgnore(_ context.Context, req *pingpong.GetPingIgnoreRequest) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Identifier,
		Value:      req.Value,
	}, nil
}

func PostPingPong(_ context.Context, req *pingpong.PostPingPongRequest) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Request.Identifier,
		Value:      req.Request.Value,
	}, nil
}

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
					GetPing:       GetPing,
					GetPingIgnore: GetPingIgnore,
					PostPingPong:  PostPingPong,
					PostPongPong:  PostPongPong,
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
