package config

import (
	"encoding/json"
	"reflect"

	"github.com/anz-bank/sysl-go/validator"
	"github.com/mitchellh/mapstructure"
)

const DefaultReplacementText = "****************"

type SensitiveString struct {
	s           string
	replacement *string
}

func NewSensitiveString(from string) SensitiveString {
	r := DefaultReplacementText
	return SensitiveString{from, &r}
}

func (s SensitiveString) String() string {
	if s.replacement == nil {
		r := DefaultReplacementText
		s.replacement = &r
	}
	return *s.replacement
}
func (s *SensitiveString) Value() string {
	return s.s
}

func (s *SensitiveString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val string
	if err := unmarshal(&val); err != nil {
		return err
	}
	s.s = val
	return nil
}

// Note, this one needs to be an object receiver NOT a pointer receiver.
func (s SensitiveString) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}

func (s *SensitiveString) UnmarshalJSON(data []byte) error {
	var val string
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	s.s = val
	return nil
}

func (s *SensitiveString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func sensitiveStringValidator(field reflect.Value) interface{} {
	switch field.Interface().(type) {
	case SensitiveString:
		val := field.Interface().(SensitiveString)
		return val.Value()
	case *SensitiveString:
		val := field.Interface().(*SensitiveString)
		return val.Value()
	}

	return nil
}

// StringToSensitiveStringHookFunc returns a DecodeHookFunc that converts
// strings to SensitiveString.
func StringToSensitiveStringHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() == reflect.String && t == reflect.TypeOf(SensitiveString{}) {
			return NewSensitiveString(data.(string)), nil
		}
		return data, nil
	}
}

//nolint:gochecknoinits // We must use init here to setup a custom validator
func init() {
	validator.RegisterCustomValidator(sensitiveStringValidator, SensitiveString{})
}
