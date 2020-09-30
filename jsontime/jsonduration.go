package jsontime

import (
	"reflect"
	"time"
)

// Duration is an alias of time.Duration
//
// Can be marshalled to and from json and yaml.
type Duration time.Duration

func (d Duration) String() string {
	return time.Duration(d).String()
}

// Duration returns the equivalent time.Duration object.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// UnmarshalJSON implements json.Unmarshaller.
func (d *Duration) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*d = 0
		return nil
	}
	dur, err := time.ParseDuration(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// MarshalJSON implements json.Marshaller.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalYAML implements yaml.Unmarshaller.
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if s == "" {
		*d = 0
		return nil
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// MarshalYAML implements yaml.Marshaller.
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// DurationMapstructureDecodeHookFunc is a DecodeHookFunc for mapstructure.
//
// Use this if your config manager uses mapstructure (eg: viper, koanf).
func DurationMapstructureDecodeHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(Duration(0)) && t != reflect.TypeOf(time.Duration(0)) {
		return data, nil
	}

	// Convert it by parsing
	return time.ParseDuration(data.(string))
}
