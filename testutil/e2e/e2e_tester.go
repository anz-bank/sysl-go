package e2e

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"gopkg.in/yaml.v2"

	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/status"
	"github.com/anz-bank/sysl-go/syslgo"
)

type Endpoint interface {
	Expect(tests ...Tests) Endpoint
	ExpectSimple(expectedHeaders map[string]string, expectedBody []byte, returnCode int,
		returnHeaders map[string]string, returnBody []byte, extraTests ...Tests) Endpoint
}

type Tester struct {
	t syslgo.TestingT

	restServer      *httptest.Server
	restDownstreams map[string]*restDownstream
	restCfgBasePath string

	grpcServer   *grpc.Server
	grpcListener *bufconn.Listener

	bm status.BuildMetadata
}

func ConfigToYamlData(cfg interface{}, appCfgType reflect.Type) ([]byte, error) {
	switch cfg := cfg.(type) {
	case []byte:
		return cfg, nil
	case string:
		return []byte(cfg), nil
	default:
		reflectValue := reflect.ValueOf(cfg)
		for reflectValue.Kind() == reflect.Ptr {
			reflectValue = reflectValue.Elem()
		}
		if reflectValue.Type() == appCfgType {
			cfg = struct {
				App interface{} `yaml:"app" mapstructure:"app"`
			}{App: cfg}
		}
		return yaml.Marshal(cfg)
	}
}

//nolint:golint // context.Context should be the first parameter of a function
func NewTester(t syslgo.TestingT, ctx context.Context, yamlConfigData []byte) (*Tester, context.Context, *core.Hooks) {
	return NewTesterWithBuildMetadata(t, ctx, status.BuildMetadata{}, yamlConfigData)
}

//nolint:golint // context.Context should be the first parameter of a function
func NewTesterWithBuildMetadata(t syslgo.TestingT, ctx context.Context, bm status.BuildMetadata, yamlConfigData []byte) (*Tester, context.Context, *core.Hooks) {
	e2eTester := &Tester{
		t:               t,
		restDownstreams: make(map[string]*restDownstream),
		bm:              bm,
	}

	ctx = core.WithConfigFile(ctx, yamlConfigData)

	hooks := &core.Hooks{
		ShouldSetGrpcGlobalLogger:  func() bool { return false },
		HTTPClientBuilder:          e2eTester.HTTPClientGetter,
		StoppableServerBuilder:     e2eTester.prepareServerListener,
		StoppableGrpcServerBuilder: e2eTester.prepareGrpcServerListener,
	}

	return e2eTester, ctx, hooks
}

func (b *Tester) T() syslgo.TestingT {
	return b.t
}

func (b *Tester) EndpointURL(suffix string) string {
	return b.restServer.URL + suffix
}

func (b *Tester) Close() {
	for _, be := range b.restDownstreams {
		be.close()
	}
	if b.restServer != nil {
		b.restServer.Close()
	}
	if b.grpcServer != nil {
		b.grpcServer.Stop()
	}
}

func (b *Tester) BuildMetadata() *status.BuildMetadata {
	return &b.bm
}

func (b *Tester) BuildID() string {
	return fmt.Sprintf("%s/%s", b.bm.Name, b.bm.Version)
}

func (b *Tester) CfgBasePath() string {
	return b.restCfgBasePath
}

func (b *Tester) NewDownstream(host, method, path string) Endpoint {
	be, ok := b.restDownstreams[host]
	if !ok {
		panic(fmt.Sprintf("downstream %s not expected", host))
	}

	return be.init(method, path)
}

func (b *Tester) Do(tc TestCall) {
	reqBody := strings.NewReader(tc.Body)
	req, err := http.NewRequestWithContext(context.Background(), tc.Method, b.EndpointURL(tc.URL), reqBody)
	require.NoError(b.t, err)
	req.Header = makeHeader(tc.Headers)
	// nolint: bodyclose // helper calls close()
	resp, err := b.restServer.Client().Do(req)
	require.NoError(b.t, err)
	if tc.TestCodeFn != nil {
		tc.TestCodeFn(b.t, tc.ExpectedCode, resp.StatusCode)
	} else {
		assert.Equal(b.t, tc.ExpectedCode, resp.StatusCode)
	}
	actualResp := GetResponseBodyAndClose(resp.Body)
	if tc.TestBodyFn != nil {
		tc.TestBodyFn(b.t, tc.ExpectedBody, string(actualResp))
	} else {
		assert.JSONEq(b.t, tc.ExpectedBody, string(actualResp))
	}
}

func (b *Tester) Do2(tc TestCall2) {
	var reqBody io.Reader
	if tc.Body != nil {
		reqBody = bytes.NewReader(tc.Body)
	}
	req, err := http.NewRequestWithContext(context.Background(), tc.Method, b.EndpointURL(tc.URL), reqBody)
	require.NoError(b.t, err)
	req.Header = makeHeader(tc.Headers)
	// nolint: bodyclose // helper calls close()
	resp, err := b.restServer.Client().Do(req)
	require.NoError(b.t, err)
	require.NotNil(b.t, resp)

	if tc.ExpectedCode != nil {
		assert.Equal(b.t, *tc.ExpectedCode, resp.StatusCode)
	}
	if tc.TestCodeFn != nil {
		tc.TestCodeFn(b.t, resp.StatusCode)
	}

	actualResp := GetResponseBodyAndClose(resp.Body)
	if tc.ExpectedBody != nil {
		if ct, ok := resp.Header[ContentTypeKey]; ok && strings.Contains(ct[0], "application/json") {
			assert.JSONEq(b.t, string(tc.ExpectedBody), string(actualResp))
		} else {
			assert.Equal(b.t, tc.ExpectedBody, actualResp)
		}
	}
	if tc.TestBodyFn != nil {
		tc.TestBodyFn(b.t, actualResp)
	}

	for _, testRespFns := range tc.TestRespFns {
		testRespFns(b.t, resp)
	}
}

func (b *Tester) HTTPClientGetter(host string) (*http.Client, string, error) {
	if be, ok := b.restDownstreams[host]; ok {
		return be.getClient()
	}
	be := newBackEnd(b.t, host)
	b.restDownstreams[host] = be

	return be.getClient()
}

func (b *Tester) prepareServerListener(ctx context.Context, rootRouter http.Handler, _ *tls.Config, cfg config.CommonHTTPServerConfig, _ string) core.StoppableServer {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rootRouter.ServeHTTP(w, r.WithContext(ctx))
	})
	b.restServer = httptest.NewServer(handler)
	b.restCfgBasePath = cfg.BasePath

	return b
}

const bufSize = 1024 * 1024

func (b *Tester) prepareGrpcServerListener(_ context.Context, server *grpc.Server, _ config.GRPCServerConfig, _ string) core.StoppableServer {
	b.grpcServer = server
	b.grpcListener = bufconn.Listen(bufSize)

	return b
}

func (b *Tester) GetBufDialer(context.Context, string) (net.Conn, error) {
	return b.grpcListener.Dial()
}

func (b *Tester) Start() error {
	if b.grpcServer != nil {
		go func() {
			if err := b.grpcServer.Serve(b.grpcListener); err != nil {
				panic(err)
			}
		}()
	}

	return nil
}

func (b *Tester) GracefulStop() error {
	if b.grpcServer != nil {
		b.grpcServer.GracefulStop()
		b.grpcServer = nil
	}

	b.Close()

	return nil
}

func (b *Tester) Stop() error {
	b.Close()

	return nil
}

func (b *Tester) GetName() string {
	return "Test Server"
}
