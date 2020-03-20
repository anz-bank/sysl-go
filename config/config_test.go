package config

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
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

func TestSReadConfig(t *testing.T) {
	t.Parallel()

	defaultConfig := DefaultConfig{
		Library: LibraryConfig{},
		GenCode: GenCodeConfig{
			Downstream: &TestDownstreamConfig{},
		},
	}
	myConfig := TestMyConfig{}
	err := ReadConfig("testdata/config.yaml", &defaultConfig, &myConfig)

	require.Nil(t, err)

	require.Equal(t, 2*time.Second, myConfig.Server.AdminServer.ContextTimeout)
	require.Equal(t, "/admintest", myConfig.Server.AdminServer.HTTP.BasePath)

	require.True(t, defaultConfig.Library.Log.ReportCaller)
	require.Equal(t, logrus.WarnLevel, defaultConfig.Library.Log.Level)

	require.Equal(t, 8080, defaultConfig.GenCode.Upstream.HTTP.Common.Port)
	require.Equal(t, 8081, defaultConfig.GenCode.Upstream.GRPC.Port)
	require.Equal(t, 120*time.Second, defaultConfig.GenCode.Downstream.(*TestDownstreamConfig).Foo.ClientTimeout)
	require.Equal(t, "https://bar.example.com", defaultConfig.GenCode.Downstream.(*TestDownstreamConfig).Bar.ServiceURL)
}
