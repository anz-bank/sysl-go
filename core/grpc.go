package core

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// Deprecated: prefer GrpcServerManager.
type GrpcManager interface {
	Interceptors() []grpc.UnaryServerInterceptor
	EnabledGrpcHandlers() []handlerinitialiser.GrpcHandlerInitialiser
	GrpcAdminServerConfig() *config.CommonServerConfig
	GrpcPublicServerConfig() *config.CommonServerConfig
}

type GrpcServerManager struct {
	GrpcServerOptions      []grpc.ServerOption
	EnabledGrpcHandlers    []handlerinitialiser.GrpcHandlerInitialiser
	GrpcPublicServerConfig *config.CommonServerConfig
}

func DefaultGrpcServerOptions(ctx context.Context, grpcPublicServerConfig *config.CommonServerConfig) ([]grpc.ServerOption, error) {
	opts, err := config.ExtractGrpcServerOptions(grpcPublicServerConfig)
	if err != nil {
		return nil, err
	}

	// Get a logger. This will EITHER return a custom logger if one was
	// previously prepared and put in the context, OR it will magically
	// create a new logger with default configuration that you're not
	// able to control.
	logger := log.From(ctx)
	// Inject the logger into the ctx so we can log when we're serving rpc calls.
	opts = append(opts, grpc.ChainUnaryInterceptor(makeLoggerInterceptor(logger)))

	opts = append(opts, grpc.ChainUnaryInterceptor(TraceidLogInterceptor))
	return opts, nil
}

func newGrpcServerManagerFromGrpcManager(hl GrpcManager) (*GrpcServerManager, error) {
	opts, err := extractGrpcServerOptionsFromGrpcManager(hl)
	if err != nil {
		return nil, err
	}
	return &GrpcServerManager{
		GrpcServerOptions:      opts,
		EnabledGrpcHandlers:    hl.EnabledGrpcHandlers(),
		GrpcPublicServerConfig: hl.GrpcPublicServerConfig(),
	}, nil
}

func extractGrpcServerOptionsFromGrpcManager(hl GrpcManager) ([]grpc.ServerOption, error) {
	opts, err := config.ExtractGrpcServerOptions(hl.GrpcPublicServerConfig())
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.ChainUnaryInterceptor(hl.Interceptors()...))
	opts = append(opts, grpc.ChainUnaryInterceptor(TraceidLogInterceptor)) // seems wrong to have this last in chain, but that was old behaviour.
	return opts, nil
}

func configurePublicGrpcServerListener(ctx context.Context, m GrpcServerManager) StoppableServer {
	server := grpc.NewServer(m.GrpcServerOptions...)
	// Not sure if it is possible to register multiple servers
	for _, h := range m.EnabledGrpcHandlers {
		h.RegisterServer(ctx, server)
	}

	return prepareGrpcServerListener(ctx, server, *m.GrpcPublicServerConfig)
}

type grpcServer struct {
	ctx    context.Context
	cfg    config.CommonServerConfig
	server *grpc.Server
}

func (s grpcServer) Start() error {
	if s.cfg.TLS != nil {
		log.Infof(s.ctx, "TLS configuration present. Preparing to serve gRPC/HTTPS for address: %s:%d", s.cfg.HostName, s.cfg.Port)
	} else {
		log.Infof(s.ctx, "TLS configuration NOT present. Preparing to serve gRPC/HTTP for address: %s:%d", s.cfg.HostName, s.cfg.Port)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.cfg.HostName, s.cfg.Port))
	if err != nil {
		return err
	}
	return s.server.Serve(lis)
}

func (s grpcServer) GracefulStop() error {
	s.server.GracefulStop()
	return nil
}

func (s grpcServer) Stop() error {
	s.server.Stop()
	return nil
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

func prepareGrpcServerListener(ctx context.Context, server *grpc.Server, commonConfig config.CommonServerConfig) StoppableServer {
	grpclog.SetLoggerV2(
		grpclog.NewLoggerV2(
			&logWriterInfo{logger: log.From(ctx)},
			&logWriterInfo{logger: log.From(ctx)},
			&logWriterError{logger: log.From(ctx)}))

	log.Infof(ctx, "configured gRPC listener for address: %s:%d", commonConfig.HostName, commonConfig.Port)
	return grpcServer{ctx: ctx, cfg: commonConfig, server: server}
}

func makeLoggerInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = log.WithLogger(logger).Onto(ctx)
		return handler(ctx, req)
	}
}

func TraceidLogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(log.With("traceid", "traceid").Onto(ctx), req)
}
