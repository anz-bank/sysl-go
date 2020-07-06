package core

// MARKED TO IGNORE COVERAGE

import (
	"context"
	"errors"

	"github.com/anz-bank/sysl-go/logconfig"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type ServerParams struct {
	Ctx                context.Context
	Name               string
	logrusLogger       *logrus.Logger
	pkgLoggerConfigs   []log.Config
	restManager        Manager
	grpcManager        GrpcManager
	prometheusRegistry *prometheus.Registry
}
type emptyWriter struct {
}

func (e *emptyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

//nolint:gocognit // Long method names are okay because only generated code will call this, not humans.
func NewServerParams(ctx context.Context, name string, opts ...ServerOption) *ServerParams {
	params := &ServerParams{Ctx: ctx, Name: name}
	for _, o := range opts {
		o.apply(params)
	}
	return params
}

//nolint:gocognit,funlen // Long method are okay because only generated code will call this, not humans.
func (params *ServerParams) Start() error {
	ctx := params.Ctx

	// initialise the logger
	// sysl-go always uses a pkg logger internally. if custom code passes in a logrus logger, a
	// mechanism which is deprecated, then a hook is added to the internal pkg logger that forwards
	// logged events to the provided logrus logger.
	// sysl-go can be requested to log in a verbose manner. logger in a verbose manner logs additional
	// details within log events where appropriate. the mechanism to set this verbose manner is to
	// either have a sufficiently high logrus log level or the verbose mode set against the pkg logger.
	configs := params.pkgLoggerConfigs
	verboseLogging := false
	if params.logrusLogger != nil {
		// set an empty io writter against pkg logger
		// pkg logger just becomes a proxy that forwards all logs to logrus
		configs = append(configs, log.SetOutput(&emptyWriter{}))
		configs = append(configs, log.AddHooks(&logrusHook{params.logrusLogger}))
		configs = append(configs, log.SetLogCaller(params.logrusLogger.ReportCaller))
		ctx = common.LoggerToContext(ctx, params.logrusLogger, nil)
		verboseLogging = params.logrusLogger.Level >= logrus.DebugLevel
	}
	ctx = log.WithConfigs(configs...).Onto(ctx)
	verboseMode := log.SetVerboseMode(true)
	for _, config := range configs {
		if config == verboseMode {
			verboseLogging = true
			break
		}
	}

	// prepare the middleware
	ctx = logconfig.SetVerboseLogging(ctx, verboseLogging)
	mWare := prepareMiddleware(ctx, params.Name, params.prometheusRegistry)

	var restIsRunning, grpcIsRunning bool

	// Run the REST server
	var listenAdmin func() error
	if params.restManager != nil && params.restManager.AdminServerConfig() != nil {
		var err error
		listenAdmin, err = configureAdminServerListener(ctx, params.restManager, params.prometheusRegistry, mWare.admin)
		if err != nil {
			return err
		}
	} else {
		// set up a dummy listener which will never exit if admin disabled
		listenAdmin = func() error { select {} }
	}

	var listenPublic func() error
	if params.restManager != nil && params.restManager.PublicServerConfig() != nil {
		var err error
		listenPublic, err = configurePublicServerListener(ctx, params.restManager, mWare.public)
		if err != nil {
			return err
		}
		restIsRunning = true
	} else {
		listenPublic = func() error { select {} }
	}

	// Run the gRPC server
	var listenPublicGrpc func() error
	if params.grpcManager != nil && params.grpcManager.GrpcPublicServerConfig() != nil {
		var err error
		listenPublicGrpc, err = configurePublicGrpcServerListener(ctx, params.grpcManager)
		if err != nil {
			return err
		}

		grpcIsRunning = true
	} else {
		listenPublicGrpc = func() error { select {} }
	}

	// Panic if REST&gRPC are not running
	if !restIsRunning && !grpcIsRunning {
		err := errors.New("REST and gRPC servers cannot both be nil")
		log.Error(ctx, err)
		panic(err)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- listenPublic()
	}()
	go func() {
		errChan <- listenAdmin()
	}()
	go func() {
		errChan <- listenPublicGrpc()
	}()

	return <-errChan
}

type ServerOption interface {
	apply(params *ServerParams)
}

type restManagerOption struct {
	restManager Manager
}

func (o *restManagerOption) apply(params *ServerParams) {
	params.restManager = o.restManager
}

func WithRestManager(manager Manager) ServerOption {
	return &restManagerOption{manager}
}

type logrusLoggerOption struct {
	logger *logrus.Logger
}

func (o *logrusLoggerOption) apply(params *ServerParams) {
	params.logrusLogger = o.logger
}

// Deprecated: Use WithPkgLogger instead
func WithLogrusLogger(logger *logrus.Logger) ServerOption {
	return &logrusLoggerOption{logger}
}

type pkgLoggerOption struct {
	configs []log.Config
}

func (o *pkgLoggerOption) apply(params *ServerParams) {
	params.pkgLoggerConfigs = o.configs
}

func WithPkgLogger(configs ...log.Config) ServerOption {
	return &pkgLoggerOption{configs}
}

func WithPrometheusRegistry(prometheusRegistry *prometheus.Registry) ServerOption {
	return &prometheusRegistryOption{prometheusRegistry}
}

type prometheusRegistryOption struct {
	prometheusRegistry *prometheus.Registry
}

func (o *prometheusRegistryOption) apply(params *ServerParams) {
	params.prometheusRegistry = o.prometheusRegistry
}

type grpcManagerOption struct {
	grpcManager GrpcManager
}

func (o *grpcManagerOption) apply(params *ServerParams) {
	params.grpcManager = o.grpcManager
}

func WithGrpcManager(manager GrpcManager) ServerOption {
	return &grpcManagerOption{manager}
}

// Deprecated: Use ServerParams instead
//nolint:gocognit // Long method names are okay because only generated code will call this, not humans.
func Server(ctx context.Context, name string, hl Manager, grpcHl GrpcManager, logger *logrus.Logger, promRegistry *prometheus.Registry) error {
	return NewServerParams(ctx, name,
		WithPkgLogger(),
		WithLogrusLogger(logger),
		WithRestManager(hl),
		WithGrpcManager(grpcHl),
		WithPrometheusRegistry(promRegistry)).Start()
}
