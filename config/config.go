package config

import (
	"context"

	"github.com/anz-bank/sysl-go/validator"
)

type defaultConfigKey struct{}

type DefaultConfig struct {
	Library LibraryConfig `yaml:"library" mapstructure:"library"`

	// config used for setting up the sysl-go admin server
	Admin   *AdminConfig  `yaml:"admin" mapstructure:"admin"`
	GenCode GenCodeConfig `yaml:"genCode" mapstructure:"genCode"`

	// development config can be used to set some config options only appropriate for dev/test environments.
	Development *DevelopmentConfig `yaml:"development" mapstructure:"development"`
}

// GetDefaultConfig retrieves the externally-provided config from the context.
// The default config is injected into the server context during bootstrapping and can therefore
// be called from anywhere within the running application.
func GetDefaultConfig(ctx context.Context) *DefaultConfig {
	m, _ := ctx.Value(defaultConfigKey{}).(*DefaultConfig)
	return m
}

// PutDefaultConfig puts the externally-provided config into the given context, returning the new context.
func PutDefaultConfig(ctx context.Context, config *DefaultConfig) context.Context {
	return context.WithValue(ctx, defaultConfigKey{}, config)
}

// LoadConfig reads and validates a configuration loaded from file.
// file: the path to the yaml-encoded config file
// defaultConfig: a pointer to the default config struct to populate
// customConfig: a pointer to the custom config struct to populate.
func LoadConfig(file string, defaultConfig *DefaultConfig, customConfig interface{}) error {
	b := NewConfigReaderBuilder().WithConfigFile(file).WithDefaults(SetDefaults)
	err := b.Build().Unmarshal(defaultConfig)
	if err != nil {
		return err
	}
	err = b.Build().Unmarshal(customConfig)
	if err != nil {
		return err
	}
	err = validator.Validate(defaultConfig)
	if err != nil {
		return err
	}

	err = validator.Validate(customConfig)
	if err != nil {
		return err
	}

	return err
}

func SetDefaults(setter func(key string, value interface{})) {
	SetLibraryConfigDefaults("Library.", setter)
	SetGenCodeConfigDefaults("GenCode.", setter)
}
