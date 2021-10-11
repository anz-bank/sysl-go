package main

import (
	"context"
	"os"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	"rest_miscellaneous/internal/gen/pkg/servers/gateway"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/encoder_backend"
)

type AppConfig struct{}

func GetPingList(ctx context.Context, req *gateway.GetPingIdListRequest, client gateway.GetPingIdListClient) (*gateway.Pong, error) {
	backendReq := &encoder_backend.GetPingListRequest{
		ID: req.ID,
	}

	encoderResponse, err := client.Encoder_backendGetPingList(ctx, backendReq)
	if err != nil {
		return nil, err
	}

	return &gateway.Pong{
		Identifier: encoderResponse.Identifier,
	}, nil
}

func GetPingString(ctx context.Context, req *gateway.GetPingStringSRequest, client gateway.GetPingStringSClient) (*gateway.PongString, error) {
	backendReq := &encoder_backend.GetPingStringSRequest{
		S: req.S,
	}

	encoderResponse, err := client.Encoder_backendGetPingStringS(ctx, backendReq)
	if err != nil {
		return nil, err
	}

	return &gateway.PongString{
		S: encoderResponse.S,
	}, nil
}

func PostPingBinary(_ context.Context, req *gateway.PostPingBinaryRequest) (*gateway.GatewayBinaryResponse, error) {
	return &gateway.GatewayBinaryResponse{
		Content: req.Request.Content,
	}, nil
}

func createService(_ context.Context, _ AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
	return &gateway.ServiceInterface{
		GetPingIdList:  GetPingList,
		GetPingStringS: GetPingString,
		PostPingBinary: PostPingBinary,
	}, nil, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx, createService)
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
