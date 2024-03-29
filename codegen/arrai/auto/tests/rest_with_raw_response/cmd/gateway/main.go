package main

import (
	"context"
	"net/http"
	"os"

	"github.com/anz-bank/sysl-go/common"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	"rest_with_raw_response/internal/gen/pkg/servers/gateway"
	"rest_with_raw_response/internal/gen/pkg/servers/gateway/encoder_backend"
)

type AppConfig struct{}

func PostReverseBytesN(ctx context.Context, req *gateway.PostReverseBytesNRequest, client gateway.PostReverseBytesNClient) ([]byte, error) {
	result := req.Request

	for i := 0; i < int(req.Count); i++ {
		header := make(http.Header)
		header.Set("Content-Type", "application/octet-stream")
		ctx = common.RequestHeaderToContext(ctx, header)
		response, err := client.Encoder_backendPostReverseBytes(ctx, &encoder_backend.PostReverseBytesRequest{
			Request: req.Request,
		})
		if err != nil {
			return nil, err
		}
		result = response
	}

	return result, nil
}

func PostReverseStringN(ctx context.Context, req *gateway.PostReverseStringNRequest, client gateway.PostReverseStringNClient) (string, error) {
	result := req.Request

	for i := 0; i < int(req.Count); i++ {
		header := make(http.Header)
		header.Set("Content-Type", "text/plain")
		ctx = common.RequestHeaderToContext(ctx, header)
		response, err := client.Encoder_backendPostReverseString(ctx, &encoder_backend.PostReverseStringRequest{
			Request: req.Request,
		})
		if err != nil {
			return "", err
		}
		result = response
	}

	return result, nil
}

func PostPingStringAlias(ctx context.Context, req *gateway.PostPingStringAliasRequest, client gateway.PostPingStringAliasClient) (*gateway.PingStringResponse, error) {
	encoderReq := &encoder_backend.PostPingStringAliasRequest{
		Request: encoder_backend.PingStringRequest(req.Request),
	}

	header := make(http.Header)
	header.Set("Content-Type", "text/plain")
	ctx = common.RequestHeaderToContext(ctx, header)
	response, err := client.Encoder_backendPostPingStringAlias(ctx, encoderReq)
	if err != nil {
		return nil, err
	}

	resp := gateway.PingStringResponse(*response)
	return &resp, nil
}

func PostPingByteAlias(ctx context.Context, req *gateway.PostPingByteAliasRequest, client gateway.PostPingByteAliasClient) (*gateway.PingByteResponse, error) {
	encoderReq := &encoder_backend.PostPingByteAliasRequest{
		Request: encoder_backend.PingByteRequest(req.Request),
	}

	header := make(http.Header)
	header.Set("Content-Type", "application/octet-stream")
	ctx = common.RequestHeaderToContext(ctx, header)
	response, err := client.Encoder_backendPostPingByteAlias(ctx, encoderReq)
	if err != nil {
		return nil, err
	}

	resp := gateway.PingByteResponse(*response)
	return &resp, nil
}

func createService(_ context.Context, _ AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
	return &gateway.ServiceInterface{
		PostPingStringAlias: PostPingStringAlias,
		PostPingByteAlias:   PostPingByteAlias,
		PostReverseBytesN:   PostReverseBytesN,
		PostReverseStringN:  PostReverseStringN,
	}, nil, nil
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx, createService)
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
