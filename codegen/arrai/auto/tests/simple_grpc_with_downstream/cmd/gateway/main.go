package main

import (
	"context"
	"fmt"
	"os"

	pb "simple_grpc_with_downstream/internal/gen/pb/gateway"
	gateway "simple_grpc_with_downstream/internal/gen/pkg/servers/gateway"
	encoder_backend "simple_grpc_with_downstream/internal/gen/pkg/servers/gateway/encoder_backend"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct{}

func Encode(ctx context.Context, req *pb.EncodeRequest, client gateway.EncodeClient) (*pb.EncodeResponse, error) {
	if req.EncoderId == "rot13" {
		encoderReq := &encoder_backend.EncodingRequest{
			Content: req.Content,
		}

		encoderResponse, err := client.Encoder_backendRot13(ctx, encoderReq)
		if err != nil {
			return nil, err
		}

		return &pb.EncodeResponse{
			Content: encoderResponse.Content,
		}, nil

	} else {
		return nil, fmt.Errorf("custom response from app business logic: encoder not supported")
	}
}

func newAppServer(ctx context.Context) (core.StoppableServer, error) {
	return gateway.NewServer(ctx,
		func(ctx context.Context, cfg AppConfig) (*gateway.GrpcServiceInterface, *core.Hooks, error) {
			return &gateway.GrpcServiceInterface{
				Encode: Encode,
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
