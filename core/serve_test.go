package core

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"

	pkg "github.com/anz-bank/pkg/log"

	"github.com/anz-bank/sysl-go/common"
	"github.com/sirupsen/logrus"

	"github.com/anz-bank/sysl-go/log"

	"github.com/anz-bank/sysl-go/config"
	"github.com/stretchr/testify/assert"
)

type TestServiceInterface struct{}

type TestAppConfig struct {
	testFn func()
	field1 int `mapstructure:"field1"`
	field2 int `yaml:"field2"`
	Field3 int
}

const errString = "not happening"

func TestNewServerReturnsErrorIfNewManagerReturnsError(t *testing.T) {
	// Override sysl-go app command line interface to directly pass in app config
	ctx := WithConfigFile(context.Background(), []byte(""))

	srv, err := NewServer(
		ctx,
		struct{}{},
		func(ctx context.Context, config TestAppConfig) (*TestServiceInterface, *Hooks, error) {
			return &TestServiceInterface{}, nil, nil
		},
		&TestServiceInterface{},
		func(ctx context.Context, serviceIntf interface{}, _ *Hooks) (Manager, *GrpcServerManager, error) {
			return nil, nil, fmt.Errorf(errString)
		},
	)
	assert.Nil(t, srv)
	assert.EqualError(t, err, errString)
}

func TestNewServerReturnsErrorIfValidateConfigReturnsError(t *testing.T) {
	// Override sysl-go app command line interface to directly pass in app config
	ctx := WithConfigFile(context.Background(), []byte(""))

	hooks := &Hooks{
		ValidateConfig: func(ctx context.Context, cfg *config.DefaultConfig) error {
			return fmt.Errorf(errString)
		},
	}

	srv, err := NewServer(
		ctx,
		struct{}{},
		func(ctx context.Context, config TestAppConfig) (*TestServiceInterface, *Hooks, error) {
			return &TestServiceInterface{}, hooks, nil
		},
		&TestServiceInterface{},
		func(ctx context.Context, serviceIntf interface{}, _ *Hooks) (Manager, *GrpcServerManager, error) {
			return nil, nil, nil
		},
	)
	assert.Nil(t, srv)
	assert.EqualError(t, err, errString)
}

// Test a new server initialises a logger.
func TestNewServerInitialisesLogger(t *testing.T) {
	ctx, err := newServerContext(context.Background())
	assert.Nil(t, err)
	logger := log.GetLogger(ctx)
	assert.NotNil(t, logger)
}

// Test a new server initialises a bootstrap logger that gets overwritten with a custom logger.
func TestNewServerInitialisesLogger_bootstrap(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx, err := newServerContextWithCreateService(context.Background(),
		func(ctx context.Context, config TestAppConfig) (*TestServiceInterface, *Hooks, error) {
			log.Info(ctx, "bootstrap log")
			return &TestServiceInterface{}, &Hooks{
				Logger: func() log.Logger {
					return log.NewPkgLogger(pkg.Fields{}.WithConfigs(pkg.SetOutput(buf)))
				},
			}, nil
		})

	assert.Nil(t, err)
	logger := log.GetLogger(ctx)
	assert.NotNil(t, logger)
	logger.Info("hello")
	assert.Contains(t, buf.String(), "hello")
	assert.NotContains(t, buf.String(), "bootstrap")
}

// Test a new server initialises a logger with an appropriate log level.
func TestNewServerInitialisesLogger_logLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := pkg.Fields{}.WithConfigs(pkg.SetOutput(buf)).Onto(context.Background())
	ctx = WithConfigFile(ctx, []byte("library:\n  log:\n    level: error"))
	ctx, err := newServerContext(ctx)

	assert.Nil(t, err)
	logger := log.GetLogger(ctx)
	assert.NotNil(t, logger)
	logger.Debug("hello")
	assert.NotContains(t, buf.String(), "hello")
	logger.Info("hello")
	assert.Contains(t, buf.String(), "hello")
}

// Test a new server initialises a logger from the hooks.
func TestNewServerInitialisesLogger_hooks(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx, err := newServerContextWithHooks(context.Background(), &Hooks{
		Logger: func() log.Logger {
			return log.NewPkgLogger(pkg.Fields{}.WithConfigs(pkg.SetOutput(buf)))
		},
	})

	assert.Nil(t, err)
	logger := log.GetLogger(ctx)
	assert.NotNil(t, logger)
	logger.Info("hello")
	assert.Contains(t, buf.String(), "hello")
}

// Test a new server initialises a suitable logger if the logrus logger is found in the context.
func TestNewServerInitialisesLogger_externalLogrusLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logrus.New()
	l.Out = buf
	ctx := common.LoggerToContext(context.Background(), l, nil) //nolint:staticcheck
	ctx, err := newServerContextWithHooks(ctx, &Hooks{
		Logger: func() log.Logger {
			t.Fatal("hook should not be called")
			return nil
		},
	})

	assert.Nil(t, err)
	logger := log.GetLogger(ctx)
	assert.NotNil(t, logger)
	logger.Info("hello")
	assert.Contains(t, buf.String(), "hello")
}

// Test a new server initialises a suitable logger if the pkg logger is found in the context.
func TestNewServerInitialisesLogger_externalPkgLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := pkg.Fields{}.WithConfigs(pkg.SetOutput(buf)).Onto(context.Background())
	ctx, err := newServerContextWithHooks(ctx, &Hooks{
		Logger: func() log.Logger {
			t.Fatal("hook should not be called")
			return nil
		},
	})

	assert.Nil(t, err)
	logger := log.GetLogger(ctx)
	assert.NotNil(t, logger)
	logger.Info("hello")
	assert.Contains(t, buf.String(), "hello")
}

// newServerContext returns the context used against the server returned from NewServer.
func newServerContext(ctx context.Context) (context.Context, error) {
	return newServerContextWithHooks(ctx, nil)
}

// newServerContextWithHooks returns the context used against the server returned from NewServer.
func newServerContextWithHooks(ctx context.Context, hooks *Hooks) (context.Context, error) {
	return newServerContextWithCreateService(ctx, func(ctx context.Context, config TestAppConfig) (*TestServiceInterface, *Hooks, error) {
		return &TestServiceInterface{}, hooks, nil
	})
}

// newServerContextWithHooks returns the context used against the server returned from NewServer.
func newServerContextWithCreateService(ctx context.Context,
	createService func(ctx context.Context, config TestAppConfig) (*TestServiceInterface, *Hooks, error)) (context.Context, error) {
	cfg := ctx.Value(serveYAMLConfigFileKey)
	if cfg == nil {
		ctx = WithConfigFile(ctx, []byte(""))
	}
	srv, err := NewServer(ctx, struct{}{}, createService,
		&TestServiceInterface{},
		func(ctx context.Context, serviceIntf interface{}, _ *Hooks) (Manager, *GrpcServerManager, error) {
			return nil, nil, nil
		},
	)
	if err != nil {
		return nil, err
	}
	return srv.(*autogenServer).ctx, nil
}

func TestDescribeYAMLForType(t *testing.T) {
	t.Parallel()

	// for log.Level
	w := bytes.Buffer{}
	describeYAMLForType(&w, reflect.TypeOf(log.DebugLevel), map[reflect.Type]string{}, 0)
	assert.Equal(t, " \x1b[1m0\x1b[0m", w.String())
}

func TestDescribeYAMLForTypeContainsFuncs(t *testing.T) {
	t.Parallel()

	w := bytes.Buffer{}
	describeYAMLForType(&w, reflect.TypeOf(TestAppConfig{
		testFn: func() {},
		field1: 0,
		field2: 1,
	}), map[reflect.Type]string{}, 0)
	assert.Equal(t, "\nfield1: \x1b[1m0\x1b[0m\nfield2: \x1b[1m0\x1b[0m\nField3: \x1b[1m0\x1b[0m",
		w.String())
}

type testStoppableServer struct {
	start func() error
}

func (s *testStoppableServer) Start() error {
	if s.start != nil {
		return s.start()
	}
	return nil
}

func (s *testStoppableServer) Stop() error {
	return nil
}

func (s *testStoppableServer) GracefulStop() error {
	return nil
}

func (s testStoppableServer) GetName() string {
	return "testStoppableServer"
}

// Ensure that a panic in the Start of a subServer gets recovered.
func TestMultiStoppableServer_Start_WithPanic(t *testing.T) {
	server := &testStoppableServer{start: func() error {
		panic("panic")
		return nil
	}}

	ctx, err := newServerContext(ctx)
	assert.Nil(t, err)

	mServer := NewMultiStoppableServer(ctx, []StoppableServer{server})
	assert.NotPanics(t, func() {
		err = mServer.Start()
		assert.Error(t, err)
	})
}
