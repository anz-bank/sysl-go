package config

import (
	"time"
)

// GenCodeConfig struct
type GenCodeConfig struct {
	Upstream   Server             `yaml:"upstream"`
	Downstream ProviderTransports `yaml:"downstream"`
}

// Server struct
type Server struct {
	ContextTimeout time.Duration          `yaml:"contextTimeout" validate:"nonnil"`
	HTTP           CommonHTTPServerConfig `yaml:"http"`
	GRPC           CommonServerConfig     `yaml:"grpc"`
}

// ProviderTransports struct
type ProviderTransports struct {
	ContextTimeout time.Duration        `yaml:"contextTimeout"`
	Fenergo        CommonDownstreamData `yaml:"fenergo"` // FIXME: grpc/http
	Qas            CommonDownstreamData `yaml:"qas"`     // FIXME
}
