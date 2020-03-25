package config

import (
	"fmt"
	"testing"

	"github.com/anz-bank/sysl-go/common"
	"github.com/stretchr/testify/assert"
)

func TestSecretKeyConfigNil(t *testing.T) {
	var cfg *SecretKeyConfig

	err := cfg.Validate()

	assert.NoError(t, err)
}

//nolint:funlen // Needs to be long due to struct
func TestSecretKeyConfigValidation(t *testing.T) {
	testData := []struct {
		name string
		in   *SecretKeyConfig
		out  error
	}{
		{
			"secretKey.encoding is nil",
			&SecretKeyConfig{},
			fmt.Errorf("encoding config missing"),
		},
		{
			"secretKey.encoding is invalid",
			&SecretKeyConfig{
				Encoding: NewString("abcdefg"),
			},
			fmt.Errorf("encoding `abcdefg` is invalid, must be one of [\"base64\"]"),
		},
		{
			"secretKey.value config missing",
			&SecretKeyConfig{
				Encoding: NewString("base64"),
			},
			fmt.Errorf("value config missing"),
		},
		{
			"secretKey.value is invalid",
			&SecretKeyConfig{
				Encoding: NewString("base64"),
				Value:    NewSecret("abcdefg"),
			},
			fmt.Errorf("value config is invalid"),
		},
	}

	for _, tt := range testData {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.in.Validate()
			assert.Error(t, err, tt.name)
			assert.Equal(t, tt.out, err, tt.name)
		})
	}
}

func TestMakeSecretKeyBase64Success(t *testing.T) {
	cfg := &SecretKeyConfig{
		Encoding: NewString("base64"),
		Value:    NewSecret("dGVzdHBhc3N3b3JkMTIzNA=="),
	}

	// When
	secretKey, err := MakeSecretKey(cfg)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, secretKey)
	assert.Equal(t, "testpassword1234", secretKey.Value())
	assert.Equal(t, common.DefaultReplacementText, secretKey.String())
}

func TestMakeSecretKeyNilConfig(t *testing.T) {
	// When
	secretKey, err := MakeSecretKey(nil)

	// Then
	assert.Nil(t, secretKey)
	assert.EqualError(t, err, "secret key load error, config is nil")
}

func TestMakeSecretKeyInvalidConfig(t *testing.T) {
	// Given
	cfg := &SecretKeyConfig{}

	// When
	secretKey, err := MakeSecretKey(cfg)

	// Then
	assert.Nil(t, secretKey)
	assert.EqualError(t, err, "encoding config missing")
}
