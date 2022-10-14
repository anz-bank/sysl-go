package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	"rest_miscellaneous/internal/gen/pkg/servers/gateway"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/encoder_backend"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/multi_contenttype_backend"
	"rest_miscellaneous/internal/gen/pkg/servers/gateway/oneof_backend"
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
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{One: &oneof_backend.One{One: v.One.One}})
		case v.Two != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{Two: &oneof_backend.Two{Two: v.Two.Two}})
		case v.Three != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{Three: &oneof_backend.Three{Three: v.Three.Three}})
		case v.EmptyType != nil:
			proor.Request.Values = append(proor.Request.Values, oneof_backend.OneOfRequest_values{EmptyType: &oneof_backend.EmptyType{}})
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
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{One: &gateway.One{One: v.One.One}})
		case v.Two != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{Two: &gateway.Two{Two: v.Two.Two}})
		case v.Three != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{Three: &gateway.Three{Three: v.Three.Three}})
		case v.EmptyType != nil:
			goor.Values = append(goor.Values, gateway.OneOfResponse_values{EmptyType: &gateway.EmptyType{}})
		}
	}

	return goor, nil
}

func GetPingMultiCode(_ context.Context, req *gateway.GetPingMultiCodeRequest) (*gateway.Pong, *gateway.PongString, error) {
	if req.Code == 0 {
		return &gateway.Pong{Identifier: 0}, nil, nil
	} else if req.Code == 1 {
		return nil, &gateway.PongString{S: "One"}, nil
	}

	return nil, nil, fmt.Errorf("code can only be 0 or 1")
}

func GetPingAsync(ctx context.Context, req *gateway.GetPingAsyncdownstreamsListRequest, client gateway.GetPingAsyncdownstreamsListClient) (*gateway.Pong, error) {
	backend1Req := &encoder_backend.GetPingListRequest{
		ID: req.ID,
	}
	backend2Req := &multi_contenttype_backend.PostPingMultiColonRequest{}

	ctx1 := common.RespHeaderAndStatusToContext(ctx, make(http.Header), 0)
	ctx2 := common.RespHeaderAndStatusToContext(ctx, make(http.Header), 0)

	backend1Future := common.Async(ctx1, func(ctxInt context.Context) (interface{}, error) {
		return client.Encoder_backendGetPingList(ctxInt, backend1Req)
	})
	backend2Future := common.Async(ctx2, func(ctxInt context.Context) (interface{}, error) {
		return client.Multi_contenttype_backendPostPingMultiColon(ctxInt, backend2Req)
	})

	// above 2 calls have been made asynchronously, lets get their results
	encoderResponseInterface, err := backend1Future.Get()
	if err != nil {
		return nil, err
	}

	// for this test I am ignoring the value of the second result, as long as it's not an error
	_, err = backend2Future.Get()
	if err != nil {
		return nil, err
	}

	encoderResponse := encoderResponseInterface.(*encoder_backend.Pong)

	return &gateway.Pong{
		Identifier: encoderResponse.Identifier,
	}, nil
}

func createService(_ context.Context, _ AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
	return &gateway.ServiceInterface{
		PostRotateOneOf:             PostRotateOneOf,
		GetPingIdList:               GetPingList,
		GetPingStringS:              GetPingString,
		PatchPing:                   PatchPing,
		PostPingBinary:              PostPingBinary,
		GetPingMultiCode:            GetPingMultiCode,
		GetPingAsyncdownstreamsList: GetPingAsync,
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
