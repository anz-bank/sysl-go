package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"google.golang.org/grpc/metadata"
	"grpc_jwt_authorization/internal/gen/pkg/servers/gateway"

	"github.com/anz-bank/sysl-go/syslgo"

	pb "grpc_jwt_authorization/internal/gen/pb/gateway"

	"github.com/anz-bank/sysl-go/jwtauth/jwttest"

	"github.com/stretchr/testify/require"
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
`)

var appCfgSix = []byte(`---
app:
development:
  disableAllAuthorizationRules: true
`)

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
			expectedError: "rpc error: code = Unknown desc = jwtauth err 1: jwt parse error: go-jose/go-jose: compact JWS format must have three parts", // FIXME impl detail leak in error msg
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
			gatewayTester := gateway.NewTestServer(t, context.Background(), createService, scenario.appCfg)
			defer gatewayTester.Close()

			gatewayTester.Hello().
				WithRequest(&pb.HelloRequest{Content: "echo"}).
				WithContext(metadata.AppendToOutgoingContext(context.Background(), "Authorization", "bearer "+scenario.rawJWT)).
				TestResponse(func(t syslgo.TestingT, actualResponse *pb.HelloResponse, err error) {
					if len(scenario.expectedError) > 0 {
						require.Error(t, err)
						require.Equal(t, scenario.expectedError, err.Error())
					} else {
						require.NoError(t, err)
						for _, expectedFragment := range scenario.expectedResponseFragments {
							require.Contains(t, actualResponse.Content, expectedFragment)
						}
					}
				}).
				Send()
		})
	}
}
