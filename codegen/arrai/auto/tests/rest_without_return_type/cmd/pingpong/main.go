package main

import (
	"context"
	"os"
	"time"

	"rest_without_return_type/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct{}

func GetPing1(_ context.Context, req *pingpong.GetPing1Request) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Identifier,
	}, nil
}

func GetPing2(_ context.Context, req *pingpong.GetPing2Request) (*pingpong.Pong, error) {
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

func createService(_ context.Context, _ AppConfig) (*pingpong.ServiceInterface, *core.Hooks, error) {
	return &pingpong.ServiceInterface{
			GetPing1:       GetPing1,
			GetPing2:       GetPing2,
			GetPingtimeout: GetPingTimeout,
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
