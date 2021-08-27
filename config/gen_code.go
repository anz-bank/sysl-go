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
	GRPC           GRPCServerConfig       `yaml:"grpc" mapstructure:"grpc"`
}

func (c *UpstreamConfig) Validate() error {
	// TODO: Actually validate
	return nil
}

func SetGenCodeConfigDefaults(prefix string, set func(key string, value interface{})) {
	set(prefix+"Upstream.ContextTimeout", "30s")
	set(prefix+"Upstream.HTTP.ReadTimeout", "30s")
	set(prefix+"Upstream.HTTP.WriteTimeout", "30s")
}
