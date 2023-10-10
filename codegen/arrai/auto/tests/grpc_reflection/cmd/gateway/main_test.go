package main

import (
	"context"
	"fmt"

	reflection "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	"testing"
	"time"

	"google.golang.org/grpc"

	pb "grpc_reflection/internal/gen/pb/gateway"

	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
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
`

const applicationConfigReflection = `---
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9021 # FIXME no guarantee this port is free
      enableReflection: true
  downstream:
    contextTimeout: "30s"
`

func doGatewayEncode(ctx context.Context, content string) (string, error) {
	conn, err := grpc.Dial("localhost:9021", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("test client failed to connect to gateway: %s\n", err.Error())
		return "", err
	}
	defer conn.Close()

	client := pb.NewGatewayClient(conn)
	response, err := client.Encode(ctx, &pb.EncodeReq{Content: content, EncoderId: "rot13"})
	if err != nil {
		fmt.Printf("test client got error after making Encoding request to gateway: %s\n", err.Error())
		return "", err
	}
	return response.Content, nil
}

func doGatewayReflectionList(ctx context.Context) ([]string, error) {
	conn, err := grpc.Dial("localhost:9021", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("test client failed to connect to gateway: %s\n", err.Error())
		return nil, err
	}
	defer conn.Close()

	client := reflection.NewServerReflectionClient(conn)
	reflectionInfo, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		fmt.Printf("error retrieving the reflection info: %s\n", err.Error())
		return nil, err
	}
	err = reflectionInfo.Send(&reflection.ServerReflectionRequest{
		MessageRequest: &reflection.ServerReflectionRequest_ListServices{},
	})
	if err != nil {
		fmt.Printf("error sending reflection request: %s\n", err.Error())
		return nil, err
	}
	recv, err := reflectionInfo.Recv()
	if err != nil {
		fmt.Printf("error retrieving reflection request: %s\n", err.Error())
		return nil, err
	}

	list, ok := recv.MessageResponse.(*reflection.ServerReflectionResponse_ListServicesResponse)
	if !ok {
		fmt.Printf("error casting list response\n")
		return nil, err
	}

	var names []string
	for _, service := range list.ListServicesResponse.Service {
		names = append(names, service.Name)
	}
	return names, nil
}

func TestGRPCServerReflectionDisabled(t *testing.T) {
	testGRPCServerReflection(t, false)
}

func TestGRPCServerReflectionEnabled(t *testing.T) {
	testGRPCServerReflection(t, true)
}

func testGRPCServerReflection(t *testing.T, enabled bool) {
	// Set an appropriate application config based on the enabled state
	config := applicationConfig
	if enabled {
		config = applicationConfigReflection
	}
	ctx := core.WithConfigFile(context.Background(), []byte(config))

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
	backoff, err := retry.NewFibonacci(20 * time.Millisecond)
	require.Nil(t, err)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, err := doGatewayEncode(ctx, "testing; one two, one two; is this thing on?")
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Verify that the reflection service is appropriately set
	names, err := doGatewayReflectionList(ctx)
	if enabled {
		require.Nil(t, err)
		require.Equal(t, []string{"gateway.Gateway", "grpc.reflection.v1.ServerReflection", "grpc.reflection.v1alpha.ServerReflection"}, names)
	} else {
		require.NotNil(t, err)
	}
}
