package main

import (
	"context"
	"fmt"
	"os"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/core"

	gateway "simple_rest_with_downstream/internal/gen/pkg/servers/gateway"
	encoder_backend "simple_rest_with_downstream/internal/gen/pkg/servers/gateway/encoder_backend"
)

type AppConfig struct{}

func PostEncodeEncoder_id(ctx context.Context, req *gateway.PostEncodeEncoder_idRequest, client gateway.PostEncodeEncoder_idClient) (*gateway.GatewayResponse, error) {
	if req.Encoder_id == "rot13" {

		encoderReq := &encoder_backend.PostRot13Request{
			Request: encoder_backend.EncodingRequest{
				Content: req.Request.Content,
			},
		}

		encoderResponse, err := client.Encoder_backendPostRot13(ctx, encoderReq)
		if err != nil {
			return nil, err
		}

		return &gateway.GatewayResponse{
			Content: encoderResponse.Content,
		}, nil

	} else {
		return nil, fmt.Errorf("encoder not supported")
	}
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx,
		func(ctx context.Context, config AppConfig) (*gateway.ServiceInterface, *core.Hooks, error) {
			return &gateway.ServiceInterface{
				PostEncodeEncoder_id: PostEncodeEncoder_id,
			}, nil, nil
		},
	)
}

func main() {
	// initialise context with pkg logger
	logger := log.NewStandardLogger()
	ctx := log.WithLogger(logger).WithConfigs(log.SetVerboseMode(true)).Onto(context.Background())

	handleError := func(err error) {
		if err != nil {
			log.Error(ctx, err)
			os.Exit(1)
		}
	}

	srv, err := newAppServer(ctx)
	handleError(err)
	err = srv.Start()
	handleError(err)
}
