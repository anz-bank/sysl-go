package config

import (
	"fmt"
	rawlog "log"
	"reflect"
	"strings"

	"github.com/anz-bank/sysl-go/log"

	"github.com/anz-bank/sysl-go/jsontime"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// configReaderImpl exposes a wrapper api for viper.
type configReaderImpl struct {
	envVars               *viper.Viper
	strictMode            bool
	strictModeIgnoredKeys []string
}

// Get returns an interface{}.
// For a specific value use one of the Get____ methods.
func (m configReaderImpl) Get(key string) (interface{}, error) {
	val := m.envVars.Get(key)
	if val == nil {
		return nil, NilValueError{fmt.Sprintf("%s: key value is nil", key)}
	}
	return m.envVars.Get(key), nil
}

func (m configReaderImpl) buildValueConversionErr(key, valueType string) error {
	return errors.Wrap(ValueConversionError{
		fmt.Sprintf("%s: key value is incompatible with %s", key, valueType)},
		"value conversion failed")
}

// GetString retrieves the associated key value as a string.
func (m configReaderImpl) GetString(key string) (string, error) {
	val, err := m.Get(key)
	if err != nil {
		return "", err
	}
	str, err := cast.ToStringE(val)
	if err != nil {
		return "", m.buildValueConversionErr(key, "string")
	}
	return str, nil
}

// Unmarshal deserializes the loaded cofig into a struct.
func (m configReaderImpl) Unmarshal(config interface{}) error {
	opts := []viper.DecoderConfigOption{}

	metadata := &mapstructure.Metadata{}

	// If "strict mode" is set then regard unused config keys
	// -- that is, config keys that don't correspond to any known
	// config field -- as errors. Unless we are configured to explicitly
	// ignore them.
	if m.strictMode {
		opts = append(opts, func(cfg *mapstructure.DecoderConfig) {
			// Instead of setting cfg.ErrorUnused we instead collect Metadata,
			// as this gives us a way to suppress and ignore some unused
			// config keys.
			cfg.Metadata = metadata
		})
	}

	decodeHook := viper.DecodeHook(makeDefaultDecodeHook())

	opts = append(opts, decodeHook)

	if err := m.envVars.Unmarshal(config, opts...); err != nil {
		return fmt.Errorf("Unable to decode into struct %s", err)
	}

	if m.strictMode {
		return m.validateNoUnusedKeys(metadata)
	}

	return nil
}

func (m configReaderImpl) validateNoUnusedKeys(metadata *mapstructure.Metadata) error {
	// Filter away any unused keys that should be ignored.
	// Beware: for nested keys, mapstructure will not
	// necessarily report the full key in metadata.Unused:
	// For example, if we unmarshal into a config structure
	// with no "fizz" key, and there is a nested key named
	// "fizz.buzz" in the input, then mapstructure will report
	// the name of the unused key as "fizz". However, if there
	// is a "fizz" key in the config structure, but no
	// "buzz" key in that "fizz" structure, then mapstructure
	// will report the name of the unused key as "fizz.buzz".
	toIgnore := make(map[string]struct{})
	for _, k := range m.strictModeIgnoredKeys {
		k = strings.ToLower(k)
		toIgnore[k] = struct{}{}
	}
	unusedNotIgnored := make([]string, 0)
	for _, unusedKey := range metadata.Unused {
		_, ok := toIgnore[unusedKey]
		if ok {
			continue
		}
		unusedNotIgnored = append(unusedNotIgnored, unusedKey)
	}
	if len(unusedNotIgnored) > 0 {
		msg := fmt.Sprintf("Misconfiguration error: found unexpected config key(s): %s", strings.Join(unusedNotIgnored, ","))
		return fmt.Errorf(msg)
	}
	return nil
}

func makeDefaultDecodeHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		// Function to accommodate for log level.
		func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
			if f.Kind() != reflect.String {
				return data, nil
			}

			deprecated := func(str string, level log.Level) (log.Level, error) {
				rawlog.Printf("deprecated config log level: %s, use %s instead", str, level.String())
				return level, nil
			}

			switch str := strings.ToLower(data.(string)); str {
			case "panic":
				return deprecated(str, log.ErrorLevel)
			case "fatal":
				return deprecated(str, log.ErrorLevel)
			case "error":
				return log.ErrorLevel, nil
			case "warn":
				return deprecated(str, log.InfoLevel)
			case "info":
				return log.InfoLevel, nil
			case "debug":
				return log.DebugLevel, nil
			case "trace":
				return deprecated(str, log.DebugLevel)
			default:
				return data, nil
			}
		},
		// Function to support jsontime.Duration
		jsontime.DurationMapstructureDecodeHookFunc,
		// Appended by the two default functions
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		// Function to support config.SensitiveString
		StringToSensitiveStringHookFunc(),
	)
}
