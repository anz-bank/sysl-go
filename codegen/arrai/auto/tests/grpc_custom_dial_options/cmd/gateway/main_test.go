package main

import (
	"context"
	"strconv"
	"testing"

	pb "grpc_custom_dial_options/internal/gen/pb/gateway"
	"grpc_custom_dial_options/internal/gen/pkg/servers/gateway"
	"grpc_custom_dial_options/internal/gen/pkg/servers/gateway/encoder_backend"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestGrpcCustomDialOptions(t *testing.T) {
	t.Parallel()

	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Encoder_backend.Rot13.
		ExpectRequest(&encoder_backend.EncodingRequest{Content: "hello world"}).
		Mock(func(t *testing.T, ctx context.Context, _ *encoder_backend.EncodingRequest) (*encoder_backend.EncodingResponse, error) {
			var md metadata.MD
			md, ok := metadata.FromIncomingContext(ctx)
			require.True(t, ok)

			values := md.Get("rot-parameter-override")
			require.Len(t, values, 1)
			tau, err := strconv.Atoi(values[0])
			require.NoError(t, err, "rot-parameter-override metadata had unexpected value: %v, expected an integer", values)
			require.Equal(t, 17, tau)

			return &encoder_backend.EncodingResponse{Content: "yvccf nficu"}, nil
		})

	gatewayTester.Encode().
		WithRequest(&pb.EncodeRequest{Content: "hello world", EncoderId: "rot13"}).
		ExpectResponse(&pb.EncodeResponse{Content: "yvccf nficu"}).
		Send()
}
