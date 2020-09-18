package core

import (
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
)

type HTTPManagerShim struct {
	libraryConfig      *config.LibraryConfig
	adminServerConfig  *config.CommonHTTPServerConfig
	publicServerConfig *config.CommonHTTPServerConfig
	enabledHandlers    []handlerinitialiser.HandlerInitialiser
}

func NewHTTPManagerShim(libraryConfig *config.LibraryConfig, adminServerConfig *config.CommonHTTPServerConfig, publicServerConfig *config.CommonHTTPServerConfig, enabledHandlers []handlerinitialiser.HandlerInitialiser) *HTTPManagerShim {
	return &HTTPManagerShim{
		libraryConfig:      libraryConfig,
		adminServerConfig:  adminServerConfig,
		publicServerConfig: publicServerConfig,
		enabledHandlers:    enabledHandlers,
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

func (m *HTTPManagerShim) PublicServerConfig() *config.CommonHTTPServerConfig {
	return m.publicServerConfig
}
