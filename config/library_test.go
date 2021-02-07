package config

import (
	"testing"

	"github.com/anz-bank/sysl-go/log"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/validator"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vv9 "gopkg.in/go-playground/validator.v9"
)

func ErrorDueToFields(t assert.TestingT, err error, field ...string) {
	found := map[string]bool{}
	for _, e := range err.(vv9.ValidationErrors) {
		found[e.StructField()] = true
	}

	if len(found) != len(field) {
		for _, f := range field {
			if _, ok := found[f]; !ok {
				assert.Failf(t, "Field '%s' expected to fail but didn't", f)
			}
		}
	}
}

func NewString(from string) *string {
	return &from
}

func NewSecret(s string) *common.SensitiveString {
	ss := common.NewSensitiveString(s)
	return &ss
}

func defaultConfig() *LibraryConfig {
	return &LibraryConfig{
		Log: defaultLogConfig(),
	}
}

func defaultLogConfig() LogConfig {
	logLevel := log.InfoLevel
	logFormat := "color"
	logCaller := false
	return LogConfig{
		Level:        logLevel,
		Format:       logFormat,
		ReportCaller: logCaller,
	}
}

func TestValidateDefaultConfig(t *testing.T) {
	config := defaultConfig()
	err := config.Validate()
	require.NoError(t, err)
}

func TestValidateDefaultLogConfig(t *testing.T) {
	cfg := defaultLogConfig()
	err := validator.Validate(cfg)
	require.NoError(t, err)
}

func TestValidateLogInvalidFormatter(t *testing.T) {
	config := defaultConfig()
	config.Log.Format = "BAD-FORMATTER"
	err := config.Validate()
	require.Error(t, err)
}

func TestValidateLogDebugLevel(t *testing.T) {
	config := defaultConfig()
	config.Log.Level = log.DebugLevel
	err := config.Validate()
	require.NoError(t, err)
}
