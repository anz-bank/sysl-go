package e2e

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/status"
)

type Endpoint interface {
	Expect(tests ...Tests) Endpoint
	ExpectSimple(expectedHeaders map[string]string, expectedBody []byte, returnCode int,
		returnHeaders map[string]string, returnBody []byte, extraTests ...Tests) Endpoint
}

type Tester struct {
	t *testing.T
	*httptest.Server
	restDownstreams map[string]*restDownstream
	bm              status.BuildMetadata
	cfgBasePath     string
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
func NewTester(t *testing.T, ctx context.Context, yamlConfigData []byte) (*Tester, context.Context, *core.Hooks) {
	return NewTesterWithBuildMetadata(t, ctx, status.BuildMetadata{}, yamlConfigData)
}

//nolint:golint // context.Context should be the first parameter of a function
func NewTesterWithBuildMetadata(t *testing.T, ctx context.Context, bm status.BuildMetadata, yamlConfigData []byte) (*Tester, context.Context, *core.Hooks) {
	e2eTester := &Tester{
		t:               t,
		restDownstreams: make(map[string]*restDownstream),
		bm:              bm,
	}

	ctx = core.WithConfigFile(ctx, yamlConfigData)

	hooks := &core.Hooks{
		HTTPClientBuilder:      e2eTester.HTTPClientGetter,
		StoppableServerBuilder: e2eTester.prepareServerListener,
	}

	return e2eTester, ctx, hooks
}

func (b *Tester) EndpointURL(suffix string) string {
	return b.URL + suffix
}

func (b *Tester) Close() {
	for _, be := range b.restDownstreams {
		be.close()
	}
	b.Server.Close()
}

func (b *Tester) BuildMetadata() *status.BuildMetadata {
	return &b.bm
}

func (b *Tester) BuildID() string {
	return fmt.Sprintf("%s/%s", b.bm.Name, b.bm.Version)
}

func (b *Tester) CfgBasePath() string {
	return b.cfgBasePath
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
	resp, err := b.Client().Do(req)
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
	resp, err := b.Client().Do(req)
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
	b.Server = httptest.NewServer(handler)
	b.cfgBasePath = cfg.BasePath

	return b
}

func (b *Tester) Start() error {
	return nil
}

func (b *Tester) GracefulStop() error {
	b.Close()

	return nil
}

func (b *Tester) Stop() error {
	b.Close()

	return nil
}

func (b *Tester) GetName() string {
	return "REST Test Server"
}
