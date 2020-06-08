package core

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/common"

	"github.com/anz-bank/sysl-go/status"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/anz-bank/sysl-go/config"

	"github.com/stretchr/testify/assert"
)

func newString(s string) *string {
	return &s
}

func Test_prepareMiddleware(t *testing.T) {
	ctx, _ := common.NewTestContextWithLoggerHook()

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
				common.CoreRequestContextMiddleware(),
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
	ctx, _ := common.NewTestContextWithLoggerHook()

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
