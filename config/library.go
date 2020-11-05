package config

import (
	"encoding/base64"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/jwtauth"
	"github.com/anz-bank/sysl-go/validator"
	"github.com/sirupsen/logrus"
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
	Format       string        `yaml:"format" mapstructure:"format" validate:"nonnil,oneof=color json text"`
	Splunk       *SplunkConfig `yaml:"splunk" mapstructure:"splunk"`
	Level        logrus.Level  `yaml:"level" mapstructure:"level" validate:"nonnil"`
	ReportCaller bool          `yaml:"caller" mapstructure:"caller"`
}

// SplunkConfig struct.
type SplunkConfig struct {
	TokenBase64 common.SensitiveString `yaml:"tokenBase64" mapstructure:"tokenBase64" validate:"nonnil,base64"`
	Index       string                 `yaml:"index" mapstructure:"index" validate:"nonnil"`
	Target      string                 `yaml:"target" mapstructure:"target" validate:"nonnil,url"`
	Source      string                 `yaml:"source" mapstructure:"source" validate:"nonnil"`
	SourceType  string                 `yaml:"sourceType" mapstructure:"sourceType" validate:"nonnil"`
}

// AuthenticationConfig struct.
type AuthenticationConfig struct {
	JWTAuth *jwtauth.Config `yaml:"jwtauth" mapstructure:"jwtauth"`
}

func (s *SplunkConfig) Token() string {
	b, _ := base64.StdEncoding.DecodeString(s.TokenBase64.Value())

	return string(b)
}

func (c *LibraryConfig) Validate() error {
	// existing validation
	if err := validator.Validate(c); err != nil {
		return err
	}

	return nil
}
