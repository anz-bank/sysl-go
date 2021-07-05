package main

import (
	"context"
	"os"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	"rest_post_urlencoded_form/internal/gen/pkg/servers/gateway"
)

type AppConfig struct{}

func PostPingBinary(_ context.Context, req *gateway.PostPingBinaryRequest) (*gateway.GatewayBinaryResponse, error) {
	return &gateway.GatewayBinaryResponse{
		Content: req.Request.Content,
	}, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
			return &gateway.ServiceInterface{
				PostPingBinary: PostPingBinary,
			}, nil, nil
		},
	)
}

func main() {
	ctx := log.PutLogger(context.Background(), log.NewDefaultLogger())

	handleError := func(err error) {
		if err != nil {
			log.Error(ctx, err, "something went wrong")
			os.Exit(1)
		}
	}

	srv, err := newAppServer(ctx)
	handleError(err)
	err = srv.Start()
	handleError(err)
}
