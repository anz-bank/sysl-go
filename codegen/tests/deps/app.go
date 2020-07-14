// Code generated by sysl DO NOT EDIT.
package deps

import (
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
)

// DownstreamClients for Deps
type DownstreamClients struct {
}

// BuildRestHandlerInitialiser ...
func BuildRestHandlerInitialiser(serviceInterface ServiceInterface, callback core.RestGenCallback, downstream *DownstreamClients) handlerinitialiser.HandlerInitialiser {
	serviceHandler := NewServiceHandler(callback, &serviceInterface)
	serviceRouter := NewServiceRouter(callback, serviceHandler)
	return serviceRouter
}

// BuildDownstreamClients ...
func BuildDownstreamClients(cfg *config.DefaultConfig) (*DownstreamClients, error) {
	var err error = nil

	return &DownstreamClients{}, err
}

// NewDefaultConfig ...
func NewDefaultConfig() config.DefaultConfig {
	return config.DefaultConfig{
		Library: config.LibraryConfig{},
		GenCode: config.GenCodeConfig{Downstream: &DownstreamConfig{}},
	}
}
