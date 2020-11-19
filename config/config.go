package config

import (
	"context"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/config/envvar"
	"github.com/anz-bank/sysl-go/validator"
	"github.com/go-chi/chi"
)

type DefaultConfig struct {
	Library LibraryConfig `yaml:"library" mapstructure:"library"`

	// config used for setting up the sysl-go admin server
	Admin   *AdminConfig  `yaml:"admin" mapstructure:"admin"`
	GenCode GenCodeConfig `yaml:"genCode" mapstructure:"genCode"`

	// development config can be used to set some config options only appropriate for dev/test environments.
	Development *DevelopmentConfig `yaml:"development" mapstructure:"development"`
}

// LoadConfig reads and validates a configuration loaded from file.
// file: the path to the yaml-encoded config file
// defaultConfig: a pointer to the default config struct to populate
// customConfig: a pointer to the custom config struct to populate.
func LoadConfig(file string, defaultConfig *DefaultConfig, customConfig interface{}) error {
	b := envvar.NewConfigReaderBuilder().WithConfigFile(file)
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

func NewCallbackV2(
	config *GenCodeConfig,
	downstreamTimeOut time.Duration,
	mapError func(ctx context.Context, err error) *common.HTTPError,
	addMiddleware func(ctx context.Context, r chi.Router),
) common.Callback {
	// construct the rest configuration (aka. gen callback)
	return common.Callback{
		UpstreamTimeout:   config.Upstream.ContextTimeout,
		DownstreamTimeout: downstreamTimeOut,
		RouterBasePath:    config.Upstream.HTTP.BasePath,
		UpstreamConfig:    &config.Upstream,
		MapErrorFunc:      mapError,
		AddMiddlewareFunc: addMiddleware,
	}
}

// NewCallback is deprecated, prefer NewCallbackV2.
func NewCallback(
	config *GenCodeConfig,
	downstreamTimeOut time.Duration,
	mapError func(ctx context.Context, err error) *common.HTTPError,
) common.Callback {
	return NewCallbackV2(
		config,
		downstreamTimeOut,
		mapError,
		nil,
	)
}
