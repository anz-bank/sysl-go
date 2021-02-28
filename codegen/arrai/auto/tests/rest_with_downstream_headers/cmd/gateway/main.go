package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	gateway "rest_with_downstream_headers/internal/gen/pkg/servers/gateway"
	encoder_backend "rest_with_downstream_headers/internal/gen/pkg/servers/gateway/encoder_backend"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct{}

func PostEncodeEncoder_id(ctx context.Context, req *gateway.PostEncodeEncoder_idRequest, client gateway.PostEncodeEncoder_idClient) (*gateway.GatewayResponse, error) {
	if req.Encoder_id == "rot13" {

		encoderReq := &encoder_backend.PostRot13Request{
			Request: encoder_backend.EncodingRequest{
				Content: req.Request.Content,
			},
		}

		// Non-obvious: generated code assumes all HTTP headers will be carried about in the context.
		header := common.RequestHeaderFromContext(ctx)
		callHeader := make(http.Header)
		x := header.Get("x")
		if x == "" {
			x = "imputed-x-header-value"
		}
		callHeader.Add("x", x) // backend regards x as required header
		callHeader.Add("y", header.Get("y"))
		callHeader.Add("z", "custom-z-header-value") // backend regards z as required header
		callCtx := common.RequestHeaderToContext(ctx, callHeader)
		encoderResponse, err := client.Encoder_backendPostRot13(callCtx, encoderReq)
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
