package config

import (
	"fmt"
	"testing"

	"github.com/anz-bank/sysl-go/validator"

	"encoding/json"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const (
	testYAML = `secret: "test string"`
	testJSON = `{ "secret":"test string" }`

	secretText   = "test string"
	redactedText = DefaultReplacementText
)

type testStruct struct {
	Secret SensitiveString `json:"secret" yaml:"secret" validate:"alpha"`
}

type testnonsecretStruct struct {
	Secret string `json:"secret" yaml:"secret"`
}

func TestPMarshaltypes(t *testing.T) {
	v := NewSensitiveString(secretText)

	box := interface{}(&v)
	_, ok := box.(json.Marshaler)
	require.True(t, ok)
	_, ok = box.(json.Unmarshaler)
	require.True(t, ok)
	_, ok = box.(yaml.Marshaler)
	require.True(t, ok)
	_, ok = box.(yaml.Unmarshaler)
	require.True(t, ok)
}

func TestPrintfHidesSecret(t *testing.T) {
	v := NewSensitiveString(secretText)

	require.Equal(t, redactedText, fmt.Sprintf("%v", v))
}

func TestPrintfAllowsSecret(t *testing.T) {
	v := NewSensitiveString(secretText)

	require.Equal(t, secretText, fmt.Sprintf("%v", v.Value()))
}

func TestYAMLRoundTripHidesSecret(t *testing.T) {
	var secretObj testStruct
	var clearObj testnonsecretStruct

	err := yaml.UnmarshalStrict([]byte(testYAML), &secretObj)
	require.NoError(t, err)

	require.Equal(t, redactedText, secretObj.Secret.String())
	require.Equal(t, secretText, secretObj.Secret.Value())

	redactedBytes, err := yaml.Marshal(&secretObj)
	require.NoError(t, err)
	err = yaml.UnmarshalStrict(redactedBytes, &clearObj)
	require.NoError(t, err)

	require.Equal(t, redactedText, clearObj.Secret)
}

func TestJSONRoundTripHidesSecret(t *testing.T) {
	var secretObj testStruct
	var clearObj testnonsecretStruct

	err := json.Unmarshal([]byte(testJSON), &secretObj)
	require.NoError(t, err)

	require.Equal(t, redactedText, secretObj.Secret.String())
	require.Equal(t, secretText, secretObj.Secret.Value())

	redactedBytes, err := json.Marshal(&secretObj)
	require.NoError(t, err)
	err = json.Unmarshal(redactedBytes, &clearObj)
	require.NoError(t, err)

	require.Equal(t, redactedText, clearObj.Secret)
}

func TestSensitiveString_UnmarshalErrors(t *testing.T) {
	invalidTestYAML := `secret: [test string]`
	invalidTestJSON := `{ "secret":{"test string":"f"}} `

	var secretObj testStruct
	err := json.Unmarshal([]byte(invalidTestJSON), &secretObj)
	require.Error(t, err)

	err = yaml.UnmarshalStrict([]byte(invalidTestYAML), &secretObj)
	require.Error(t, err)
}

func TestSensitiveStringValidation(t *testing.T) {
	err := validator.Validate(testStruct{NewSensitiveString("abcds")})
	require.NoError(t, err)

	err = validator.Validate(testStruct{NewSensitiveString("123")})
	require.Error(t, err)
}
