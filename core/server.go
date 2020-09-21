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

func NewServerParams(ctx context.Context, name string, opts ...ServerOption) *ServerParams {
	params := &ServerParams{Ctx: ctx, Name: name}
	for _, o := range opts {
		o.apply(params)
	}
	return params
}

// initialise the logger
// sysl-go always uses a pkg logger internally. if custom code passes in a logrus logger, a
// mechanism which is deprecated, then a hook is added to the internal pkg logger that forwards
// logged events to the provided logrus logger.
// sysl-go can be requested to log in a verbose manner. logger in a verbose manner logs additional
// details within log events where appropriate. the mechanism to set this verbose manner is to
// either have a sufficiently high logrus log level or the verbose mode set against the pkg logger.
func InitialiseLogging(ctx context.Context, configs []log.Config, logrusLogger *logrus.Logger) context.Context {
	verboseLogging := false
	if logrusLogger != nil {
		// set an empty io writter against pkg logger
		// pkg logger just becomes a proxy that forwards all logs to logrus
		configs = append(configs, log.SetOutput(&emptyWriter{}))
		configs = append(configs, log.AddHooks(&logrusHook{logrusLogger}))
		configs = append(configs, log.SetLogCaller(logrusLogger.ReportCaller))
		ctx = common.LoggerToContext(ctx, logrusLogger, nil)
		verboseLogging = logrusLogger.Level >= logrus.DebugLevel
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
	return logconfig.SetVerboseLogging(ctx, verboseLogging)
}

//nolint:funlen // Long method are okay because only generated code will call this, not humans.
func (params *ServerParams) Start() error {
	ctx := params.Ctx

	// initialise the logger
	ctx = InitialiseLogging(ctx, params.pkgLoggerConfigs, params.logrusLogger)

	// prepare the middleware
	mWare := prepareMiddleware(ctx, params.Name, params.prometheusRegistry)

	var restIsRunning, grpcIsRunning bool

	listeners := make([]func() error, 0)

	// Make the listener function for the REST Admin server
	if params.restManager != nil && params.restManager.AdminServerConfig() != nil {
		log.Info(ctx, "found AdminServerConfig for REST")
		listenAdmin, err := configureAdminServerListener(ctx, params.restManager, params.prometheusRegistry, mWare.admin)
		if err != nil {
			return err
		}
		listeners = append(listeners, listenAdmin)
	} else {
		log.Info(ctx, "no AdminServerConfig for REST was found")
	}

	// Make the listener function for the REST Public server
	if params.restManager != nil && params.restManager.PublicServerConfig() != nil {
		log.Info(ctx, "found PublicServerConfig for REST")
		listenPublic, err := configurePublicServerListener(ctx, params.restManager, mWare.public)
		if err != nil {
			return err
		}
		listeners = append(listeners, listenPublic)
		restIsRunning = true
	} else {
		log.Info(ctx, "no PublicServerConfig for REST was found")
	}

	// Make the listener function for the gRPC Public server.
	if params.grpcManager != nil && params.grpcManager.GrpcPublicServerConfig() != nil {
		log.Info(ctx, "found GrpcPublicServerConfig for gRPC")
		listenPublicGrpc, err := configurePublicGrpcServerListener(ctx, params.grpcManager)
		if err != nil {
			return err
		}
		listeners = append(listeners, listenPublicGrpc)
		grpcIsRunning = true
	} else {
		log.Info(ctx, "no GrpcPublicServerConfig for gRPC was found")
	}

	// Refuse to start and panic if neither of the public servers are enabled.
	if !restIsRunning && !grpcIsRunning {
		err := errors.New("REST and gRPC servers cannot both be nil")
		log.Error(ctx, err)
		panic(err)
	}

	// Run all listeners for all configured servers and block until the first one terminates.
	errChan := make(chan error, 1)
	for i := range listeners {
		listener := listeners[i]
		go func() {
			errChan <- listener()
		}()
	}
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

// Deprecated: Use WithPkgLogger instead.
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
