package main

import (
	"context"
	"os"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	"rest_miscellaneous/internal/gen/pkg/servers/gatewayWithBff"
)

type AppConfig struct{}

func PostPingBinary(_ context.Context, req *gatewayWithBff.PostPingBinaryRequest) (*gatewayWithBff.GatewayBinaryResponse, error) {
	return &gatewayWithBff.GatewayBinaryResponse{
		Content: req.Request.Content,
	}, nil
}

func createService(_ context.Context, _ AppConfig) (*gatewayWithBff.ServiceInterface, *core.Hooks, error) {
	return &gatewayWithBff.ServiceInterface{
		PostPingBinary: PostPingBinary,
	}, nil, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gatewayWithBff.NewServer(ctx, createService)
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
