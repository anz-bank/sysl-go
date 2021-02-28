package main

import (
	"context"
	"os"

	app "rest_admin_server/internal/gen/pkg/servers/app"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct{}

func GetHello(ctx context.Context, req *app.GetHelloListRequest) error {
	return nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return app.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*app.ServiceInterface, *core.Hooks, error) {
			return &app.ServiceInterface{
					GetHelloList: GetHello,
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
