package logging

import (
	"os"
	"testing"

	"github.com/anz-bank/sysl-go/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func defaultLogConfig() config.LogConfig {
	logLevel := logrus.InfoLevel
	logFormat := "color"
	logCaller := false
	return config.LogConfig{
		Level:        logLevel,
		Format:       logFormat,
		ReportCaller: logCaller,
	}
}

func TestLogger(t *testing.T) {
	req := require.New(t)

	cfg := defaultLogConfig()
	actual, err := Logger(os.Stderr, &cfg)
	textFormatter := &logrus.TextFormatter{DisableColors: false}
	expected := &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    textFormatter,
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ReportCaller: false,
	}
	req.NoError(err)
	req.Equal(expected, actual)

	cfg.ReportCaller = true
	actual, err = Logger(os.Stderr, &cfg)
	expected.ReportCaller = true
	req.NoError(err)
	req.Equal(expected, actual)

	cfg.Format = "BAD-FORMAT"
	_, err = Logger(os.Stderr, &cfg)
	req.Error(err)
}

var validLoggerFormatterTests = []struct {
	in   string
	out  interface{}
	name string
}{
	{"color", &logrus.TextFormatter{}, "TEST: validLoggerFormatterTests #1"},
	{"text", &logrus.TextFormatter{DisableColors: true}, "TEST: validLoggerFormatterTests #2"},
	{"json", &logrus.JSONFormatter{}, "TEST: validLoggerFormatterTests #3"},
}

func TestLogrusFormatterCreation(t *testing.T) {
	for _, tt := range validLoggerFormatterTests {
		actual, err := LogFormatter(tt.in)
		require.NoError(t, err, tt.name)
		require.Equal(t, tt.out, actual, tt.name)
	}
}

func TestLogFormatter(t *testing.T) {
	actual, err := LogFormatter("BadFormatter")
	require.Error(t, err)
	require.Nil(t, actual)
}
