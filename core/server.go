package core

// MARKED TO IGNORE COVERAGE

import (
	"context"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

//nolint:gocognit // Long method names are okay because only generated code will call this, not humans.
func Server(ctx context.Context, name string, libraryConfig *config.LibraryConfig,
	hl RestManager, grpcHl GrpcManager,
	logger *logrus.Logger, promRegistry *prometheus.Registry) error {

	mWare := prepareMiddleware(name, logger, promRegistry)

	var restIsRunning, grpcIsRunning bool

	// Run the admin server
	var listenAdmin func() error
	if libraryConfig != nil && libraryConfig.AdminServer != nil {
		var err error
		handlers := make([]handlerinitialiser.HandlerInitialiser, 0)
		if hl != nil {
			for _, h := range hl.EnabledHandlers() {
				handlers = append(handlers, h)
			}
		}
		if grpcHl != nil {
			for _, h := range grpcHl.EnabledGrpcHandlers() {
				handlers = append(handlers, h)
			}
		}
		listenAdmin, err = configureAdminServerListener(libraryConfig, handlers, logger, promRegistry, mWare.admin)
		if err != nil {
			return err
		}
	} else {
		// set up a dummy listener which will never exit if admin disabled
		listenAdmin = func() error { select {} }
	}

	// Run the REST server
	var listenPublic func() error
	if hl != nil && hl.PublicServerConfig() != nil {
		var err error
		listenPublic, err = configurePublicServerListener(ctx, hl, logger, mWare.public)
		if err != nil {
			return err
		}
		restIsRunning = true
	} else {
		listenPublic = func() error { select {} }
	}

	// Run the gRPC server
	var listenPublicGrpc func() error
	if grpcHl != nil && grpcHl.GrpcPublicServerConfig() != nil {
		var err error
		listenPublicGrpc, err = configurePublicGrpcServerListener(ctx, grpcHl, logger)
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
