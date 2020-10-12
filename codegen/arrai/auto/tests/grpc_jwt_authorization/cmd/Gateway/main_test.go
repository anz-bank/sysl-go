package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "grpc_jwt_authorization/gen/pb/gateway"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/jwtauth/jwttest"

	"github.com/sethvargo/go-retry"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v2"
)

var appCfgOne = []byte(`---
app:
library:
  log:
    format: text
    level: debug
  authentication:
    jwtauth:
      issuers:
        - name: "izzy-the-sysl-go-test-issuer"
          jwksUrl: http://localhost:9029/.well-known/jwks.json
          cacheTTL: 1m
          cacheRefresh: 1m
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`)

var appCfgTwo = []byte(`---
app:
library:
  authentication:
    jwtauth:
      issuers:
        - name: "izzy-the-sysl-go-test-issuer"
          jwksUrl: http://localhost:9029/.well-known/jwks.json
          cacheTTL: 1m
          cacheRefresh: 1m
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9022 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`)

var appCfgThree = []byte(`---
app:
library:
  authentication:
    jwtauth:
      issuers:
        - name: "izzy-the-sysl-go-test-issuer"
          jwksUrl: http://localhost:9029/.well-known/jwks.json
          cacheTTL: 1m
          cacheRefresh: 1m
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9023 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`)

var appCfgFour = []byte(`---
app:
library:
  authentication:
    jwtauth:
      issuers:
        - name: "izzy-the-sysl-go-test-issuer"
          jwksUrl: http://localhost:9029/.well-known/jwks.json
          cacheTTL: 1m
          cacheRefresh: 1m
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9024 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`)

var appCfgFive = []byte(`---
app:
library:
  authentication:
    jwtauth:
      issuers:
        - name: "izzy-the-sysl-go-test-issuer"
          jwksUrl: http://localhost:9029/.well-known/jwks.json
          cacheTTL: 1m
          cacheRefresh: 1m
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9025 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`)

var appCfgSix = []byte(`---
app:
development:
  disableAllAuthorizationRules: true
genCode:
  upstream:
    grpc:
      hostName: "localhost"
      port: 9026 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`)

func getServerAddr(appCfg []byte) (string, error) {
	cfg := config.DefaultConfig{}
	err := yaml.Unmarshal(appCfg, &cfg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", cfg.GenCode.Upstream.GRPC.HostName, cfg.GenCode.Upstream.GRPC.Port), nil
}

func doGatewayRequestResponse(ctx context.Context, addr string, rawJWT string) (string, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("test client failed to connect to gateway: %s\n", err.Error())
		return "", err
	}
	defer conn.Close()

	ctx = metadata.AppendToOutgoingContext(ctx, "Authorization", "bearer "+rawJWT)

	client := pb.NewGatewayClient(conn)
	response, err := client.Hello(ctx, &pb.HelloRequest{})
	if err != nil {
		fmt.Printf("test client got error after making Hello request to gateway: %s\n", err.Error())
		return "", err
	}
	return response.Content, nil
}

// implementation of a JWT issuer jkws endpoint for the application to trust
func serveIssuerJKWS(addr string, issuer jwttest.Issuer) (stopServer func() error) {

	// define and start http server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/jwks.json", issuer.ServeHTTP)
	server := &http.Server{Addr: addr, Handler: mux}

	c := make(chan error, 1)

	go func() {
		c <- server.ListenAndServe()
	}()

	stopServer = func() error {
		// If the server stopped with some error before the caller
		// tried to stop it, return that error instead.
		select {
		case err := <-c:
			return err
		default:
		}
		return server.Close()
	}
	return stopServer
}

func TestJWTAuthorizationOfGRPCEndpoints(t *testing.T) {

	keySize := 2048
	trustedIssuer, err := jwttest.NewIssuer("izzy-the-sysl-go-test-issuer", keySize)
	require.NoError(t, err)

	// The untrusted issuer is similar to the trusted one, with has a distinct
	// public-private keypair to the keypair used by the above issuer. Our
	// sysl-go generated application is NOT configured to trust this untrusted issuer.
	untrustedIssuer, err := jwttest.NewIssuer("untrusted-sysl-go-test-issuer", keySize)
	require.NoError(t, err)

	mustMakeTestJWT := func(i jwttest.Issuer, claims map[string]interface{}) string {
		rawToken, err := i.IssueFromMap(claims)
		require.NoError(t, err)
		return rawToken
	}

	stopIssuerServer := serveIssuerJKWS("localhost:9029", trustedIssuer)
	defer func() {
		err := stopIssuerServer()
		if err != nil {
			panic(fmt.Sprintf("issuer server died with error: %s", err))
		}
	}()

	type Scenario struct {
		name                      string
		appCfg                    []byte
		rawJWT                    string
		expectedResponseFragments []string
		expectedError             string
	}

	scenarios := []Scenario{
		{
			name:                      "request with authorised claims signed by trusted issuer succeeds",
			appCfg:                    appCfgOne,
			rawJWT:                    mustMakeTestJWT(trustedIssuer, map[string]interface{}{"scope": "hello"}),
			expectedResponseFragments: []string{"why hello there"},
		},
		{
			name:          "request with unauthorised claims signed by trusted issuer fails",
			appCfg:        appCfgTwo,
			rawJWT:        mustMakeTestJWT(trustedIssuer, map[string]interface{}{"scope": "banana"}),
			expectedError: "rpc error: code = PermissionDenied desc = insufficient permissions",
		},
		{
			name:                      "request with authorised claims and extra claims signed by trusted issuer succeeds",
			appCfg:                    appCfgThree,
			rawJWT:                    mustMakeTestJWT(trustedIssuer, map[string]interface{}{"scope": "banana hello"}),
			expectedResponseFragments: []string{"why hello there"},
		},
		{
			name:          "request with gibberish instead of JWT fails",
			appCfg:        appCfgFour,
			rawJWT:        "surprise!",
			expectedError: "rpc error: code = Unknown desc = jwtauth err 1: jwt parse error: square/go-jose: compact JWS format must have three parts", // FIXME impl detail leak in error msg
		},
		{
			name:          "request with authorised claims and signed by untrusted issuer fails",
			appCfg:        appCfgFive,
			rawJWT:        mustMakeTestJWT(untrustedIssuer, map[string]interface{}{"scope": "hello"}),
			expectedError: "rpc error: code = Unknown desc = jwtauth err 2: issuer not registered: untrusted-sysl-go-test-issuer",
		},
		{
			name:                      "request with authorised claims and signed by untrusted issuer succeeds if app configured to disable auth rules",
			appCfg:                    appCfgSix,
			rawJWT:                    mustMakeTestJWT(untrustedIssuer, map[string]interface{}{"scope": "hello"}),
			expectedResponseFragments: []string{"why hello there"},
			expectedError:             "",
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
			go func() {
				err := application(ctx)
				if err != nil {
					panic(err)
				}
			}()

			isResponseExpected := func(response string, err error) bool {
				if len(scenario.expectedError) > 0 {
					return err != nil && err.Error() == scenario.expectedError
				}
				if err != nil {
					return false
				}
				for _, expectedFragment := range scenario.expectedResponseFragments {
					if !strings.Contains(response, expectedFragment) {
						return false
					}
				}
				return true
			}

			// Test if the endpoint of our gateway application server works.
			// There is a retry loop here since we might need to wait a bit
			// for the application server to come up.
			backoff, err := retry.NewFibonacci(20 * time.Millisecond)
			require.Nil(t, err)

			var actualResponse string
			backoff = retry.WithMaxDuration(5*time.Second, backoff)
			_ = retry.Do(ctx, backoff, func(ctx context.Context) error {
				actualResponse, err = doGatewayRequestResponse(ctx, serverAddr, scenario.rawJWT)
				if isResponseExpected(actualResponse, err) {
					return nil
				}
				if err != nil {
					return retry.RetryableError(err)
				}
				return nil
			})
			if len(scenario.expectedError) > 0 {
				require.Error(t, err)
				require.Equal(t, scenario.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				for _, expectedFragment := range scenario.expectedResponseFragments {
					require.Contains(t, actualResponse, expectedFragment)
				}
			}
			// FIXME how do we stop the application server?
		})
	}
}
