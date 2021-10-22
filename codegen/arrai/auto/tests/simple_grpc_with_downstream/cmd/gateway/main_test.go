package main

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "simple_grpc_with_downstream/internal/gen/pb/gateway"
	"simple_grpc_with_downstream/internal/gen/pkg/servers/gateway"
	"simple_grpc_with_downstream/internal/gen/pkg/servers/gateway/encoder_backend"
)

func TestSimpleGRPCWithDownstream(t *testing.T) {
	t.Parallel()

	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Encoder_backend.Rot13.
		ExpectRequest(&encoder_backend.EncodingRequest{Content: "hello world"}).
		MockResponse(&encoder_backend.EncodingResponse{Content: "uryyb jbeyq"})

	gatewayTester.Mocks.Encoder_backend.Rot13.
		ExpectRequest(&encoder_backend.EncodingRequest{Content: "hello world!"}).
		MockResponse(&encoder_backend.EncodingResponse{Content: "uryyb jbeyq!"})

	gatewayTester.Encode().
		WithRequest(&pb.EncodeRequest{Content: "hello world", EncoderId: "rot13"}).
		ExpectResponse(&pb.EncodeResponse{Content: "uryyb jbeyq"}).
		Send()

	gatewayTester.Encode().
		WithRequest(&pb.EncodeRequest{Content: "hello world!", EncoderId: "rot13"}).
		ExpectResponse(&pb.EncodeResponse{Content: "uryyb jbeyq!"}).
		Send()
}

func TestSimpleGRPCWithDownstream_Fail(t *testing.T) {
	t.Parallel()

	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Encoder_backend.Rot13.
		ExpectRequest(&encoder_backend.EncodingRequest{Content: "hello world"}).
		MockError(status.Error(codes.Unknown, "Failed!"))

	gatewayTester.Encode().
		WithRequest(&pb.EncodeRequest{Content: "hello world", EncoderId: "rot13"}).
		ExpectError(status.Error(codes.Unknown, "Failed!")).
		Send()
}
