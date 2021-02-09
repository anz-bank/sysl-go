package config

import (
	"time"

	"github.com/anz-bank/sysl-go/log"

	"github.com/anz-bank/sysl-go/jwtauth"
	"github.com/anz-bank/sysl-go/validator"
)

// LibraryConfig struct.
type LibraryConfig struct {
	Log            LogConfig             `yaml:"log" mapstructure:"log"`
	Profiling      bool                  `yaml:"profiling" mapstructure:"profiling"`
	Health         bool                  `yaml:"health" mapstructure:"health"`
	Authentication *AuthenticationConfig `yaml:"authentication" mapstructure:"authentication"`
}

type AdminConfig struct {
	ContextTimeout time.Duration          `yaml:"contextTimeout" mapstructure:"contextTimeout" validate:"nonnil"`
	HTTP           CommonHTTPServerConfig `yaml:"http" mapstructure:"http"`
}

// LogConfig struct.
type LogConfig struct {
	Format       string    `yaml:"format" mapstructure:"format" validate:"nonnil,oneof=color json text"` // Deprecated: Use Hooks#Logger
	Level        log.Level `yaml:"level" mapstructure:"level" validate:"nonnil"`
	ReportCaller bool      `yaml:"caller" mapstructure:"caller"` // Deprecated: Use Hooks#Logger
}

// AuthenticationConfig struct.
type AuthenticationConfig struct {
	JWTAuth *jwtauth.Config `yaml:"jwtauth" mapstructure:"jwtauth"`
}

func (c *LibraryConfig) Validate() error {
	// existing validation
	if err := validator.Validate(c); err != nil {
		return err
	}

	return nil
}
