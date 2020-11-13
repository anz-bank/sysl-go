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

// WithStrictMode controls if ConfigReader.Unmarshal handles unknown
// keys. If strict mode is false (the default), config keys with no
// corresponding config field are ignored. If strict mode is true,
// any config key with no corresponding config field will be regarded
// as a decoding error that will cause Unmarshal to return an error.
//
// Also, optionally, a list of keys to ignore and exclude from strict
// mode checking can be provided. Beware, there's some subtleties
// to how ignored keys must be named, see the comments inside
// configreaderimpl.go for details.
func (b ConfigReaderBuilder) WithStrictMode(strict bool, ignoredKeys ...string) ConfigReaderBuilder {
	b.evarReader.strictMode = strict
	b.evarReader.strictModeIgnoredKeys = ignoredKeys
	return b
}

// Build Builds and returns the ConfigReader.
func (b ConfigReaderBuilder) Build() ConfigReader {
	if err := b.evarReader.envVars.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}
	return b.evarReader
}
