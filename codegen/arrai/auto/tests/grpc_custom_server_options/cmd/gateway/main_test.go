package main

import (
	"context"
	"testing"

	"grpc_custom_server_options/internal/gen/pkg/servers/gateway"

	pb "grpc_custom_server_options/internal/gen/pb/gateway"

	"github.com/stretchr/testify/require"
)

func TestGrpcCustomServerOptions(t *testing.T) {
	customMetadata := []KVPair{
		{"coffee", "freshly ground"},
		{"weather", "windy"},
	}

	type Scenario struct {
		name                           string
		setAdditionalGrpcServerOptions bool
		setOverrideGrpcServerOptions   bool
		expectedResponseFragments      []string
	}

	scenarios := []Scenario{
		{
			name:                           "default",
			setAdditionalGrpcServerOptions: false,
			setOverrideGrpcServerOptions:   false,
			expectedResponseFragments:      []string{"echo"},
		},
		{
			name:                           "additional-options",
			setAdditionalGrpcServerOptions: true,
			setOverrideGrpcServerOptions:   false,
			expectedResponseFragments:      []string{"echo", "coffee:[freshly ground]", "weather:[windy]"},
		},
		{
			name:                           "override",
			setAdditionalGrpcServerOptions: false,
			setOverrideGrpcServerOptions:   true,
			expectedResponseFragments:      []string{"echo", "coffee:[freshly ground]", "weather:[windy]"},
		},
	}

	for i := range scenarios {
		scenario := scenarios[i]
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()

			gatewayTester := gateway.NewTestServer(t, context.Background(), createService,
				AppConfig{
					CustomMetadata:                 customMetadata,
					SetAdditionalGrpcServerOptions: scenario.setAdditionalGrpcServerOptions,
					SetOverrideGrpcServerOptions:   scenario.setOverrideGrpcServerOptions,
				},
			)
			defer gatewayTester.Close()

			gatewayTester.Hello().
				WithRequest(&pb.HelloRequest{Content: "echo"}).
				TestResponse(func(t *testing.T, actual *pb.HelloResponse, err error) {
					require.NoError(t, err)
					for _, expectedFragment := range scenario.expectedResponseFragments {
						require.Contains(t, actual.Content, expectedFragment)
					}
				}).
				Send()
		})
	}
}
