package main

import (
	"context"
	"os"
	"time"

	"simple_rest/internal/gen/pkg/servers/pingpong"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct{}

func GetPing(_ context.Context, req *pingpong.GetPingRequest) (*pingpong.Pong, error) {
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

func GetPingoneof(_ context.Context, req *pingpong.GetGetoneofRequest) (*pingpong.OneOfResponse, error) {
	var ret pingpong.OneOfResponse
	if req.Identifier == 1 {
		ret = pingpong.OneOfResponse{OneOfResponseOne: &pingpong.OneOfResponseOne{1}}
	} else {
		ret = pingpong.OneOfResponse{OneOfResponseTwo: &pingpong.OneOfResponseTwo{"Two"}}
	}

	return &ret, nil
}

func createService(_ context.Context, _ AppConfig) (*pingpong.ServiceInterface, *core.Hooks, error) {
	return &pingpong.ServiceInterface{
			GetPing:        GetPing,
			GetPingtimeout: GetPingTimeout,
			GetGetoneof:    GetPingoneof,
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
