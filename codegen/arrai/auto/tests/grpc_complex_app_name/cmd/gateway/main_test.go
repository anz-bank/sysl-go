package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
	"unicode"

	"grpc_complex_app_name/internal/gen/pkg/servers/gateway"
	"grpc_complex_app_name/internal/gen/pkg/servers/gateway/encoder_backend"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	ebpb "grpc_complex_app_name/internal/gen/pb/encoder_backend"
	pb "grpc_complex_app_name/internal/gen/pb/gateway"

	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const applicationConfig = `---
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
    encoder_backend:
      serviceAddress: localhost:9022
`

func doGatewayRequestResponse(ctx context.Context, content string) (string, error) {
	conn, err := grpc.Dial("localhost:9021", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("test client failed to connect to gateway: %s\n", err.Error())
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewGatewayClient(conn)
	response, err := client.Encode(ctx, &pb.EncodeReq{Content: content, EncoderId: "rot13"})
	if err != nil {
		fmt.Printf("test client got error after making Encoding request to gateway: %s\n", err.Error())
		return "", err
	}
	return response.Content, nil
}

type dummyEncoderBackend struct {
	ebpb.UnimplementedEncoderBackendServer
}

func (s dummyEncoderBackend) Rot13(_ context.Context, req *ebpb.EncodingRequest) (*ebpb.EncodingResponse, error) {
	// valuable business logic as used in our dummy implementation of EncoderBackend service
	toRot13 := make(map[rune]rune)
	az := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
	for i, r := range az {
		toRot13[r] = az[(i+13)%len(az)]
	}
	rot13 := func(s string) string {
		var b strings.Builder
		for _, r := range s {
			s, ok := toRot13[unicode.ToLower(r)]
			if ok {
				b.WriteRune(s)
			} else {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	return &ebpb.EncodingResponse{Content: rot13(req.Content)}, nil
}

func grpcLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		fmt.Printf("dummy encoder backend: Call %q OK", info.FullMethod)
	} else {
		fmt.Printf("dummy encoder backend: Call %q FAIL: %s", info.FullMethod, err)
	}
	return resp, err
}

func startDummyEncoderBackendServer(addr string) (stopServer func() error) {
	// Starts a hand-written implementation of the EncoderBackend service running on given TCP Address.
	// Returns a function that can be used to stop the server.

	server := grpc.NewServer(grpc.UnaryInterceptor(grpcLogger))
	ebpb.RegisterEncoderBackendServer(server, dummyEncoderBackend{})

	c := make(chan error, 1)

	go func() {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			c <- err
			return
		}
		fmt.Printf("dummy encoder backend server will listen on %s...\n", addr)
		c <- server.Serve(listener)
	}()

	stopServer = func() error {
		// If the server stopped with some error before the caller
		// tried to stop it, return that error instead.
		select {
		case err := <-c:
			return err
		default:
		}
		server.Stop()
		return nil
	}
	return stopServer
}

func TestGRPCComplexAppNameSmokeTest(t *testing.T) {
	// Override sysl-go app command line interface to directly pass in app config
	ctx := core.WithConfigFile(context.Background(), []byte(applicationConfig))

	// Start the dummy encoder backend service running
	stopEncoderBackendServer := startDummyEncoderBackendServer("localhost:9022")
	defer func() {
		err := stopEncoderBackendServer()
		require.NoError(t, err)
	}()

	appServer, err := newAppServer(ctx)
	require.NoError(t, err)
	defer func() {
		err := appServer.Stop()
		if err != nil {
			panic(err)
		}
	}()

	// Start application server
	go func() {
		err := appServer.Start()
		if err != nil {
			panic(err)
		}
	}()

	// Wait for application to come up
	backoff := retry.NewFibonacci(20 * time.Millisecond)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := doGatewayRequestResponse(ctx, "testing; one two, one two; is this thing on?")
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test if the endpoint of our gateway application server works
	expected := "uryyb jbeyq"
	actual, err := doGatewayRequestResponse(ctx, "hello world")
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}

func TestGRPCComplexAppName_Mocked(t *testing.T) {
	gatewayTester := gateway.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	gatewayTester.Mocks.Encoder_backend.Rot13.
		ExpectRequest(&encoder_backend.EncodingRequest{Content: "hello world"}).
		MockResponse(&encoder_backend.EncodingResponse{Content: "uryyb jbeyq"})

	gatewayTester.Mocks.Encoder_backend.Rot13.
		ExpectRequest(&encoder_backend.EncodingRequest{Content: "hello world!"}).
		MockResponse(&encoder_backend.EncodingResponse{Content: "uryyb jbeyq!"})

	gatewayTester.Encode().
		WithRequest(&pb.EncodeReq{Content: "hello world", EncoderId: "rot13"}).
		ExpectResponse(&pb.EncodeResp{Content: "uryyb jbeyq"}).
		Send()

	gatewayTester.Encode().
		WithRequest(&pb.EncodeReq{Content: "hello world!", EncoderId: "rot13"}).
		ExpectResponse(&pb.EncodeResp{Content: "uryyb jbeyq!"}).
		Send()
}

func TestGRPCComplexAppName_Integration(t *testing.T) {
	// Start the dummy encoder backend service running
	stopEncoderBackendServer := startDummyEncoderBackendServer("localhost:9022")
	defer func() {
		err := stopEncoderBackendServer()
		require.NoError(t, err)
	}()

	gatewayTester := gateway.NewIntegrationTestServer(t, context.Background(), createService, applicationConfig)
	defer gatewayTester.Close()

	gatewayTester.Encode().
		WithRequest(&pb.EncodeReq{Content: "hello world", EncoderId: "rot13"}).
		ExpectResponse(&pb.EncodeResp{Content: "uryyb jbeyq"}).
		Send()

	gatewayTester.Encode().
		WithRequest(&pb.EncodeReq{Content: "hello world!", EncoderId: "rot13"}).
		ExpectResponse(&pb.EncodeResp{Content: "uryyb jbeyq!"}).
		Send()
}

func TestGRPCComplexAppName_Health(t *testing.T) {
	gatewayTester := gateway.NewIntegrationTestServer(t, context.Background(), createService, applicationConfig)
	defer gatewayTester.Close()

	// FIXME: this should be part of integration test server API. This is copied from how the integration test server's
	// internal client dials the server.
	conn, err := grpc.Dial(
		"testGateway",
		grpc.WithContextDialer(gatewayTester.GetE2eTester().GetBufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{
		Service: "aaaaaa",
	})
	require.NoError(t, err, fmt.Sprintf("healthcheck Check() should return no error but got %s", err))
	assert.Equal(t, resp.Status, grpc_health_v1.HealthCheckResponse_SERVING)

	resp, err = client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: "are you alive"})
	require.NoError(t, err, fmt.Sprintf("healthcheck Check() should return no error but got %s", err))
	assert.Equal(t, resp.Status, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
}
