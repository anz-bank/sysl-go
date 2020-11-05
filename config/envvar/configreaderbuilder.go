package envvar

import (
	"log"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// ConfigReaderBuilder exposes the builder api for configReaderImpl.
// Use NewConfigReaderBuilder() and AttachEnvPrefix() to build a ConfigReaderBuilder. Follow it up one or more calls
// to WithConfigFile() and/or WithConfigName() and finally use Build() to build the configReaderImpl.
type ConfigReaderBuilder struct {
	evarReader configReaderImpl
}

// NewConfigReaderBuilder builds a new ConfigReaderBuilder.
func NewConfigReaderBuilder() ConfigReaderBuilder {
	b := ConfigReaderBuilder{
		evarReader: configReaderImpl{
			envVars: viper.New(),
		},
	}
	return b
}

// AttachEnvPrefix attaches appName as prefix.
func (b ConfigReaderBuilder) AttachEnvPrefix(appName string) ConfigReaderBuilder {
	b.evarReader.envVars.SetEnvPrefix(appName)
	b.evarReader.envVars.AutomaticEnv()
	b.evarReader.envVars.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return b
}

// WithConfigFile attaches the passed config file.
func (b ConfigReaderBuilder) WithConfigFile(configFile string) ConfigReaderBuilder {
	b.evarReader.envVars.SetConfigFile(configFile)
	if err := b.evarReader.envVars.MergeInConfig(); err != nil {
		log.Fatalln(err)
	}
	return b
}

// WithConfigName attaches the passed config path and name.
func (b ConfigReaderBuilder) WithConfigName(configName string, configPath ...string) ConfigReaderBuilder {
	b.evarReader.envVars.SetConfigName(configName)
	for _, path := range configPath {
		b.evarReader.envVars.AddConfigPath(path)
	}
	if err := b.evarReader.envVars.MergeInConfig(); err != nil {
		log.Fatalln(err)
	}
	return b
}

// WithFs attaches the file system to use.
func (b ConfigReaderBuilder) WithFs(fs afero.Fs) ConfigReaderBuilder {
	b.evarReader.envVars.SetFs(fs)
	return b
}

// Build Builds and returns the ConfigReader.
func (b ConfigReaderBuilder) Build() ConfigReader {
	if err := b.evarReader.envVars.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}
	return b.evarReader
}
