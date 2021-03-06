package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"

	pb "grpc_custom_server_options/internal/gen/pb/gateway"

	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
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
	memFs := afero.NewMemMapFs()
	err := afero.Afero{Fs: memFs}.WriteFile("config.yaml", appCfg, 0777)
	if err != nil {
		return "", err
	}
	b := config.NewConfigReaderBuilder().WithFs(memFs).WithConfigFile("config.yaml")

	err = b.Build().Unmarshal(&cfg)
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

			// Override sysl-go app command line interface to directly pass in app config
			ctx := core.WithConfigFile(context.Background(), scenario.appCfg)

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
		})
	}
}
