package main

import (
	"context"
	"fmt"
	"os"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/oneof_backend"

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

func PatchPing(_ context.Context, req *gateway.PatchPingRequest) (*gateway.GatewayPatchResponse, error) {
	return &gateway.GatewayPatchResponse{
		Content: req.Request.Content,
	}, nil
}

func PostRotateOneOf(ctx context.Context, req *gateway.PostRotateOneOfRequest, client gateway.PostRotateOneOfClient) (*gateway.OneOfResponse, error) {
	proor := &oneof_backend.PostRotateOneOfRequest{}
	for _, v := range req.Request.Values {
		switch {
		case v.One != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{One: &oneof_backend.One{v.One.One}})
		case v.Two != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{Two: &oneof_backend.Two{v.Two.Two}})
		case v.Three != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{Three: &oneof_backend.Three{v.Three.Three}})
		}
	}

	resp, err := client.Oneof_backendPostRotateOneOf(ctx, proor)
	if err != nil {
		return nil, err
	}

	goor := &gateway.OneOfResponse{}
	for _, v := range resp.Values {
		switch {
		case v.One != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{One: &gateway.One{v.One.One}})
		case v.Two != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{Two: &gateway.Two{v.Two.Two}})
		case v.Three != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{Three: &gateway.Three{v.Three.Three}})
		}
	}

	return goor, nil
}

func GetPingMultiCode(_ context.Context, req *gateway.GetPingMultiCodeRequest) (*gateway.Pong, *gateway.PongString, error) {
	if req.Code == 0 {
		return &gateway.Pong{0}, nil, nil
	} else if req.Code == 1 {
		return nil, &gateway.PongString{"One"}, nil
	}

	return nil, nil, fmt.Errorf("Code can only be 0 or 1")
}

func createService(_ context.Context, _ AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
	return &gateway.ServiceInterface{
		PostRotateOneOf:  PostRotateOneOf,
		GetPingIdList:    GetPingList,
		GetPingStringS:   GetPingString,
		PatchPing:        PatchPing,
		PostPingBinary:   PostPingBinary,
		GetPingMultiCode: GetPingMultiCode,
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
