package core

// MARKED TO IGNORE COVERAGE

import (
	"context"

	"github.com/anz-bank/sysl-go/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type ServerParams struct {
	Ctx                context.Context
	Name               string
	Config             *config.DefaultConfig
	logger             *logrus.Logger
	restManager        RestManager
	grpcManager        GrpcManager
	prometheusRegistry *prometheus.Registry
}

func NewServerParams(ctx context.Context, name string, config *config.DefaultConfig) *ServerParams {
	return &ServerParams{Ctx: ctx, Name: name, Config: config}
}

//nolint:gocognit // Long method names are okay because only generated code will call this, not humans.
func (params *ServerParams) Start(opts ...ServerOption) error {
	for _, o := range opts {
		if err := o.Apply(params); err != nil {
			return err
		}
	}
	return Server(params)
}

type ServerOption interface {
	Apply(params *ServerParams) error
}

type restManagerOption struct {
	restManager RestManager
}

func (o *restManagerOption) Apply(params *ServerParams) error {
	params.restManager = o.restManager
	return nil
}

func WithRestManager(manager RestManager) ServerOption {
	return &restManagerOption{manager}
}

type loggerOption struct {
	logger *logrus.Logger
}

func (o *loggerOption) Apply(params *ServerParams) error {
	params.logger = o.logger
	return nil
}

func WithLogger(logger *logrus.Logger) ServerOption {
	return &loggerOption{logger}
}

func WithPrometheusRegistry(prometheusRegistry *prometheus.Registry) ServerOption {
	return &prometheusRegistryOption{prometheusRegistry}
}

type prometheusRegistryOption struct {
	prometheusRegistry *prometheus.Registry
}

func (o *prometheusRegistryOption) Apply(params *ServerParams) error {
	params.prometheusRegistry = o.prometheusRegistry
	return nil
}

type grpcManagerOption struct {
	grpcManager GrpcManager
}

func (o *grpcManagerOption) Apply(params *ServerParams) error {
	params.grpcManager = o.grpcManager
	return nil
}

func WithGrpcManager(manager GrpcManager) ServerOption {
	return &grpcManagerOption{manager}
}

//nolint:gocognit // Long method names are okay because only generated code will call this, not humans.
func Server(params *ServerParams) error {
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
