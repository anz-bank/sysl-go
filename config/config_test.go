package config

import (
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/log"

	"gopkg.in/go-playground/validator.v9"

	"github.com/stretchr/testify/require"
)

type TestMyConfig struct {
	Server TestServer `yaml:"server"`
}

type TestServer struct {
	AdminServer TestAdminServer `yaml:"adminServer"`
}

type TestAdminServer struct {
	ContextTimeout time.Duration `yaml:"contextTimeout"`
	HTTP           TestHTTP      `yaml:"http"`
}

type TestHTTP struct {
	BasePath     string `yaml:"basePath"`
	ReadTimeout  string `yaml:"readTimeout"`
	WriteTimeout string `yaml:"writeTimeout"`
}

type TestDownstreamConfig struct {
	ContextTimeout time.Duration        `yaml:"contextTimeout"`
	Foo            CommonDownstreamData `yaml:"foo"`
	Bar            CommonDownstreamData `yaml:"bar"`
}

func testDefaultConfig() DefaultConfig {
	return DefaultConfig{
		Library: LibraryConfig{},
		GenCode: GenCodeConfig{
			Downstream: &TestDownstreamConfig{},
		},
	}
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	defaultConfig := testDefaultConfig()
	myConfig := TestMyConfig{}
	err := LoadConfig("testdata/config.yaml", &defaultConfig, &myConfig)

	require.Nil(t, err)

	require.Equal(t, 2*time.Second, myConfig.Server.AdminServer.ContextTimeout)
	require.Equal(t, "/admintest", myConfig.Server.AdminServer.HTTP.BasePath)

	require.True(t, defaultConfig.Library.Log.ReportCaller)
	require.Equal(t, log.InfoLevel, defaultConfig.Library.Log.Level)

	require.Equal(t, 8080, defaultConfig.GenCode.Upstream.HTTP.Common.Port)
	require.Equal(t, 8081, defaultConfig.GenCode.Upstream.GRPC.Port)
	require.Equal(t, 10*time.Second, defaultConfig.GenCode.Downstream.(*TestDownstreamConfig).Foo.ClientTimeout)
	require.Equal(t, "https://bar.example.com", defaultConfig.GenCode.Downstream.(*TestDownstreamConfig).Bar.ServiceURL)
}

func TestLoadInvalidConfig(t *testing.T) {
	t.Parallel()

	defaultConfig := testDefaultConfig()
	myConfig := TestMyConfig{}
	err := LoadConfig("testdata/config_invalid.yaml", &defaultConfig, &myConfig)

	_, ok := err.(validator.ValidationErrors)
	require.True(t, ok)
}
