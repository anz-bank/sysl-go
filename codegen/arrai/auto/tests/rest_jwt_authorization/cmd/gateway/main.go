package main

import (
	"context"
	"os"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	gateway "rest_jwt_authorization/internal/gen/pkg/servers/gateway"
)

type AppConfig struct {
}

func Hello(ctx context.Context, req *gateway.PostHelloRequest) (*gateway.HelloResponse, error) {
	return &gateway.HelloResponse{
		Content: "why hello there",
	}, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx,
		func(ctx context.Context, cfg AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
			return &gateway.ServiceInterface{
					PostHello: Hello,
				},
				&core.Hooks{},
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
