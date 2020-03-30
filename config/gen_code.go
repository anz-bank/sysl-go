package config

import (
	"time"
)

// GenCodeConfig struct
type GenCodeConfig struct {
	Upstream   UpstreamConfig `yaml:"upstream"`
	Downstream interface{}    `yaml:"downstream"`
}

// UpstreamConfig struct
type UpstreamConfig struct {
	ContextTimeout time.Duration          `yaml:"contextTimeout" validate:"nonnil"`
	HTTP           CommonHTTPServerConfig `yaml:"http"`
	GRPC           CommonServerConfig     `yaml:"grpc"`
}
