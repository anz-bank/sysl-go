package config

import (
	"fmt"
	"io/ioutil"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

type DefaultConfig struct {
	Library LibraryConfig `yaml:"library"`
	GenCode GenCodeConfig `yaml:"genCode"`
}

// ReadConfig reads from a single config file and populates both custom, library and genCode config structs
// cfgFile: path to config file
// config: a pointer to the custom config struct
func ReadConfig(cfgFile string, defaultConfig *DefaultConfig, customConfig interface{}) error {
	b, err := ioutil.ReadFile(cfgFile)
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

	return decoder.Decode(c)
}
