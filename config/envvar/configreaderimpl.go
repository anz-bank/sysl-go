package envvar

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/anz-bank/sysl-go/jsontime"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// configReaderImpl exposes a wrapper api for viper.
type configReaderImpl struct {
	envVars *viper.Viper
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
	if err := m.envVars.Unmarshal(config, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			// Function to accommodate for log level.
			func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if f.Kind() != reflect.String {
					return data, nil
				}
				switch strings.ToLower(data.(string)) {
				case "panic":
					return logrus.PanicLevel, nil
				case "fatal":
					return logrus.FatalLevel, nil
				case "error":
					return logrus.ErrorLevel, nil
				case "warn":
					return logrus.WarnLevel, nil
				case "info":
					return logrus.InfoLevel, nil
				case "debug":
					return logrus.DebugLevel, nil
				case "trace":
					return logrus.TraceLevel, nil
				default:
					return data, nil
				}
			},
			// Function to support jsontime.Duration
			jsontime.DurationMapstructureDecodeHookFunc,
			// Appended by the two default functions
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	)); err != nil {
		return fmt.Errorf("Unable to decode into struct %s", err)
	}
	return nil
}
