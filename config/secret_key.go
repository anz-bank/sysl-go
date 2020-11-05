package config

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/anz-bank/sysl-go/common"
)

const (
	SecretKeyEncodingBase64 = "base64"
)

type SecretKeyConfig struct {
	Encoding         *string                 `yaml:"encoding" mapstructure:"encoding" json:"encoding"`
	Alias            *string                 `yaml:"alias,omitempty" mapstructure:"alias,omitempty" json:"alias,omitempty"`
	KeyStore         *string                 `yaml:"keyStore,omitempty" mapstructure:"keyStore,omitempty" json:"keyStore,omitempty"`
	KeyStorePassword *common.SensitiveString `yaml:"keyStorePassword,omitempty" mapstructure:"keyStorePassword,omitempty" json:"keyStorePassword,omitempty"`
	Value            *common.SensitiveString `yaml:"value,omitempty" mapstructure:"value,omitempty" json:"value,omitempty"`
}

var SecretKeyValidators = map[string]func(cfg *SecretKeyConfig) error{
	SecretKeyEncodingBase64: validateBase64Value,
}

func (s *SecretKeyConfig) Validate() error {
	if s == nil {
		return nil
	}

	if s.Encoding == nil {
		return fmt.Errorf("encoding config missing")
	}

	if validateFn, ok := SecretKeyValidators[strings.ToLower(*s.Encoding)]; ok {
		return validateFn(s)
	}

	validEncoding := make([]string, 0, len(SecretKeyValidators))
	for encoding := range SecretKeyValidators {
		validEncoding = append(validEncoding, encoding)
	}
	sort.Strings(validEncoding)

	return fmt.Errorf("encoding `%s` is invalid, must be one of %+q", *s.Encoding, validEncoding)
}

func validateBase64Value(cfg *SecretKeyConfig) error {
	if cfg.Value == nil {
		return fmt.Errorf("value config missing")
	}
	if _, err := base64.StdEncoding.DecodeString(cfg.Value.Value()); err != nil {
		return fmt.Errorf("value config is invalid")
	}
	return nil
}

type SecretKey struct {
	common.SensitiveString
}

var SecretKeyReader = map[string]func(cfg *SecretKeyConfig) ([]byte, error){
	SecretKeyEncodingBase64: readBase64Value,
}

func MakeSecretKey(cfg *SecretKeyConfig) (*SecretKey, error) {
	if cfg == nil {
		return nil, fmt.Errorf("secret key load error, config is nil")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	readFn, ok := SecretKeyReader[strings.ToLower(*cfg.Encoding)]
	if !ok {
		validEncoding := make([]string, 0, len(SecretKeyReader))
		for encoding := range SecretKeyReader {
			validEncoding = append(validEncoding, encoding)
		}
		sort.Strings(validEncoding)
		return nil, fmt.Errorf("secret key load error, encoding `%s` is invalid, must be one of %+q", *cfg.Encoding, validEncoding)
	}

	key, err := readFn(cfg)
	if err != nil {
		return nil, err
	}

	return &SecretKey{common.NewSensitiveString(string(key))}, nil
}

func readBase64Value(cfg *SecretKeyConfig) ([]byte, error) {
	return base64.StdEncoding.DecodeString(cfg.Value.Value())
}
