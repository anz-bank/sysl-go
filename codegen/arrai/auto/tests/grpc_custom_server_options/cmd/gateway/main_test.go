package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"

	pb "grpc_custom_server_options/internal/gen/pb/gateway"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v2"
)

const applicationConfigNoCustomServerOptions = `---
app:
  setAdditionalGrpcServerOptions: false
  setOverrideGrpcServerOptions: false
  customMetadata:
    - key: coffee
      value: freshly ground
    - key: weather
      value: windy
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`

const applicationConfigAdditionalCustomServerOptions = `---
app:
  setAdditionalGrpcServerOptions: true
  setOverrideGrpcServerOptions: false
  customMetadata:
    - key: coffee
      value: freshly ground
    - key: weather
      value: windy
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9022 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`

const applicationConfigOverrideServerOptions = `---
app:
  setAdditionalGrpcServerOptions: false
  setOverrideGrpcServerOptions: true
  customMetadata:
    - key: coffee
      value: freshly ground
    - key: weather
      value: windy
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9023 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`

func getServerAddr(appCfg []byte) (string, error) {
	cfg := config.DefaultConfig{}
	err := yaml.Unmarshal(appCfg, &cfg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", cfg.GenCode.Upstream.GRPC.HostName, cfg.GenCode.Upstream.GRPC.Port), nil
}

func doGatewayRequestResponse(ctx context.Context, addr string, content string) (string, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("test client failed to connect to gateway: %s\n", err.Error())
		return "", err
	}
	defer conn.Close()

	client := pb.NewGatewayClient(conn)
	response, err := client.Hello(ctx, &pb.HelloRequest{Content: content})
	if err != nil {
		fmt.Printf("test client got error after making Hello request to gateway: %s\n", err.Error())
		return "", err
	}
	return response.Content, nil
}

func TestCustomisationOfServerOptions(t *testing.T) {
	type Scenario struct {
		name                      string
		appCfg                    []byte
		expectedResponseFragments []string
	}

	scenarios := []Scenario{
		{
			name:                      "default",
			appCfg:                    []byte(applicationConfigNoCustomServerOptions),
			expectedResponseFragments: []string{"echo"},
		},
		{
			name:                      "additional-options",
			appCfg:                    []byte(applicationConfigAdditionalCustomServerOptions),
			expectedResponseFragments: []string{"coffee:[freshly ground]", "weather:[windy]"},
		},
		{
			name:                      "override",
			appCfg:                    []byte(applicationConfigOverrideServerOptions),
			expectedResponseFragments: []string{"coffee:[freshly ground]", "weather:[windy]"},
		},
	}

	for i := range scenarios {
		scenario := scenarios[i]
		t.Run(scenario.name, func(t *testing.T) {
			// Figure out what address our server will listening on
			serverAddr, err := getServerAddr(scenario.appCfg)
			require.NoError(t, err)

			// Initialise context with pkg logger
			logger := log.NewStandardLogger()
			ctx := log.WithLogger(logger).Onto(context.Background())

			// Add in a fake filesystem to pass in config
			memFs := afero.NewMemMapFs()
			err = afero.Afero{Fs: memFs}.WriteFile("config.yaml", scenario.appCfg, 0777)
			require.NoError(t, err)
			ctx = core.ConfigFileSystemOnto(ctx, memFs)

			// FIXME patch core.Serve to allow it to optionally load app config path from ctx
			args := os.Args
			defer func() { os.Args = args }()
			os.Args = []string{"./gateway.out", "config.yaml"}

			// Start gateway application running as server
			go application(ctx)

			// Wait for application to come up
			backoff, err := retry.NewFibonacci(20 * time.Millisecond)
			require.Nil(t, err)
			backoff = retry.WithMaxDuration(5*time.Second, backoff)
			err = retry.Do(ctx, backoff, func(ctx context.Context) error {
				_, err := doGatewayRequestResponse(ctx, serverAddr, "testing; one two, one two; is this thing on?")
				if err != nil {
					return retry.RetryableError(err)
				}
				return nil
			})
			require.NoError(t, err)

			// Test if the endpoint of our gateway application server works
			actual, err := doGatewayRequestResponse(ctx, serverAddr, "echo")
			require.NoError(t, err)
			for _, expectedFragment := range scenario.expectedResponseFragments {
				require.Contains(t, actual, expectedFragment)
			}
			// FIXME how do we stop the application server?
		})
	}
}
