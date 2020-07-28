package core

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type GrpcManager interface {
	Interceptors() []grpc.UnaryServerInterceptor
	EnabledGrpcHandlers() []handlerinitialiser.GrpcHandlerInitialiser
	GrpcAdminServerConfig() *config.CommonServerConfig
	GrpcPublicServerConfig() *config.CommonServerConfig
}

func configurePublicGrpcServerListener(ctx context.Context, hl GrpcManager) (func() error, error) {
	server, err := newGrpcServer(hl.GrpcPublicServerConfig(), hl.Interceptors()...)
	if err != nil {
		return nil, err
	}

	// Not sure if it is possible to register multiple servers
	for _, h := range hl.EnabledGrpcHandlers() {
		h.RegisterServer(ctx, server)
	}

	var listenPublic func() error
	if len(hl.EnabledGrpcHandlers()) > 0 {
		listenPublic = prepareGrpcServerListener(ctx, server, *hl.GrpcPublicServerConfig())
	}

	return listenPublic, nil
}

func makeGrpcListenFunc(ctx context.Context, server *grpc.Server, cfg config.CommonServerConfig) func() error {
	return func() error {
		if cfg.TLS != nil {
			log.Infof(ctx, "TLS configuration present. Preparing to serve gRPC/HTTPS for address: %s:%d", cfg.HostName, cfg.Port)
		} else {
			log.Infof(ctx, "TLS configuration NOT present. Preparing to serve gRPC/HTTP for address: %s:%d", cfg.HostName, cfg.Port)
		}
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.HostName, cfg.Port))
		if err != nil {
			return err
		}
		return server.Serve(lis)
	}
}

type logWriterInfo struct {
	logger log.Logger
}

func (lw *logWriterInfo) Write(p []byte) (n int, err error) {
	lw.logger.Info(string(p))
	return len(p), nil
}

type logWriterError struct {
	logger log.Logger
}

func (lw *logWriterError) Write(p []byte) (n int, err error) {
	lw.logger.Error(errors.New(string(p)))
	return len(p), nil
}

func prepareGrpcServerListener(ctx context.Context, server *grpc.Server, commonConfig config.CommonServerConfig) func() error {
	grpclog.SetLoggerV2(
		grpclog.NewLoggerV2(
			&logWriterInfo{logger: log.From(ctx)},
			&logWriterInfo{logger: log.From(ctx)},
			&logWriterError{logger: log.From(ctx)}))

	listener := makeGrpcListenFunc(ctx, server, commonConfig)
	log.Infof(ctx, "configured gRPC listener for address: %s:%d", commonConfig.HostName, commonConfig.Port)

	return listener
}

func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(log.With("traceid", "traceid").Onto(ctx), req)
}

// NewGrpcServer creates a grpc.Server based on passed configuration.
func newGrpcServer(cfg *config.CommonServerConfig, interceptors ...grpc.UnaryServerInterceptor) (*grpc.Server, error) {
	opts, err := config.ExtractGrpcServerOptions(cfg)
	if err != nil {
		return nil, err
	}

	interceptors = append(interceptors, unaryInterceptor)
	opts = append(opts, grpcMiddleware.WithUnaryServerChain(
		interceptors...,
	))

	return grpc.NewServer(opts...), nil
}
