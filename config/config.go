package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// ReadConfig reads from a single config file and populates both custom, library and genCode config structs
// cfgFile: path to config file
// config: a pointer to the custom config struct
func ReadConfig(cfgFile string, lib *LibraryConfig, gen *GenCodeConfig, config interface{}) {
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
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("config file not found: %s \n", err))
		} else {
			panic(fmt.Errorf("config file was found but: %s \n", err))
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		log.Fatal("Unmarshal error:", err)
	}

	if err := viper.UnmarshalKey("library", lib); err != nil {
		log.Fatal("Unmarshal error:", err)
	}

	if err := viper.UnmarshalKey("genCode", gen); err != nil {
		log.Fatal("Unmarshal error:", err)
	}

	return
}
