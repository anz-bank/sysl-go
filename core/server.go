package core

// MARKED TO IGNORE COVERAGE

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type ServerParams struct {
	Ctx                context.Context
	Name               string
	logger             *logrus.Logger
	restManager        Manager
	grpcManager        GrpcManager
	prometheusRegistry *prometheus.Registry
}

//nolint:gocognit // Long method names are okay because only generated code will call this, not humans.
func NewServerParams(ctx context.Context, name string, opts ...ServerOption) *ServerParams {
	params := &ServerParams{Ctx: ctx, Name: name}
	for _, o := range opts {
		o.apply(params)
	}
	return params
}

//nolint:gocognit // Long method are okay because only generated code will call this, not humans.
func (params *ServerParams) Start() error {
	mWare := prepareMiddleware(params.Name, params.logger, params.prometheusRegistry)

	var restIsRunning, grpcIsRunning bool

	// Run the REST server
	var listenAdmin func() error
	if params.restManager != nil && params.restManager.AdminServerConfig() != nil {
		var err error
		listenAdmin, err = configureAdminServerListener(params.restManager, params.logger, params.prometheusRegistry, mWare.admin)
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
		listenPublic, err = configurePublicServerListener(params.Ctx, params.restManager, params.logger, mWare.public)
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
		listenPublicGrpc, err = configurePublicGrpcServerListener(params.Ctx, params.grpcManager, params.logger)
		if err != nil {
			return err
		}

		grpcIsRunning = true
	} else {
		listenPublicGrpc = func() error { select {} }
	}

	// Panic if REST&gRPC are not running
	if !restIsRunning && !grpcIsRunning {
		panic("Both servers are set to nil")
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

type loggerOption struct {
	logger *logrus.Logger
}

func (o *loggerOption) apply(params *ServerParams) {
	params.logger = o.logger
}

func WithLogrusLogger(logger *logrus.Logger) ServerOption {
	return &loggerOption{logger}
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

//nolint:gocognit // Long method names are okay because only generated code will call this, not humans.
func Server(ctx context.Context, name string, hl Manager, grpcHl GrpcManager, logger *logrus.Logger, promRegistry *prometheus.Registry) error {
	return NewServerParams(ctx, name,
		WithLogrusLogger(logger),
		WithRestManager(hl),
		WithGrpcManager(grpcHl),
		WithPrometheusRegistry(promRegistry)).Start()
}
