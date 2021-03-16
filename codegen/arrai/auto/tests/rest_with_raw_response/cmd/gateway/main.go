package main

import (
	"context"
	"net/http"
	"os"

	"github.com/anz-bank/sysl-go/common"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"

	gateway "rest_with_raw_response/internal/gen/pkg/servers/gateway"
	encoder_backend "rest_with_raw_response/internal/gen/pkg/servers/gateway/encoder_backend"
)

type AppConfig struct{}

func PostPingBytes(ctx context.Context, req *gateway.PostReverseBytesNRequest, client gateway.PostReverseBytesNClient) ([]byte, error) {
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

func PostPingString(ctx context.Context, req *gateway.PostReverseStringNRequest, client gateway.PostReverseStringNClient) (string, error) {
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

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
			return &gateway.ServiceInterface{
				PostReverseBytesN:  PostPingBytes,
				PostReverseStringN: PostPingString,
			}, nil, nil
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
