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
	Trace          TraceConfig           `yaml:"trace" mapstructure:"trace"`
}

type AdminConfig struct {
	ContextTimeout time.Duration          `yaml:"contextTimeout" mapstructure:"contextTimeout" validate:"nonnil"`
	HTTP           CommonHTTPServerConfig `yaml:"http" mapstructure:"http"`
}

// LogConfig struct.
type LogConfig struct {
	Format       string    `yaml:"format" mapstructure:"format" validate:"oneof=color json text"` // Deprecated: Use Hooks#Logger
	Level        log.Level `yaml:"level" mapstructure:"level" validate:"nonnil"`
	ReportCaller bool      `yaml:"caller" mapstructure:"caller"` // Deprecated: Use Hooks#Logger

	// LogPayload logs the contents of request and response objects.
	LogPayload bool `yaml:"logPayload" mapstructure:"logPayload"`
}

// AuthenticationConfig struct.
type AuthenticationConfig struct {
	JWTAuth *jwtauth.Config `yaml:"jwtauth" mapstructure:"jwtauth"`
}

// TraceConfig struct.
type TraceConfig struct {
	IncomingHeaderForID string `yaml:"incomingHeaderForID" mapstructure:"incomingHeaderForID"`
}

func (c *LibraryConfig) Validate() error {
	// existing validation
	if err := validator.Validate(c); err != nil {
		return err
	}

	return nil
}

func SetLibraryConfigDefaults(prefix string, set func(key string, value interface{})) {
	set(prefix+"Log.Format", "text")
	set(prefix+"Log.Level", log.InfoLevel)
}
