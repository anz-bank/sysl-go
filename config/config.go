package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ReadConfig reads from a single config file and populates both custom, library and genCode config structs
// cfgFile: path to config file
// config: a pointer to the custom config struct
func ReadConfig(cfgFile string, lib *LibraryConfig, gen *GenCodeConfig, config interface{}) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./")
		viper.AddConfigPath("../configs")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("fatal error config file: %s", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return fmt.Errorf("config file not found: %s", err)
		}
		return fmt.Errorf("config file was found but: %s", err)
	}

	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("unmarshal config error: %s", err)
	}

	if err := viper.UnmarshalKey("library", lib); err != nil {
		return fmt.Errorf("unmarshal library error: %s", err)
	}

	if err := viper.UnmarshalKey("genCode", gen); err != nil {
		return fmt.Errorf("unmarshal genCode error: %s", err)
	}

	return nil
}
