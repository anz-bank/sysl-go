package core

import (
	"context"

	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"github.com/go-chi/chi"
)

type HTTPManagerShim struct {
	libraryConfig          *config.LibraryConfig
	adminServerConfig      *config.CommonHTTPServerConfig
	publicServerConfig     *config.UpstreamConfig
	enabledHandlers        []handlerinitialiser.HandlerInitialiser
	addAdminHTTPMiddleware func(ctx context.Context, r chi.Router)
}

func NewHTTPManagerShim(libraryConfig *config.LibraryConfig, adminServerConfig *config.CommonHTTPServerConfig, publicServerConfig *config.UpstreamConfig, enabledHandlers []handlerinitialiser.HandlerInitialiser, addAdminHTTPMiddleware func(ctx context.Context, r chi.Router)) *HTTPManagerShim {
	return &HTTPManagerShim{
		libraryConfig:          libraryConfig,
		adminServerConfig:      adminServerConfig,
		publicServerConfig:     publicServerConfig,
		enabledHandlers:        enabledHandlers,
		addAdminHTTPMiddleware: addAdminHTTPMiddleware,
	}
}

func (m *HTTPManagerShim) EnabledHandlers() []handlerinitialiser.HandlerInitialiser {
	return m.enabledHandlers
}

func (m *HTTPManagerShim) LibraryConfig() *config.LibraryConfig {
	return m.libraryConfig
}

func (m *HTTPManagerShim) AdminServerConfig() *config.CommonHTTPServerConfig {
	return m.adminServerConfig
}

func (m *HTTPManagerShim) PublicServerConfig() *config.UpstreamConfig {
	return m.publicServerConfig
}

func (m *HTTPManagerShim) AddAdminHTTPMiddleware() func(ctx context.Context, r chi.Router) {
	return m.addAdminHTTPMiddleware
}
