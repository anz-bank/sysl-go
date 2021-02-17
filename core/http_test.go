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

	"github.com/anz-bank/sysl-go/log"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"github.com/anz-bank/sysl-go/testutil"

	"github.com/anz-bank/sysl-go/status"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/anz-bank/sysl-go/config"

	"github.com/stretchr/testify/assert"

	"github.com/sethvargo/go-retry"
)

var contextTimeout = time.Second

func newString(s string) *string {
	return &s
}

type restManagerImpl struct {
	handlers               func() []handlerinitialiser.HandlerInitialiser
	library                func() *config.LibraryConfig
	admin                  func() *config.CommonHTTPServerConfig
	public                 func() *config.UpstreamConfig
	addAdminHTTPMiddleware func() func(ctx context.Context, r chi.Router)
}

func (r *restManagerImpl) EnabledHandlers() []handlerinitialiser.HandlerInitialiser {
	return r.handlers()
}

func (r *restManagerImpl) LibraryConfig() *config.LibraryConfig {
	return r.library()
}

func (r *restManagerImpl) AdminServerConfig() *config.CommonHTTPServerConfig {
	return r.admin()
}

func (r *restManagerImpl) PublicServerConfig() *config.UpstreamConfig {
	return r.public()
}

func (r *restManagerImpl) AddAdminHTTPMiddleware() func(ctx context.Context, r chi.Router) {
	if r.addAdminHTTPMiddleware != nil {
		return r.addAdminHTTPMiddleware()
	}

	return nil
}

func Test_prepareMiddleware(t *testing.T) {
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
				Recoverer,
				common.TraceabilityMiddleware,
				common.CoreRequestContextMiddleware,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := prepareMiddleware("server", tt.args.promRegistry, contextTimeout)
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
			type testContextKey struct{}
			ctx := context.WithValue(context.Background(), testContextKey{}, 5)
			got := makeNewServer(ctx, tt.args.router, tt.args.tlsConfig, tt.args.serverConfig, nil)
			assert.Equal(t, ctx, got.BaseContext(nil))
			got.BaseContext = nil
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeNewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prepareServerListener(t *testing.T) {
	ctx := testutil.NewTestContext()

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
			assert.NotNil(t, prepareServerListener(ctx, tt.args.rootRouter, tt.args.tlsConfig, tt.args.commonConfig, tt.name))
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
	ctx := testutil.NewTestContext()
	cfg := config.CommonHTTPServerConfig{
		BasePath: "/",
		Common: config.CommonServerConfig{
			HostName: "localhost",
			Port:     8082,
			TLS:      nil,
		},
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "hello")
	})

	s := prepareServerListener(ctx, h, nil, cfg, "")

	go func() {
		err := s.Start()
		require.NoError(t, err)
	}()

	healthCheck := func() error {
		resp, err := http.Get("http://localhost:8082/")
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
	require.NoError(t, s.Start())
}

func TestHTTPStoppableServerCanBeGracefullyStopped(t *testing.T) {
	ctx := testutil.NewTestContext()
	cfg := config.CommonHTTPServerConfig{
		BasePath: "/",
		Common: config.CommonServerConfig{
			HostName: "localhost",
			Port:     8083,
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

	s := prepareServerListener(ctx, h, nil, cfg, "")

	go func() {
		err := s.Start()
		require.NoError(t, err)
	}()

	healthCheck := func(suffix string) error {
		resp, err := http.Get("http://localhost:8083/" + suffix)
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
	require.NoError(t, s.Start())
}

func TestHTTPStoppableServerGracefulStopTimeout(t *testing.T) {
	ctx := testutil.NewTestContext()
	cfg := config.CommonHTTPServerConfig{
		BasePath: "/",
		Common: config.CommonServerConfig{
			HostName: "localhost",
			Port:     8084,
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

	s := prepareServerListener(ctx, h, nil, cfg, "")
	s.gracefulStopTimeout = 10 * time.Millisecond

	go func() {
		err := s.Start()
		require.NoError(t, err)
	}()

	healthCheck := func(suffix string) error {
		resp, err := http.Get("http://localhost:8084/" + suffix)
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
		// tell server to gracefully stop. in this scenario, since
		// the "slow" request is rigged to never return, by itself,
		// we expect, this graceful stop to timeout and then actually
		// fallback to doing a non-graceful stop.
		err = s.GracefulStop()
		require.NoError(t, err)
	}()

	<-done

	// Check server has indeed stopped
	require.NoError(t, s.Start())
}

func Test_configureAdminServerListener_Valid(t *testing.T) {
	ctx := testutil.NewTestContext()

	manager := &restManagerImpl{
		handlers: func() []handlerinitialiser.HandlerInitialiser { return []handlerinitialiser.HandlerInitialiser{} },
		library: func() *config.LibraryConfig {
			return &config.LibraryConfig{
				Log: config.LogConfig{
					Format:       "text",
					Level:        log.DebugLevel,
					ReportCaller: false,
				},
				Profiling:      true,
				Health:         true,
				Authentication: nil,
			}
		},
		admin: func() *config.CommonHTTPServerConfig {
			return &config.CommonHTTPServerConfig{
				Common:       config.CommonServerConfig{HostName: "localhost", Port: 9494, TLS: nil},
				BasePath:     "/",
				ReadTimeout:  time.Minute,
				WriteTimeout: time.Minute,
			}
		},
		public: func() *config.UpstreamConfig { return &config.UpstreamConfig{ContextTimeout: contextTimeout} },
	}

	mWare := prepareMiddleware("test", nil, contextTimeout)

	srv, err := configureAdminServerListener(ctx, manager, nil, nil, mWare.admin)
	require.NotNil(t, srv)
	require.NoError(t, err)

	defer func() {
		err = srv.Stop()
	}()

	go func() {
		err := srv.Start()
		require.NoError(t, err)
	}()
}

func Test_configureAdminServerListener_MissingLibraryConfig(t *testing.T) {
	ctx := testutil.NewTestContext()

	manager := &restManagerImpl{
		handlers: func() []handlerinitialiser.HandlerInitialiser { return []handlerinitialiser.HandlerInitialiser{} },
		library:  func() *config.LibraryConfig { return nil },
		admin: func() *config.CommonHTTPServerConfig {
			return &config.CommonHTTPServerConfig{
				Common:       config.CommonServerConfig{HostName: "localhost", Port: 9595, TLS: nil},
				BasePath:     "/",
				ReadTimeout:  time.Minute,
				WriteTimeout: time.Minute,
			}
		},
		public: func() *config.UpstreamConfig { return &config.UpstreamConfig{ContextTimeout: contextTimeout} },
	}

	mWare := prepareMiddleware("test", nil, contextTimeout)

	srv, err := configureAdminServerListener(ctx, manager, nil, nil, mWare.admin)
	require.Nil(t, srv)
	require.Error(t, err)
}

func Test_configureAdminServerListener_MissingAdminConfig(t *testing.T) {
	ctx := testutil.NewTestContext()

	manager := &restManagerImpl{
		handlers: func() []handlerinitialiser.HandlerInitialiser { return []handlerinitialiser.HandlerInitialiser{} },
		library: func() *config.LibraryConfig {
			return &config.LibraryConfig{
				Log: config.LogConfig{
					Format:       "text",
					Level:        log.DebugLevel,
					ReportCaller: false,
				},
				Profiling:      true,
				Health:         true,
				Authentication: nil,
			}
		},
		admin: func() *config.CommonHTTPServerConfig {
			return nil
		},
		public: func() *config.UpstreamConfig { return &config.UpstreamConfig{ContextTimeout: contextTimeout} },
	}

	mWare := prepareMiddleware("test", nil, contextTimeout)

	srv, err := configureAdminServerListener(ctx, manager, nil, nil, mWare.admin)
	require.Nil(t, srv)
	require.Error(t, err)
}

func Test_configureAdminServerListener_MissingMiddlewareHandler(t *testing.T) {
	ctx := testutil.NewTestContext()

	manager := &restManagerImpl{
		handlers: func() []handlerinitialiser.HandlerInitialiser { return []handlerinitialiser.HandlerInitialiser{} },
		library: func() *config.LibraryConfig {
			return &config.LibraryConfig{
				Log: config.LogConfig{
					Format:       "text",
					Level:        log.DebugLevel,
					ReportCaller: false,
				},
				Profiling:      true,
				Health:         true,
				Authentication: nil,
			}
		},
		admin: func() *config.CommonHTTPServerConfig {
			return &config.CommonHTTPServerConfig{
				Common:       config.CommonServerConfig{HostName: "localhost", Port: 9494, TLS: nil},
				BasePath:     "/",
				ReadTimeout:  time.Minute,
				WriteTimeout: time.Minute,
			}
		},
		public: func() *config.UpstreamConfig { return &config.UpstreamConfig{ContextTimeout: contextTimeout} },
	}

	srv, err := configureAdminServerListener(ctx, manager, nil, nil, nil)
	require.NotNil(t, srv)
	require.NoError(t, err)
}
