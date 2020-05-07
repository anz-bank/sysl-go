package core

import (
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
)

type RestManager interface {
	EnabledHandlers() []handlerinitialiser.HandlerInitialiser
	LibraryConfig() *config.LibraryConfig
	AdminServerConfig() *config.CommonHTTPServerConfig
	PublicServerConfig() *config.CommonHTTPServerConfig
}
