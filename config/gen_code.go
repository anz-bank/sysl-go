package config

import (
	"time"
)

// GenCodeConfig struct.
type GenCodeConfig struct {
	Upstream   UpstreamConfig `yaml:"upstream" mapstructure:"upstream"`
	Downstream interface{}    `yaml:"downstream" mapstructure:"downstream"`
}

// UpstreamConfig struct.
type UpstreamConfig struct {
	ContextTimeout time.Duration          `yaml:"contextTimeout" mapstructure:"contextTimeout" validate:"nonnil"`
	HTTP           CommonHTTPServerConfig `yaml:"http" mapstructure:"http"`
	GRPC           CommonServerConfig     `yaml:"grpc" mapstructure:"grpc"`
}

func (c *UpstreamConfig) Validate() error {
	// TODO: Actually validate
	return nil
}
