package envvar

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type genCodeConfig struct {
	Downstream downstreamConfig `mapstructure:"downstream"`
}

type libraryConfig struct {
	Log       logConfig `mapstructure:"log"`
	Profiling bool      `mapstructure:"profiling"`
}

type logConfig struct {
	Format string       `mapstructure:"format"`
	Level  logrus.Level `mapstructure:"level"`
	Level1 logrus.Level `mapstructure:"level1"`
	Level2 logrus.Level `mapstructure:"level2"`
	Level3 logrus.Level `mapstructure:"level3"`
	Level4 logrus.Level `mapstructure:"level4"`
	Level5 logrus.Level `mapstructure:"level5"`
	Level6 logrus.Level `mapstructure:"level6"`
	Caller bool         `mapstructure:"caller"`
}

type commonDownstreamData struct {
	ServiceURL    string        `mapstructure:"serviceURL"`
	ClientTimeout time.Duration `mapstructure:"clientTimeout"`
	ReadTimeout   float64       `mapstructure:"readTimeout"`
	CreationTime  string        `mapstructure:"creationTime"`
}

type downstreamConfig struct {
	ContextTimeout time.Duration        `mapstructure:"contextTimeout"`
	Foo            commonDownstreamData `mapstructure:"foo"`
	Bar            commonDownstreamData `mapstructure:"bar"`
}

type config struct {
	Library libraryConfig `mapstructure:"library"`
	Gencode genCodeConfig `mapstructure:"genCode"`
}

func TestGetStringFromFile(t *testing.T) {
	t.Parallel()

	b := NewConfigReaderBuilder()
	reader := b.AttachEnvPrefix("simpleApp").WithConfigFile("../testdata/config.yaml").Build()
	fooURL, err := reader.GetString("genCode.downstream.foo.serviceURL")
	require.Nil(t, err)
	assert.Equal(t, "https://foo.example.com", fooURL)
}

func TestGetStringErr(t *testing.T) {
	t.Parallel()

	b := NewConfigReaderBuilder()
	reader := b.AttachEnvPrefix("simpleApp").WithConfigFile("../testdata/config.yaml").Build()
	s, err := reader.GetString("genCode.downstream.foo")
	require.NotNil(t, err)
	assert.Equal(t, "", s)
}

func TestGetStringFromEnv(t *testing.T) {
	t.Parallel()

	b := NewConfigReaderBuilder()
	reader := b.AttachEnvPrefix("simple").WithConfigFile("../testdata/config.yaml").Build()
	os.Setenv("SIMPLE_GENCODE_DOWNSTREAM_FOO_SERVICEURL", "https://env.foo.example.com")
	fooURL, err := reader.GetString("genCode.downstream.foo.serviceURL")
	require.Nil(t, err)
	assert.Equal(t, "https://env.foo.example.com", fooURL)
}

func TestGetStringFrom2ndSource(t *testing.T) {
	t.Parallel()

	b := NewConfigReaderBuilder()
	reader := b.AttachEnvPrefix("simple").WithConfigFile("../testdata/config.yaml").Build()
	os.Setenv("SIMPLE_GENCODE_DOWNSTREAM_BAR_SERVICEURL", "")
	barURL, err := reader.GetString("genCode.downstream.bar.serviceURL")
	require.Nil(t, err)
	assert.Equal(t, "https://bar.example.com", barURL)
}

func TestGetMultipleConfigFiles(t *testing.T) {
	t.Parallel()

	b := NewConfigReaderBuilder()
	reader := b.AttachEnvPrefix("simple").WithConfigFile("../testdata/config.yaml").WithConfigName(
		"config_log", "./", "../testdata").Build()
	calleeLog, err := reader.GetString("library.log.callee")
	require.Nil(t, err)
	assert.Equal(t, "true", calleeLog)
}

func TestUnmarshalFromFileWithPrefix(t *testing.T) {
	t.Parallel()

	conf := config{}
	b := NewConfigReaderBuilder().WithFs(afero.NewOsFs()).WithConfigFile("../testdata/config.yaml")
	fooURL, err := b.Build().GetString("genCode.downstream.foo.serviceURL")
	require.Nil(t, err)
	assert.Equal(t, "https://foo.example.com", fooURL)
	os.Setenv("ENV_GENCODE_DOWNSTREAM_FOO_SERVICEURL", "https://env.foo.example.com")
	os.Setenv("ENV_GENCODE_DOWNSTREAM_BAR_SERVICEURL", "https://env.bar.example.com")
	b.AttachEnvPrefix("env")
	err = b.Build().Unmarshal(&conf)
	require.Nil(t, err)
	assert.Equal(t, "https://env.foo.example.com", conf.Gencode.Downstream.Foo.ServiceURL)
	assert.Equal(t, "https://env.bar.example.com", conf.Gencode.Downstream.Bar.ServiceURL)
}

func TestUnmarshalFromFile(t *testing.T) {
	t.Parallel()

	conf := config{}
	b := NewConfigReaderBuilder()
	reader := b.WithFs(afero.NewOsFs()).WithConfigFile("../testdata/config.yaml").Build()
	err := reader.Unmarshal(&conf)
	require.Nil(t, err)
	assert.Equal(t, "https://foo.example.com", conf.Gencode.Downstream.Foo.ServiceURL)
	assert.Equal(t, "https://bar.example.com", conf.Gencode.Downstream.Bar.ServiceURL)
}

func TestUnmarshalFromFileWithStrictMode(t *testing.T) {
	t.Parallel()

	type DemoConfig struct {
		Barr int `mapstructure:"barr"`
	}

	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "a.yaml", []byte("foo: 123\nbarr: 456"), 0644)
	require.NoError(t, err)

	type scenario struct {
		name           string
		b              ConfigReaderBuilder
		expectedConfig DemoConfig
		expectedErr    error
	}

	scenarios := []scenario{
		{
			name:           "default",
			b:              NewConfigReaderBuilder(),
			expectedConfig: DemoConfig{Barr: 456},
			expectedErr:    nil,
		},
		{
			name:           "strict-mode-disabled",
			b:              NewConfigReaderBuilder().WithStrictMode(false),
			expectedConfig: DemoConfig{Barr: 456},
			expectedErr:    nil,
		},
		{
			name:           "strict-mode-enabled",
			b:              NewConfigReaderBuilder().WithStrictMode(true),
			expectedConfig: DemoConfig{Barr: 456},
			expectedErr:    fmt.Errorf("Misconfiguration error: found unexpected config key(s): foo"),
		},
		{
			name:           "strict-mode-enabled-with-exception-ignored",
			b:              NewConfigReaderBuilder().WithStrictMode(true, "foo"),
			expectedConfig: DemoConfig{Barr: 456},
			expectedErr:    nil,
		},
		{
			name:           "strict-mode-enabled-with-case-insensitive-exception-ignored",
			b:              NewConfigReaderBuilder().WithStrictMode(true, "fOo"),
			expectedConfig: DemoConfig{Barr: 456},
			expectedErr:    nil,
		},
		{
			name:           "strict-mode-enabled-with-some-other-exception-ignored",
			b:              NewConfigReaderBuilder().WithStrictMode(true, "fib"),
			expectedConfig: DemoConfig{Barr: 456},
			expectedErr:    fmt.Errorf("Misconfiguration error: found unexpected config key(s): foo"),
		},
	}

	for _, s := range scenarios {
		s := s // force capture.
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()

			conf := DemoConfig{}
			reader := s.b.WithFs(fs).WithConfigFile("a.yaml").Build()
			err := reader.Unmarshal(&conf)

			require.Equal(t, s.expectedConfig, conf)
			require.Equal(t, s.expectedErr, err)
		})
	}
}
