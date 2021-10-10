package main

import (
	"context"
	"fmt"
	"os"

	pb "grpc_complex_app_name/internal/gen/pb/gateway"
	"grpc_complex_app_name/internal/gen/pkg/servers/gateway"
	"grpc_complex_app_name/internal/gen/pkg/servers/gateway/encoder_backend"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
)

type AppConfig struct{}

func Encode(ctx context.Context, req *pb.EncodeReq, client gateway.EncodeClient) (*pb.EncodeResp, error) {
	if req.EncoderId == "rot13" {
		encoderReq := &encoder_backend.EncodingRequest{
			Content: req.Content,
		}

		encoderResponse, err := client.Encoder_backendRot13(ctx, encoderReq)
		if err != nil {
			return nil, err
		}

		return &pb.EncodeResp{
			Content: encoderResponse.Content,
		}, nil

	} else {
		return nil, fmt.Errorf("custom response from app business logic: encoder not supported")
	}
}

func createService(_ context.Context, _ AppConfig) (*gateway.GrpcServiceInterface, *core.Hooks, error) {
	return &gateway.GrpcServiceInterface{
		Encode: Encode,
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
