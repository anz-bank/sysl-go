package main

import (
	"context"
	"fmt"
	"os"

	pb "grpc_custom_dial_options/internal/gen/pb/gateway"
	"grpc_custom_dial_options/internal/gen/pkg/servers/gateway"
	"grpc_custom_dial_options/internal/gen/pkg/servers/gateway/encoder_backend"

	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	}
	return nil, fmt.Errorf("custom response from app business logic: encoder not supported")
}

func makeCustomGrpcMetadataInjector(key, value string) grpc.UnaryClientInterceptor {
	f := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metadata.New(nil)
		md.Set(key, value)
		ctxPrime := metadata.NewOutgoingContext(ctx, md)
		return invoker(ctxPrime, method, req, reply, cc, opts...)
	}
	return f
}

func createService(_ context.Context, _ AppConfig) (*gateway.GrpcServiceInterface, *core.Hooks, error) {
	// Customise how we connect to the backend_encoder service with gRPC
	myCustomDialOpts := []grpc.DialOption{}
	f := makeCustomGrpcMetadataInjector("rot-parameter-override", "17")
	myCustomDialOpts = append(myCustomDialOpts, grpc.WithChainUnaryInterceptor(f))

	return &gateway.GrpcServiceInterface{
			Encode: Encode,
		}, &core.Hooks{
			AdditionalGrpcDialOptions: myCustomDialOpts,
		},
		nil
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
