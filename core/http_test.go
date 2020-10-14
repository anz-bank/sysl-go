package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/testutil"

	"github.com/anz-bank/sysl-go/status"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/anz-bank/sysl-go/config"

	"github.com/stretchr/testify/assert"

	"github.com/sethvargo/go-retry"
)

func newString(s string) *string {
	return &s
}

func Test_prepareMiddleware(t *testing.T) {
	ctx, _ := testutil.NewTestContextWithLoggerHook()

	type args struct {
		cfg           *config.LibraryConfig
		buildMetadata *status.BuildMetadata
		promRegistry  *prometheus.Registry
	}
	tests := []struct {
		name    string
		args    args
		want    []func(handler http.Handler) http.Handler
		wantErr bool
	}{
		{
			name: "",
			args: args{
				cfg:           &config.LibraryConfig{},
				buildMetadata: &status.BuildMetadata{},
				promRegistry:  nil,
			},
			want: []func(http.Handler) http.Handler{
				Recoverer(ctx),
				common.TraceabilityMiddleware(ctx),
				common.CoreRequestContextMiddlewareWithContext(ctx),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := prepareMiddleware(ctx, "server", tt.args.promRegistry)
			assert.NotEmpty(t, got)
		})
	}
}

func Test_makeNewServer(t *testing.T) {
	type args struct {
		router       http.Handler
		tlsConfig    *tls.Config
		serverConfig config.CommonHTTPServerConfig
	}
	tests := []struct {
		name string
		args args
		want *http.Server
	}{
		{
			name: "Test 1 - Valid Server configuration",
			args: args{
				router: nil,
				tlsConfig: &tls.Config{
					ServerName: "Hello",
					MinVersion: tls.VersionTLS12,
					MaxVersion: tls.VersionTLS12,
				},
				serverConfig: config.CommonHTTPServerConfig{
					BasePath: "/test",
					Common: config.CommonServerConfig{
						HostName: "",
						Port:     8080,
						TLS: &config.TLSConfig{
							MinVersion: newString("1.2"),
							MaxVersion: newString("1.2"),
						},
					},
				},
			},
			want: &http.Server{
				Addr: ":8080",
				TLSConfig: &tls.Config{
					ServerName: "Hello",
					MinVersion: tls.VersionTLS12,
					MaxVersion: tls.VersionTLS12,
				},
				ReadHeaderTimeout: 10 * time.Second,
				IdleTimeout:       5 * time.Second,
				MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := makeNewServer(tt.args.router, tt.args.tlsConfig, tt.args.serverConfig, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeNewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prepareServerListener(t *testing.T) {
	ctx, _ := testutil.NewTestContextWithLoggerHook()

	type args struct {
		rootRouter   http.Handler
		tlsConfig    *tls.Config
		commonConfig config.CommonHTTPServerConfig
	}
	tests := []struct {
		name string
		args args
		want func() error
	}{
		{
			name: "",
			args: args{
				rootRouter: nil,
				tlsConfig: &tls.Config{
					ServerName: "Hello",
					MinVersion: tls.VersionTLS12,
					MaxVersion: tls.VersionTLS12,
				},
				commonConfig: config.CommonHTTPServerConfig{
					BasePath: "/test",
					Common: config.CommonServerConfig{
						HostName: "",
						Port:     8080,
						TLS: &config.TLSConfig{
							MinVersion: newString("1.2"),
							MaxVersion: newString("1.3"),
						},
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, prepareServerListener(ctx, tt.args.rootRouter, tt.args.tlsConfig, tt.args.commonConfig))
		})
	}
}

func Test_SelectBasePath_BothEmptySelectsForwardSlash(t *testing.T) {
	assert.Equal(t, "/", SelectBasePath("", ""))
}

func Test_SelectBasePath_SpecEmptySelectsDynamic(t *testing.T) {
	assert.Equal(t, "/dynamic", SelectBasePath("", "/dynamic"))
}

func Test_SelectBasePath_DynamicEmptySelectsSpec(t *testing.T) {
	assert.Equal(t, "/spec", SelectBasePath("/spec", ""))
}

func Test_SelectBasePath_BothFilledSelectsDynamic(t *testing.T) {
	assert.Equal(t, "/dynamic", SelectBasePath("/spec", "/dynamic"))
}

func TestHTTPStoppableServerCanBeHardStopped(t *testing.T) {
	ctx := context.Background()
	cfg := config.CommonHTTPServerConfig{
		BasePath: "/",
		Common: config.CommonServerConfig{
			HostName: "localhost",
			Port:     8081,
			TLS:      nil,
		},
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "hello")
	})

	s := prepareServerListener(ctx, h, nil, cfg)

	go func() {
		err := s.Start()
		require.Equal(t, http.ErrServerClosed, err)
	}()

	healthCheck := func() error {
		resp, err := http.Get("http://localhost:8081/")
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http StatusCode %d", resp.StatusCode)
		}
		return nil
	}

	// Wait for server to come up
	backoff, err := retry.NewFibonacci(20 * time.Millisecond)
	require.NoError(t, err)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		err := healthCheck()
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})
	require.NoError(t, err)

	// Hard stop the server
	err = s.Stop()
	require.NoError(t, err)

	// Check server has indeed stopped
	require.Equal(t, http.ErrServerClosed, s.Start())
}

func TestHTTPStoppableServerCanBeGracefullyStopped(t *testing.T) {
	ctx := context.Background()
	cfg := config.CommonHTTPServerConfig{
		BasePath: "/",
		Common: config.CommonServerConfig{
			HostName: "localhost",
			Port:     8081,
			TLS:      nil,
		},
	}

	// use these channels as "side-channels" so we can sense when server
	// began processing a slow request and then control how long it
	// takes until it "completes" processing and begins to write a
	// response
	started := make(chan struct{})
	complete := make(chan struct{})
	gotResponse := make(chan struct{})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slowMode := strings.HasSuffix(r.URL.String(), "slow")
		if slowMode {
			started <- struct{}{}
			// wait, pretending to busily compute, until we're told to complete
			<-complete
		}
		fmt.Fprintf(w, "hello")
	})

	s := prepareServerListener(ctx, h, nil, cfg)

	go func() {
		err := s.Start()
		require.Equal(t, http.ErrServerClosed, err)
	}()

	healthCheck := func(suffix string) error {
		resp, err := http.Get("http://localhost:8081/" + suffix)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http StatusCode %d", resp.StatusCode)
		}
		return nil
	}

	// Wait for server to come up
	backoff, err := retry.NewFibonacci(20 * time.Millisecond)
	require.NoError(t, err)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		err := healthCheck("")
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})
	require.NoError(t, err)

	// make an unusally slow request
	go func() {
		defer func() {
			gotResponse <- struct{}{}
		}()
		err := healthCheck("slow")
		require.NoError(t, err)
	}()

	// wait for server to begin processing a "slow" request
	<-started

	go func() {
		// FIXME no guarantee the graceful stop operation starts happening
		// before we've received a response to the slow request.
		err = s.GracefulStop()
		require.NoError(t, err)
	}()

	// let server begin writing response to "slow" request
	complete <- struct{}{}

	// wait until we got a response
	<-gotResponse

	// Check server has indeed stopped
	require.Equal(t, http.ErrServerClosed, s.Start())
}

func TestHTTPStoppableServerGracefulStopTimeout(t *testing.T) {
	ctx := context.Background()
	cfg := config.CommonHTTPServerConfig{
		BasePath: "/",
		Common: config.CommonServerConfig{
			HostName: "localhost",
			Port:     8081,
			TLS:      nil,
		},
	}

	// use this channel as a side-channel so we can sense when server
	// began processing a slow request
	started := make(chan struct{})

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slowMode := strings.HasSuffix(r.URL.String(), "slow")
		if slowMode {
			started <- struct{}{}

			// wait, pretending to busily compute, until the request is cancelled.
			<-r.Context().Done()
			return
		}
		fmt.Fprintf(w, "hello")
	})

	s := prepareServerListener(ctx, h, nil, cfg)
	s.gracefulStopTimeout = 10 * time.Millisecond

	go func() {
		err := s.Start()
		require.Equal(t, http.ErrServerClosed, err)
	}()

	healthCheck := func(suffix string) error {
		resp, err := http.Get("http://localhost:8081/" + suffix)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http StatusCode %d", resp.StatusCode)
		}
		return nil
	}

	// Wait for server to come up
	backoff, err := retry.NewFibonacci(20 * time.Millisecond)
	require.NoError(t, err)
	backoff = retry.WithMaxDuration(5*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		err := healthCheck("")
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})
	require.NoError(t, err)

	// make an unusally slow request that will actually never get served
	done := make(chan struct{})
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		err := healthCheck("slow")
		// since we're making a request to localhost we should be lucky
		// enough to get told that the server has rudely closed the
		// connection before sending a HTTP response.
		require.Error(t, err)
		require.Contains(t, err.Error(), "EOF")
	}()

	// wait for server to begin processing the "slow" request
	<-started

	go func() {
		// FIXME no guarantee this call happens before we let the "complete" the slow request
		// tell server to gracefully stop
		err = s.GracefulStop()
		require.NoError(t, err)
	}()

	<-done

	// Check server has indeed stopped
	require.Equal(t, http.ErrServerClosed, s.Start())
}
