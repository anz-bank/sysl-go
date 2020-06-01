package core

import (
	"context"
	"fmt"
	"net"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type GrpcManager interface {
	Interceptors() []grpc.UnaryServerInterceptor
	EnabledGrpcHandlers() []handlerinitialiser.GrpcHandlerInitialiser
	GrpcAdminServerConfig() *config.CommonServerConfig
	GrpcPublicServerConfig() *config.CommonServerConfig
}

func configurePublicGrpcServerListener(ctx context.Context, hl GrpcManager, logger *logrus.Logger) (func() error, error) {
	server, err := newGrpcServer(hl.GrpcPublicServerConfig(), logger, hl.Interceptors()...)
	if err != nil {
		return nil, err
	}

	// Not sure if it is possible to register multiple servers
	for _, h := range hl.EnabledGrpcHandlers() {
		h.RegisterServer(ctx, server)
	}

	var listenPublic func() error
	if len(hl.EnabledGrpcHandlers()) > 0 {
		listenPublic = prepareGrpcServerListener(logger, server, *hl.GrpcPublicServerConfig())
	}

	return listenPublic, nil
}

func makeGrpcListenFunc(server *grpc.Server, logger *logrus.Logger, cfg config.CommonServerConfig) func() error {
	return func() error {
		if cfg.TLS != nil {
			logger.Infof("TLS configuration present. Preparing to serve gRPC/HTTPS for address: %s:%d", cfg.HostName, cfg.Port)
		} else {
			logger.Infof("TLS configuration NOT present. Preparing to serve gRPC/HTTP for address: %s:%d", cfg.HostName, cfg.Port)
		}
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.HostName, cfg.Port))
		if err != nil {
			return err
		}
		return server.Serve(lis)
	}
}

func prepareGrpcServerListener(logger *logrus.Logger, server *grpc.Server, commonConfig config.CommonServerConfig) func() error {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(logger.WriterLevel(logrus.InfoLevel),
		logger.WriterLevel(logrus.WarnLevel), logger.WriterLevel(logrus.ErrorLevel)))

	listener := makeGrpcListenFunc(server, logger, commonConfig)
	logger.Infof("configured gRPC listener for address: %s:%d", commonConfig.HostName, commonConfig.Port)

	return listener
}

type interceptor struct {
	logger *logrus.Logger
}

func (i interceptor) unaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		entry := i.logger.WithField("traceid", "traceid")
		newCtx := common.LoggerToContext(ctx, i.logger, entry)

		return handler(newCtx, req)
	}
}

// NewGrpcServer creates a grpc.Server based on passed configuration
func newGrpcServer(cfg *config.CommonServerConfig, logger *logrus.Logger, interceptors ...grpc.UnaryServerInterceptor) (*grpc.Server, error) {
	opts, err := config.ExtractGrpcServerOptions(cfg)
	if err != nil {
		return nil, err
	}

	i := interceptor{
		logger: logger,
	}

	interceptors = append(interceptors, i.unaryInterceptor())

	opts = append(opts, grpcMiddleware.WithUnaryServerChain(
		interceptors...,
	))

	return grpc.NewServer(opts...), nil
}
