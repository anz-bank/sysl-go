package config

import (
	"fmt"
	"io/ioutil"

	"github.com/anz-bank/sysl-go/validator"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type DefaultConfig struct {
	Library LibraryConfig `yaml:"library"`
	GenCode GenCodeConfig `yaml:"genCode"`
}

// LoadConfig reads and validates a configuration loaded from file.
// file: the path to the yaml-encoded config file
// defaultConfig: a pointer to the default config struct to populate
// customConfig: a pointer to the custom config struct to populate.
func LoadConfig(file string, defaultConfig *DefaultConfig, customConfig interface{}) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read config file error: %s", err)
	}

	if err = yaml.Unmarshal(b, customConfig); err != nil {
		return fmt.Errorf("unmarshal config file error: %s", err)
	}

	c := make(map[string]interface{})
	if err = yaml.Unmarshal(b, &c); err != nil {
		return fmt.Errorf("unmarshal config file error: %s", err)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:   nil,
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     defaultConfig,
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(c)
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
