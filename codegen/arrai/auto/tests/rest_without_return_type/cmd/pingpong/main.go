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

func GetPing0(_ context.Context, _ *pingpong.GetPing0Request) error {
	return nil
}

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

func GetPing3(_ context.Context, req *pingpong.GetPing3Request) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Identifier,
	}, nil
}

func GetPing4(_ context.Context, req *pingpong.GetPing4Request) (*pingpong.Pong, error) {
	return &pingpong.Pong{
		Identifier: req.Identifier,
	}, nil
}

func GetPing5(_ context.Context, req *pingpong.GetPing5Request) (*pingpong.Pong, *pingpong.Pong2, error) {
	if req.Identifier == 0 {
		return &pingpong.Pong{
			Identifier: req.Identifier,
		}, nil, nil
	}

	return nil, &pingpong.Pong2{
		ID: req.Identifier,
	}, nil
}

func GetPing6(_ context.Context, req *pingpong.GetPing6Request) (*pingpong.Pong, *pingpong.Pong2, error) {
	if req.Identifier == 0 {
		return &pingpong.Pong{
			Identifier: req.Identifier,
		}, nil, nil
	}

	return nil, &pingpong.Pong2{
		ID: req.Identifier,
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
			GetPing0:       GetPing0,
			GetPing1:       GetPing1,
			GetPing2:       GetPing2,
			GetPing3:       GetPing3,
			GetPing4:       GetPing4,
			GetPing5:       GetPing5,
			GetPing6:       GetPing6,
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
