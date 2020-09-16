package core

import (
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
)

type HTTPManager struct {
	libraryConfig      *config.LibraryConfig
	adminServerConfig  *config.CommonHTTPServerConfig
	publicServerConfig *config.CommonHTTPServerConfig
	enabledHandlers    []handlerinitialiser.HandlerInitialiser
}

func NewHTTPManager(libraryConfig *config.LibraryConfig, adminServerConfig *config.CommonHTTPServerConfig, publicServerConfig *config.CommonHTTPServerConfig, enabledHandlers []handlerinitialiser.HandlerInitialiser) *HTTPManager {
	return &HTTPManager{
		libraryConfig:      libraryConfig,
		adminServerConfig:  adminServerConfig,
		publicServerConfig: publicServerConfig,
		enabledHandlers:    enabledHandlers,
	}
}

func (h *HTTPManager) EnabledHandlers() []handlerinitialiser.HandlerInitialiser {
	return h.enabledHandlers
}

func (h *HTTPManager) LibraryConfig() *config.LibraryConfig {
	return h.libraryConfig
}

func (h *HTTPManager) AdminServerConfig() *config.CommonHTTPServerConfig {
	return h.adminServerConfig
}

func (h *HTTPManager) PublicServerConfig() *config.CommonHTTPServerConfig {
	return h.publicServerConfig
}
