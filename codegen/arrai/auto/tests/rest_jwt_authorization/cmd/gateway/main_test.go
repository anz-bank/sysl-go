package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/config/envvar"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/jwtauth/jwttest"

	"github.com/sethvargo/go-retry"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var appCfgOne = []byte(`---
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
    http:
      basePath: "/"
      common:
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
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
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
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
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
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
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
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`)

var appCfgSix = []byte(`---
app:
development:
  disableAllAuthorizationRules: true
genCode:
  upstream:
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`)

func getServerAddr(appCfg []byte) (string, error) {
	cfg := config.DefaultConfig{}
	memFs := afero.NewMemMapFs()
	err := afero.Afero{Fs: memFs}.WriteFile("config.yaml", appCfg, 0777)
	if err != nil {
		return "", err
	}
	b := envvar.NewConfigReaderBuilder().WithFs(memFs).WithConfigFile("config.yaml")

	err = b.Build().Unmarshal(&cfg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", cfg.GenCode.Upstream.HTTP.Common.HostName, cfg.GenCode.Upstream.HTTP.Common.Port), nil
}

func doGatewayRequestResponse(ctx context.Context, addr string, rawJWT string) (string, int, error) {
	// Naive hand-written http client that attempts to call the Gateway service's endpoint.
	// This does not attempt to depend on generated code or sysl-go's core libraries, as we want to be
	// able to tell if the codegen or sysl-go libraries are defective or doing something unusual.
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://"+addr+"/hello", strings.NewReader(`{"content": "hey"}`))
	if err != nil {
		return "", -1, err
	}
	req.Header.Add("Authorization", "Bearer "+rawJWT)
	resp, err := client.Do(req)
	if err != nil {
		return "", -1, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", -1, err
	}
	return string(data), resp.StatusCode, nil
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

func TestJWTAuthorizationOfRESTEndpoints(t *testing.T) {

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
		expectedStatus            int
	}

	scenarios := []Scenario{
		{
			name:                      "request with authorised claims signed by trusted issuer succeeds",
			appCfg:                    appCfgOne,
			rawJWT:                    mustMakeTestJWT(trustedIssuer, map[string]interface{}{"scope": "hello"}),
			expectedResponseFragments: []string{"why hello there"},
			expectedStatus:            http.StatusOK,
		},
		{
			name:                      "request with unauthorised claims signed by trusted issuer fails",
			appCfg:                    appCfgTwo,
			rawJWT:                    mustMakeTestJWT(trustedIssuer, map[string]interface{}{"scope": "banana"}),
			expectedResponseFragments: []string{"Unauthorized error"},
			expectedStatus:            http.StatusUnauthorized,
		},
		{
			name:                      "request with authorised claims and extra claims signed by trusted issuer succeeds",
			appCfg:                    appCfgThree,
			rawJWT:                    mustMakeTestJWT(trustedIssuer, map[string]interface{}{"scope": "banana hello"}),
			expectedResponseFragments: []string{"why hello there"},
			expectedStatus:            http.StatusOK,
		},
		{
			name:                      "request with gibberish instead of JWT fails",
			appCfg:                    appCfgFour,
			rawJWT:                    "surprise!",
			expectedResponseFragments: []string{"Unauthorized error"},
			expectedStatus:            http.StatusUnauthorized,
		},
		{
			name:                      "request with authorised claims and signed by untrusted issuer fails",
			appCfg:                    appCfgFive,
			rawJWT:                    mustMakeTestJWT(untrustedIssuer, map[string]interface{}{"scope": "hello"}),
			expectedResponseFragments: []string{"Unauthorized error"},
			expectedStatus:            http.StatusUnauthorized,
		},
		{
			name:                      "request with authorised claims and signed by untrusted issuer succeeds if app configured to disable auth rules",
			appCfg:                    appCfgSix,
			rawJWT:                    mustMakeTestJWT(untrustedIssuer, map[string]interface{}{"scope": "hello"}),
			expectedResponseFragments: []string{"why hello there"},
			expectedError:             "",
			expectedStatus:            http.StatusOK,
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

			// Override sysl-go app command line interface to directly pass in app config
			ctx = core.WithConfigFile(ctx, []byte(scenario.appCfg))

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

			isResponseExpected := func(response string, status int, err error) bool {
				return status == scenario.expectedStatus
			}

			// Test if the endpoint of our gateway application server works.
			// There is a retry loop here since we might need to wait a bit
			// for the application server to come up.
			backoff, err := retry.NewFibonacci(20 * time.Millisecond)
			require.Nil(t, err)

			var actualResponse string
			var status int
			backoff = retry.WithMaxDuration(5*time.Second, backoff)
			_ = retry.Do(ctx, backoff, func(ctx context.Context) error {
				actualResponse, status, err = doGatewayRequestResponse(ctx, serverAddr, scenario.rawJWT)
				if isResponseExpected(actualResponse, status, err) {
					return nil
				}
				if err != nil {
					return retry.RetryableError(err)
				}
				return nil
			})
			require.NoError(t, err)
			require.Equal(t, scenario.expectedStatus, status)
			for _, expectedFragment := range scenario.expectedResponseFragments {
				require.Contains(t, actualResponse, expectedFragment)
			}
		})
	}
}
