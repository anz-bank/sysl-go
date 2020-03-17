package config

import (
	"testing"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/validator"

	"github.com/sirupsen/logrus"
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
	logLevel := logrus.InfoLevel
	logFormat := "color"
	logCaller := false
	return LogConfig{
		Level:        logLevel,
		Format:       logFormat,
		ReportCaller: logCaller,
	}
}

// Config
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
	config.Log.Level = logrus.DebugLevel
	err := config.Validate()
	require.NoError(t, err)
}

func TestValidateNilSplunk(t *testing.T) {
	config := defaultLogConfig()
	config.Splunk = nil
	err := validator.Validate(config)
	require.NoError(t, err)
}
func TestValidateInvalidSplunk(t *testing.T) {
	config := defaultLogConfig()
	config.Splunk = &SplunkConfig{}
	err := validator.Validate(config)
	require.Error(t, err)
}

var splunkEmptyValueTests = []struct {
	in      SplunkConfig
	missing string // which field is the error expected for?
	name    string
}{
	{SplunkConfig{common.NewSensitiveString(""), "index", "target", "source", "sourcetype"}, "TokenBase64", "TEST: splunkNilValueTests #1"},
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "", "target", "source", "sourcetype"}, "Index", "TEST: splunkNilValueTests #2"},
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "index", "", "source", "sourcetype"}, "Target", "TEST: splunkNilValueTests #3"},
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "index", "target", "", "sourcetype"}, "Source", "TEST: splunkNilValueTests #4"},
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "index", "target", "source", ""}, "SourceType", "TEST: splunkNilValueTests #5"},
}

func TestSplunkNilValues(t *testing.T) {
	for _, tt := range splunkEmptyValueTests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.in)
			require.Error(t, err)

			ErrorDueToFields(t, err, tt.missing)
		})
	}
}

var splunkTargetTests = []struct {
	in          SplunkConfig
	expectError bool
	name        string
}{
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "index", "http://test", "source", "sourcetype"}, false, "TEST: splunkGoodTargetTests #1"},
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "index", "https://test/test", "source", "sourcetype"}, false, "TEST: splunkGoodTargetTests #2"},
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "index", "http://test/", "source", "sourcetype"}, false, "TEST: splunkGoodTargetTests #3"},
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "index", "http://test:8088/", "source", ("sourcetype")}, false, "TEST: splunkGoodTargetTests #4"},
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "index", "http://test:8088/test", "source", "sourcetype"}, false, "TEST: splunkGoodTargetTests #5"},
	{SplunkConfig{common.NewSensitiveString("dG9rZW4="), "index", "test", "source", "sourcetype"}, true, "TEST: splunkGoodTargetTests #6"},
}

func TestSplunkGoodTargets(t *testing.T) {
	for _, tt := range splunkTargetTests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.in)
			if tt.expectError {
				require.Error(t, err)
				ErrorDueToFields(t, err, "Target")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBase64Token(t *testing.T) {
	config := SplunkConfig{common.NewSensitiveString("token"), "index", "http://test:8088/test", "source", "sourcetype"}
	err := validator.Validate(config)
	require.Errorf(t, err, "TEST: NotBase64TokenTest")

	config.TokenBase64 = common.NewSensitiveString("dG9rZW4=")
	err = validator.Validate(config)
	require.NoError(t, err, "TEST: Base64TokenTest")
	require.Equal(t, "token", config.Token(), "TEST: tokenNotNilTest")
}
