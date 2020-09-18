package core

import (
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"google.golang.org/grpc"
)

type GRPCManagerShim struct {
	interceptors       []grpc.UnaryServerInterceptor
	adminServerConfig  *config.CommonServerConfig
	publicServerConfig *config.CommonServerConfig
	enabledHandlers    []handlerinitialiser.GrpcHandlerInitialiser
}

func NewGRPCManagerShim(interceptors []grpc.UnaryServerInterceptor, adminServerConfig *config.CommonServerConfig, publicServerConfig *config.CommonServerConfig, enabledHandlers []handlerinitialiser.GrpcHandlerInitialiser) *GRPCManagerShim {
	return &GRPCManagerShim{
		interceptors:       interceptors,
		adminServerConfig:  adminServerConfig,
		publicServerConfig: publicServerConfig,
		enabledHandlers:    enabledHandlers,
	}
}

func (m *GRPCManagerShim) Interceptors() []grpc.UnaryServerInterceptor {
	return m.interceptors
}

func (m *GRPCManagerShim) GrpcAdminServerConfig() *config.CommonServerConfig {
	return m.adminServerConfig
}

func (m *GRPCManagerShim) GrpcPublicServerConfig() *config.CommonServerConfig {
	return m.publicServerConfig
}

func (m *GRPCManagerShim) EnabledGrpcHandlers() []handlerinitialiser.GrpcHandlerInitialiser {
	return m.enabledHandlers
}
